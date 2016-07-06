package udp

import (
	"bytes"

	"github.com/unigornel/go-tcpip/common"
	"github.com/unigornel/go-tcpip/ipv4"
)

// Layer is an UDP layer.
type Layer interface {
	Packets(port uint16) <-chan Packet
	Send(packet Packet) error
}

type layer struct {
	ip       ipv4.Layer
	channels map[uint16]chan Packet
}

// NewLayer creates a new instance of the default UDP layer.
func NewLayer(ip ipv4.Layer) Layer {
	l := &layer{
		ip:       ip,
		channels: make(map[uint16]chan Packet),
	}
	go l.run()
	return l
}

func (layer *layer) Packets(port uint16) <-chan Packet {
	c, ok := layer.channels[port]
	if !ok {
		c = make(chan Packet)
		layer.channels[port] = c
	}
	return c
}

func (layer *layer) Send(packet Packet) error {
	packet.Checksum = packet.CalculateChecksum()
	payload := common.PacketToBytes(packet)
	p := ipv4.NewPacketTo(packet.Address, ipv4.ProtocolUDP, payload)
	return layer.ip.Send(p)
}

func (layer *layer) run() {
	for packet := range layer.ip.Packets(ipv4.ProtocolUDP) {
		p, err := NewPacket(bytes.NewBuffer(packet.Payload))
		if err != nil {
			continue
		}

		p.Address = packet.Source
		c := layer.channels[p.DestinationPort]
		if c != nil {
			c <- p
		}
	}

	for _, c := range layer.channels {
		close(c)
	}
}
