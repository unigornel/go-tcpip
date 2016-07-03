package common

import (
	"bytes"
	"io"
)

// PacketWriter writes a packet to the given writer.
type PacketWriter interface {
	Write(io.Writer) error
}

// PacketToBytes will convert a packet to a byte slice.
func PacketToBytes(w PacketWriter) []byte {
	b := bytes.NewBuffer(nil)
	if err := w.Write(b); err != nil {
		panic(err)
	}
	return b.Bytes()
}

// PacketChecksum will calculate the checksum of a packet.
//
// The function will first convert the packet to a byte slice. Then
// it will calculate the checksum from this byte slice.
func PacketChecksum(w PacketWriter) uint16 {
	return Checksum(PacketToBytes(w))
}
