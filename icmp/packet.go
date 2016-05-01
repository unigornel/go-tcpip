package icmp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"

	"github.ugent.be/unigornel/go-tcpip"
)

// Type is the type of the ICMP packet.
type Type uint8

const (
	// EchoReplyType is the ICMP type for an echo reply.
	EchoReplyType = 0
	// EchoRequestType is the ICMP type for an echo request.
	EchoRequestType = 8
)

// Code is the code of the ICMP packet.
type Code uint8

const (
	// EchoReplyCode is the ICMP code for an echo reply.
	EchoReplyCode = 0
	// EchoRequestCode is the ICMP code for an echo request.
	EchoRequestCode = 0
)

// Header is the common ICMP header.
type Header struct {
	Type     Type
	Code     Code
	Checksum uint16
}

var (
	ErrUnsupportedICMPPacket = errors.New("unsupported ICMP packet")
)

// Data is an interface to handle ICMP data.
type Data interface {
	Write(io.Writer) error
}

// Packet is an ICMP packet.
type Packet struct {
	Header Header
	Data   Data
}

// NewPacket will read a packet from a reader.
func NewPacket(r io.Reader) (packet Packet, err error) {
	if err = binary.Read(r, binary.BigEndian, &packet.Header); err != nil {
		return
	}

	if packet.Header.Type == EchoRequestType && packet.Header.Code == EchoRequestCode {
		packet.Data, err = NewEcho(r)
	} else if packet.Header.Type == EchoReplyType && packet.Header.Code == EchoReplyType {
		packet.Data, err = NewEcho(r)
	} else {
		err = ErrUnsupportedICMPPacket
	}

	return
}

// NewEchoRequest creates a new echo request packet.
func NewEchoRequest(ident, seq uint16, payload []byte) Packet {
	return Packet{
		Header: Header{Type: EchoRequestType, Code: EchoRequestCode},
		Data: Echo{
			Header: EchoHeader{
				Identifier:     ident,
				SequenceNumber: seq,
			},
			Payload: payload,
		},
	}.WithChecksum()
}

// NewEchoReply creates a new echo reply packet.
func NewEchoReply(ident, seq uint16, payload []byte) Packet {
	return Packet{
		Header: Header{Type: EchoReplyType, Code: EchoReplyCode},
		Data: Echo{
			Header: EchoHeader{
				Identifier:     ident,
				SequenceNumber: seq,
			},
			Payload: payload,
		},
	}.WithChecksum()
}

// Write will write a packet to a writer.
func (p Packet) Write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, p.Header); err != nil {
		return err
	}
	return p.Data.Write(w)
}

// WithChecksum creates a copy of the packet with the correct checksum.
func (p Packet) WithChecksum() Packet {
	c := p
	c.Header.Checksum = 0

	var b bytes.Buffer
	if err := c.Write(&b); err != nil {
		panic(err)
	}
	c.Header.Checksum = tcpip.Checksum(b.Bytes())
	return c
}

// EchoHeader is the header of echo request/reply packets.
type EchoHeader struct {
	Identifier     uint16
	SequenceNumber uint16
}

// Echo is the data for echo request/reply packets.
type Echo struct {
	Header  EchoHeader
	Payload []byte
}

// NewEcho reads echo request/reply data from a reader.
func NewEcho(r io.Reader) (data Echo, err error) {
	if err = binary.Read(r, binary.BigEndian, &data.Header); err != nil {
		return
	}
	data.Payload, err = ioutil.ReadAll(r)
	return
}

// Write the echo request/reply data to the writer.
func (d Echo) Write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, &d); err != nil {
		return err
	}
	_, err := w.Write(d.Payload)
	return err
}
