package hex

import (
	"go.k6.io/k6/js/modules"

	goHex "encoding/hex"
)

// init is called by the Go runtime at application startup.
func init() {
	modules.Register("k6/x/hex", new(Hex))
}

// Hex is the type for our custom API.
type Hex struct {
}

// Decode returns the decoded string.
func (c *Hex) HexDecode(hex string) []byte {
	decoded, err := goHex.DecodeString(hex)
	if err != nil {
		return nil
	}
	return decoded
}
