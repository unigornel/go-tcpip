package ethernet

// MAC is a 48-bit long MAC address.
type MAC [6]byte

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
