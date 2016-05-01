package icmp

import (
	"bytes"
	"log"

	"github.ugent.be/unigornel/go-tcpip/ipv4"
)

// Layer is an ICMP layer.
type Layer interface {
	Bind(ipv4.Demux)
}

type defaultLayer struct {
	tx chan<- ipv4.Packet
}

// NewLayer creates a new instance of the default ICMP layer.
func NewLayer(tx chan<- ipv4.Packet) Layer {
	return &defaultLayer{
		tx: tx,
	}
}

// Bind will bind the ICMP layer to the IPv4 layer.
func (icmp *defaultLayer) Bind(demux ipv4.Demux) {
	demux.SetOutput(ipv4.ProtocolICMP, func(packet ipv4.Packet) {
		r := bytes.NewReader(packet.Payload)
		p, err := NewPacket(r)
		if err != nil {
			log.Println("Dropping ICMP packet:", err)
			return
		}

		switch p.Header.Type {
		case EchoRequestType:
			go icmp.handleEchoRequest(packet.Source, p)
		default:
			log.Println("Dropping ICMP packet with unknown type:", p)
		}
	})
}

func (icmp *defaultLayer) handleEchoRequest(source ipv4.Address, packet Packet) {
	data := packet.Data.(Echo)
	reply := NewEchoReply(data.Header.Identifier, data.Header.SequenceNumber, data.Payload)
	p := ipv4.NewPacketTo(source, ipv4.ProtocolICMP, nil)
	p.WritePayload(reply)
	log.Println("Sending ICMP reply", p)
	icmp.tx <- p
}
