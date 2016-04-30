package tcpip

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChecksum(t *testing.T) {
	tests := []struct {
		Hex      string
		Checksum uint16
	}{
		{"450000349c8340003e060000c0a86401c0a86413", 0x56DB},
		{"450000349c8340003e060000c0a86401c0a8641300", 0x56DB},
		{"FFFF", 0xFFFF},
	}

	for _, test := range tests {
		b, err := hex.DecodeString(test.Hex)
		assert.Nil(t, err)
		assert.Equal(t, test.Checksum, Checksum(b))
	}
}
