package ethernet

import "fmt"

// MAC is a 48-bit long MAC address.
type MAC [6]byte

func (mac MAC) String() string {
	return fmt.Sprintf("%x:%x:%x:%x:%x:%x", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}

// Broadcast is the Ethernet broadcast address
var Broadcast = MAC([6]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

// EtherType is either the Ethernet packet type or the Eternet packet length.
type EtherType uint16

// Packet is an Ethernet packet.
type Packet struct {
	Destination MAC
	Source      MAC
	EtherType   EtherType
	Payload     []byte
}

func (packet Packet) String() string {
	return fmt.Sprintf(
		"Packet{%v -> %v, EtherType: %v, %v",
		packet.Source, packet.Destination,
		packet.EtherType, packet.Payload,
	)
}
