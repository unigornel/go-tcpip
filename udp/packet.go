package udp

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/unigornel/go-tcpip/common"
)

var (
	// ErrInvalidLength is an error returned when the packet length is
	// too small.
	ErrInvalidLength = errors.New("Length field is too small")

	// ErrInvalidChecksum is an error returned when the packet checksum
	// is incorrect.
	ErrInvalidChecksum = errors.New("Checksum field is incorrect")
)

// Header is the UDP packet header.
type Header struct {
	SourcePort      uint16
	DestinationPort uint16
	Length          uint16
	Checksum        uint16
}

// NewHeader reads a header from a reader.
func NewHeader(r io.Reader) (Header, error) {
	var header Header
	err := binary.Read(r, binary.BigEndian, &header)
	return header, err
}

// Write the header to a Writer.
func (h Header) Write(w io.Writer) error {
	return binary.Write(w, binary.BigEndian, h)
}

// Packet is a UDP packet
type Packet struct {
	Header
	Payload []byte
}

// NewPacket reads a packet from a reader.
func NewPacket(r io.Reader) (p Packet, err error) {
	p.Header, err = NewHeader(r)
	if err != nil {
		return
	}

	if p.Length < 8 {
		err = ErrInvalidLength
		return
	}

	p.Payload = make([]byte, p.Length-8)
	if _, err = r.Read(p.Payload); err != nil {
		return
	}

	if p.Checksum != 0 && p.CalculateChecksum() != p.Checksum {
		err = ErrInvalidChecksum
		return
	}

	return
}

// Write the packet to a Writer.
func (p Packet) Write(w io.Writer) error {
	if err := p.Header.Write(w); err != nil {
		return err
	}
	_, err := w.Write(p.Payload)
	return err
}

// CalculateChecksum calculates the correct checksum of the packet.
func (p Packet) CalculateChecksum() uint16 {
	p.Checksum = 0
	return common.PacketChecksum(p)
}
