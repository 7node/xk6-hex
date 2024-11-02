package helper

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

var parser struct {
	Decoder Decoder
	Encoder Encoder
}

var Parser = &parser

type Decoder struct{}

func (d Decoder) Decode(payload, encKey, signKey []byte) ([]byte, error) {
	plainText, err := d.Decrypt(payload, encKey)
	if err != nil {
		return nil, err
	}
	signed, err := d.jsonUnmarshalPlainText(plainText)
	if err != nil {
		return nil, err
	}
	if valid, err := d.VerifySignature(signed, signKey); valid {
		return []byte(signed.Data), nil
	} else if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("invalid signature")
}

func (d Decoder) Decrypt(payload, key []byte) ([]byte, error) {
	iv := payload[:aes.BlockSize]
	cipherText := payload[aes.BlockSize:]

	blocker, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cbc := cipher.NewCBCDecrypter(blocker, iv)

	// 理论上不管是 PKCS5Padding 还是 PKCS7Padding 使用密文长度能够正确处理
	var plainText = make([]byte, len(cipherText))
	cbc.CryptBlocks(plainText, cipherText)
	plainText, err = PKCS7UnPadding(plainText)
	if err != nil {
		return nil, err
	}
	return plainText, nil
}

func (d Decoder) jsonUnmarshalPlainText(plainText []byte) (SignedPayload, error) {
	trimmed := bytes.TrimRight(plainText, "\x00")
	payload := SignedPayload{}
	err := json.Unmarshal(trimmed, &payload)
	if err != nil {
		return payload, err
	}
	// note (joe@29/09/24): 我也不知道, 为什么 base64 encode 之后还要再额外添加一个 '0' 再末尾
	if strings.HasSuffix(payload.Data, "=0") {
		payload.Data = strings.TrimRight(payload.Data, "0")
	} else if strings.HasSuffix(payload.Data, "=equest") {
		payload.Data = strings.TrimSuffix(payload.Data, "equest")
	}
	return payload, nil
}

func (d Decoder) VerifySignature(payload SignedPayload, signKey []byte) (bool, error) {
	signature, err := hex.DecodeString(payload.Signature)
	if err != nil {
		return false, err
	}
	hash := hmac.New(sha256.New, signKey)
	hash.Write([]byte(payload.Data))
	actual := hash.Sum(nil)
	return hmac.Equal(actual, signature), nil
}

type SignedPayload struct {
	Data      string `json:"data"`
	Signature string `json:"signature"`
}

type Encoder struct{}

func (e Encoder) Encode(inputsText, encKey, signKey []byte) ([]byte, error) {
	hash := hmac.New(sha256.New, signKey)
	hash.Write(inputsText)
	signature := hash.Sum(nil)

	_payload := SignedPayload{Data: string(inputsText), Signature: hex.EncodeToString(signature)}
	plainText, err := json.Marshal(_payload)
	if err != nil {
		return nil, err
	}
	plainText = append(plainText, '\x00')
	paddedPlainText := PKCS7Padding(plainText, aes.BlockSize)

	blocker, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, err
	}

	cipherText := make([]byte, aes.BlockSize+len(paddedPlainText))
	// read IV
	n, err := rand.Read(cipherText[:aes.BlockSize])
	if err != nil {
		return nil, err
	}
	if n != aes.BlockSize {
		return nil, fmt.Errorf("expected to read %v bytes, read %v", aes.BlockSize, n)
	}

	cbc := cipher.NewCBCEncrypter(blocker, cipherText[:aes.BlockSize])
	cbc.CryptBlocks(cipherText[aes.BlockSize:], paddedPlainText)
	return cipherText, nil
}

// PKCS7UnPadding PKCS5/7 去填充
func PKCS7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("invalid padding size")
	}
	unPadding := int(data[length-1])
	if unPadding > length {
		return nil, fmt.Errorf("invalid padding size")
	}
	return data[:(length - unPadding)], nil
}

func PKCS7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}
