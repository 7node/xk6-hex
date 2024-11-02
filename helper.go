package helper

import (
	"go.k6.io/k6/js/modules"

	goHex "encoding/hex"
)

// init is called by the Go runtime at application startup.
func init() {
	modules.Register("k6/x/helper", New())
}

// Hex is the type for our custom API.
type Module struct{}

func New() *Module {
	return &Module{}
}

// Decode returns the decoded string.
func (c *Module) HexDecode(hex string) []byte {
	decoded, err := goHex.DecodeString(hex)
	if err != nil {
		return nil
	}
	return decoded
}

func (c *Module) HexEncode(data []byte) string {
	return goHex.EncodeToString(data)
}

func (c *Module) EncodeMessage(payload, encKey, signKey string) []byte {
	encKeyBytes := c.HexDecode(encKey)
	signKeyBytes := c.HexDecode(signKey)

	encoded, err := Parser.Encoder.Encode([]byte(payload), encKeyBytes, signKeyBytes)
	if err != nil {
		return nil
	}
	return encoded
}

func (c *Module) DecodeMessage(payload, encKey, signKey string) []byte {
	encKeyBytes := c.HexDecode(encKey)
	signKeyBytes := c.HexDecode(signKey)

	decoded, err := Parser.Decoder.Decode([]byte(payload), encKeyBytes, signKeyBytes)
	if err != nil {
		return nil
	}
	return decoded
}
