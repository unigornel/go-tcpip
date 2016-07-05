package icmp

import (
	"bytes"

	"github.com/unigornel/go-tcpip/common"
	"github.com/unigornel/go-tcpip/ipv4"
)

// Layer is the ICMP layer.
type Layer interface {
	Packets(p Type) <-chan Packet
}

type layer struct {
	ip       ipv4.Layer
	channels map[Type]chan Packet
}

// NewLayer creates a new instance of the default ICMP layer.
func NewLayer(ip ipv4.Layer) Layer {
	l := &layer{
		ip: ip,
	}
	go l.run()
	return l
}

func (layer *layer) Packets(t Type) <-chan Packet {
	c, ok := layer.channels[t]
	if !ok {
		c = make(chan Packet)
		layer.channels[t] = c
	}
	return c
}

func (layer *layer) run() {
	for packet := range layer.ip.Packets(ipv4.ProtocolICMP) {
		p, err := NewPacket(bytes.NewReader(packet.Payload))
		if err != nil {
			continue
		}

		p.Source = packet.Source

		switch p.Header.Type {
		case EchoRequestType:
			go layer.handleEchoRequest(p)
		default:
			c := layer.channels[p.Header.Type]
			if c != nil {
				c <- p
			}
		}
	}

	for _, c := range layer.channels {
		close(c)
	}
}

func (layer *layer) handleEchoRequest(packet Packet) {
	data := packet.Data.(Echo)
	reply := NewEchoReply(data.Header.Identifier, data.Header.SequenceNumber, data.Payload)
	p := ipv4.NewPacketTo(packet.Source, ipv4.ProtocolICMP, common.PacketToBytes(reply))
	layer.ip.Send(p)
}
