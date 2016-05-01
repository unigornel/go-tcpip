package ipv4

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync/atomic"

	"github.ugent.be/unigornel/go-tcpip"
)

var (
	// ErrInvalidIHL is a check error returned when the length of the Options
	// byte slice does not match the IHL value.
	ErrInvalidIHL = errors.New("IHL field does not match Options byte slice length")

	// ErrInvalidChecksum is a check error returned when header checksum is
	// incorrect.
	ErrInvalidChecksum = errors.New("Checksum field is incorrect")

	// ErrInvalidTotalLength is an error returned when the packet length is
	// too small.
	ErrInvalidTotalLength = errors.New("TotalLength field is too small")
)

// Address is an IPv4 address.
type Address [4]byte

// NewAddress creates a new address from a string.
//
// If the string is invalid, NewAddress returns false.
func NewAddress(s string) (Address, bool) {
	var a Address
	if i := net.ParseIP(s); i != nil {
		if i := i.To4(); i != nil {
			a[0] = i[0]
			a[1] = i[1]
			a[2] = i[2]
			a[3] = i[3]
			return a, true
		}
	}
	return a, false
}

func (a Address) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", a[0], a[1], a[2], a[3])
}

// Bytes copies an address to a new byte slice.
func (a Address) Bytes() []byte {
	s := make([]byte, len(a))
	for i, b := range a {
		s[i] = b
	}
	return s
}

// Protocol is an IPv4 protocol.
type Protocol uint8

const (
	// ProtocolICMP is used for the ICMP protocol.
	ProtocolICMP = 1
)

// Header is the logical version of an IPv4 header.
type Header struct {
	Version        uint8
	IHL            uint8
	ToS            uint8
	TotalLength    uint16
	Identification uint16
	Flags          uint8
	FragmentOffset uint16
	TTL            uint8
	Protocol       Protocol
	Checksum       uint16
	Source         Address
	Destination    Address
	Options        []byte
}

// Write the header to a Writer.
func (h Header) Write(w io.Writer) error {
	if err := h.RawHeader().Write(w); err != nil {
		return err
	}
	_, err := w.Write(h.Options)
	return err
}

// NewHeader reads a header from a Reader.
func NewHeader(r io.Reader) (Header, error) {
	raw, err := NewRawHeader(r)
	if err != nil {
		return Header{}, err
	}

	h := raw.Header()
	numOptionBytes := (int(h.IHL) - 5) * 8
	if numOptionBytes < 0 {
		return h, ErrInvalidIHL
	} else if numOptionBytes > 0 {
		h.Options = make([]byte, numOptionBytes)
		_, err = r.Read(h.Options)
	}

	return h, err
}

// CalculateChecksum calculates the header checksum.
func (h Header) CalculateChecksum() uint16 {
	c := h
	c.Checksum = 0

	var b bytes.Buffer
	if err := c.Write(&b); err != nil {
		panic(err)
	}
	return tcpip.Checksum(b.Bytes())
}

// Check checks whether the IPv4 header is valid.
//
// The function fails if the checksum is not correct, or if the IHL field does
// not match the length of the Options byte slice.
//
// This functions can return ErrInvalidIHL, ErrInvalidChecksum or
// ErrInvalidTotalLength.
func (h Header) Check() error {
	if len(h.Options)%4 != 0 || int(h.IHL) != 5+len(h.Options)/4 {
		return ErrInvalidIHL
	} else if int(h.IHL)*4 > int(h.TotalLength) {
		return ErrInvalidTotalLength
	} else if h.CalculateChecksum() != h.Checksum {
		return ErrInvalidChecksum
	}
	return nil
}

// RawHeader converts the header to a RawHeader.
func (h Header) RawHeader() RawHeader {
	var header RawHeader
	header.VersionIHL = (h.Version << 4) | h.IHL
	header.ToS = h.ToS
	header.TotalLength = h.TotalLength
	header.Identification = h.Identification
	header.FlagsFragmentOffset = (uint16(h.Flags) << 13) | h.FragmentOffset
	header.TTL = h.TTL
	header.Protocol = h.Protocol
	header.Checksum = h.Checksum
	header.Source = h.Source
	header.Destination = h.Destination
	return header
}

// RawHeader represents a raw IPv4 header.
//
// This struct can be written and read with the binary package.
type RawHeader struct {
	VersionIHL          uint8
	ToS                 uint8
	TotalLength         uint16
	Identification      uint16
	FlagsFragmentOffset uint16
	TTL                 uint8
	Protocol            Protocol
	Checksum            uint16
	Source              Address
	Destination         Address
}

// Write the header to a Writer.
func (h RawHeader) Write(w io.Writer) error {
	return binary.Write(w, binary.BigEndian, h)
}

// NewRawHeader reads a new raw header from a reader.
func NewRawHeader(r io.Reader) (RawHeader, error) {
	var header RawHeader
	err := binary.Read(r, binary.BigEndian, &header)
	return header, err
}

// Header converts the RawHeader to a logic Header.
func (h RawHeader) Header() Header {
	var header Header
	header.Version = (h.VersionIHL >> 4) & 0x0F
	header.IHL = h.VersionIHL & 0x07
	header.ToS = h.ToS
	header.TotalLength = h.TotalLength
	header.Identification = h.Identification
	header.Flags = uint8(h.FlagsFragmentOffset>>13) & 0x0007
	header.FragmentOffset = h.FlagsFragmentOffset & 0x1FFF
	header.TTL = h.TTL
	header.Protocol = h.Protocol
	header.Checksum = h.Checksum
	header.Source = h.Source
	header.Destination = h.Destination
	return header
}

// Packet is an IPv4 packet.
type Packet struct {
	Header
	Payload []byte
}

// NewPacket will read a packet from a reader.
//
// The header will be checked using the Check() function. Only valid packets
// will be returned, unless err is not nil.
func NewPacket(r io.Reader) (p Packet, err error) {
	h, err := NewHeader(r)
	if err != nil {
		return
	}
	p.Header = h

	if err = p.Header.Check(); err != nil {
		return
	}
	payloadLength := int(p.Header.TotalLength) - int(p.Header.IHL)*4
	p.Payload = make([]byte, payloadLength)
	_, err = r.Read(p.Payload)
	return
}

var identificationCounter = uint32(rand.Int())

// NewDefaultPacket constructs a default IPv4 packet.
func NewDefaultPacket() Packet {
	ident := atomic.AddUint32(&identificationCounter, 1)
	return Packet{
		Header: Header{
			Version: 4, IHL: 5, ToS: 0, TotalLength: 20, TTL: 64,
			Identification: uint16(ident),
		},
	}
}

// NewPacketTo constructs a new packet with a destination.
func NewPacketTo(to Address, proto Protocol, payload []byte) Packet {
	p := NewDefaultPacket()
	p.Header.Destination = to
	p.Header.Protocol = proto
	p.Header.TotalLength = uint16(20 + len(payload))
	p.Payload = payload
	return p
}

// Write will write a packet to a writer.
func (p Packet) Write(w io.Writer) error {
	if err := p.Header.Write(w); err != nil {
		return err
	}
	_, err := w.Write(p.Payload)
	return err
}

// PayloadWriter can write a IPv4 packet payload.
type PayloadWriter interface {
	Write(io.Writer) error
}

// WritePayload will set the payload using a PayloadWriter.
//
// This function also updates the TotalLength field of the packet.
//
// If the PayloadWriter returns an error, this function panics.
func (packet *Packet) WritePayload(writer PayloadWriter) {
	b := bytes.NewBuffer(nil)
	if err := writer.Write(b); err != nil {
		panic(err)
	}
	packet.Payload = b.Bytes()
	packet.TotalLength = uint16(packet.IHL)*4 + uint16(len(packet.Payload))
}
