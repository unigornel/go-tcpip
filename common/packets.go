package common

import (
	"bytes"
	"io"
)

type PacketWriter interface {
	Write(io.Writer) error
}

func PacketToBytes(w PacketWriter) []byte {
	b := bytes.NewBuffer(nil)
	if err := w.Write(b); err != nil {
		panic(err)
	}
	return b.Bytes()
}

func PacketChecksum(w PacketWriter) uint16 {
	return Checksum(PacketToBytes(w))
}
