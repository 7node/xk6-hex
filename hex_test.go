package hex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHexDecode(t *testing.T) {
	hex := Hex{}

	decoded := hex.HexDecode("486578206465636f6465")

	assert.Equal(t, []byte{72, 101, 120, 32, 100, 101, 99, 111, 100, 101}, decoded)
}
