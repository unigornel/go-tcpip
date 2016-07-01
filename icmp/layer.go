package icmp

import (
	"bytes"

	"github.com/unigornel/go-tcpip/common"
	"github.com/unigornel/go-tcpip/ipv4"
)

type Layer interface {
}

type layer struct {
	ip ipv4.Layer
}

// NewLayer creates a new instance of the default ICMP layer.
func NewLayer(ip ipv4.Layer) Layer {
	l := &layer{
		ip: ip,
	}
	go l.run()
	return l
}

func (layer *layer) run() {
	for packet := range layer.ip.Packets(ipv4.ProtocolICMP) {
		p, err := NewPacket(bytes.NewReader(packet.Payload))
		if err != nil {
			continue
		}

		switch p.Header.Type {
		case EchoRequestType:
			go layer.handleEchoRequest(packet.Source, p)
		default:
		}
	}
}

func (layer *layer) handleEchoRequest(source ipv4.Address, packet Packet) {
	data := packet.Data.(Echo)
	reply := NewEchoReply(data.Header.Identifier, data.Header.SequenceNumber, data.Payload)
	p := ipv4.NewPacketTo(source, ipv4.ProtocolICMP, common.PacketToBytes(reply))
	layer.ip.Send(p)
}
