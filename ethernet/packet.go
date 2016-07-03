package ethernet

import (
	"errors"
	"fmt"
)

// MACLength is 48-bits or 6 bytes
const MACLength = 6

// MAC is a 48-bit long MAC address.
type MAC [MACLength]byte

func (mac MAC) String() string {
	return fmt.Sprintf("%x:%x:%x:%x:%x:%x", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}

// Broadcast is the Ethernet broadcast address
var Broadcast = MAC([6]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

// EtherType is either the Ethernet packet type or the Eternet packet length.
type EtherType uint16

const (
	// EtherTypeIPv4 is the EtherType for IPv4 frames.
	EtherTypeIPv4 = 0x0800

	// EtherTypeARP is the EtherType for ARP frames.
	EtherTypeARP = 0x0806
)

// IsLength determines if the EtherType field contains a frame type or the
// length of the frame payload.
func (etherType EtherType) IsLength() bool {
	return etherType <= 1500
}

const (
	// HeaderSize is the size of the ethernet packet header.
	HeaderSize = 6 + 6 + 2

	// MaxPayloadSize is the maximum payload size of a standard ethernet packet.
	MaxPayloadSize = 1500

	// MaxPacketSize is the maximum packet size of a standard ethernet packet,
	// including header and payload
	MaxPacketSize = HeaderSize + MaxPayloadSize
)

// Packet is an Ethernet packet.
type Packet struct {
	Destination MAC
	Source      MAC
	EtherType   EtherType
	Payload     []byte
}

// PacketFromBytes constructs an ethernet packet from a byte slice.
func PacketFromBytes(data []byte) (Packet, error) {
	var packet Packet

	if len(data) < HeaderSize {
		return packet, errors.New("packet size too small")
	}

	for i := 0; i < MACLength; i++ {
		packet.Destination[i] = data[i]
		packet.Source[i] = data[i+6]
	}

	packet.EtherType = EtherType(data[12])<<8 | EtherType(data[13])
	packet.Payload = data[14:]

	return packet, nil
}

// Bytes converts an ethernet packet to a byte slice.
func (packet Packet) Bytes() []byte {
	data := make([]byte, HeaderSize+len(packet.Payload))

	for i := 0; i < MACLength; i++ {
		data[i] = packet.Destination[i]
		data[i+6] = packet.Source[i]
	}

	data[12] = byte(packet.EtherType >> 8)
	data[13] = byte(packet.EtherType)

	for i, b := range packet.Payload {
		data[14+i] = b
	}

	return data
}

func (packet Packet) String() string {
	return fmt.Sprintf(
		"Packet{%v -> %v, EtherType: %v, %v",
		packet.Source, packet.Destination,
		packet.EtherType, packet.Payload,
	)
}
