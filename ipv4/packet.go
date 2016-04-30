package ipv4

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.ugent.be/unigornel/go-tcpip"
)

var (
	// ErrInvalidIHL is a check error returned when the length of the Options
	// byte slice does not match the IHL value.
	ErrInvalidIHL = errors.New("IHL field does not match Options byte slice length")

	// ErrInvalidChecksum is a check error returned when header checksum is
	// incorrect.
	ErrInvalidChecksum = errors.New("Checksum field is incorrect")
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

// Protocol is an IPv4 protocol.
type Protocol uint8

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
// This functions can return ErrInvalidIHL or ErrInvalidChecksum.
func (h Header) Check() error {
	if len(h.Options)%4 != 0 || int(h.IHL) != 5+len(h.Options)/4 {
		return ErrInvalidIHL
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
