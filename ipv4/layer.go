package ipv4

import (
	"bytes"

	"github.com/unigornel/go-tcpip/common"
	"github.com/unigornel/go-tcpip/ethernet"
)

// Layer is an IPv4 layer.
type Layer interface {
	Packets(p Protocol) <-chan Packet
	Send(t Packet) error
}

type layer struct {
	address  Address
	arp      ARP
	eth      ethernet.Layer
	channels map[Protocol]chan Packet
}

// NewLayer creates a new instance of the default IPv4 layer.
func NewLayer(address Address, arp ARP, eth ethernet.Layer) Layer {
	l := &layer{
		address:  address,
		arp:      arp,
		eth:      eth,
		channels: make(map[Protocol]chan Packet),
	}
	go l.run()
	return l
}

func (layer *layer) Packets(t Protocol) <-chan Packet {
	c, ok := layer.channels[t]
	if !ok {
		c = make(chan Packet)
		layer.channels[t] = c
	}
	return c
}

func (layer *layer) Send(t Packet) error {
	mac, err := layer.arp.Resolve(t.Destination)
	if err != nil {
		return err
	}

	t.Source = layer.address
	t.Checksum = t.CalculateChecksum()
	frame := ethernet.Packet{
		Destination: mac,
		EtherType:   ethernet.EtherTypeIPv4,
		Payload:     common.PacketToBytes(t),
	}
	return layer.eth.Send(frame)
}

func (layer *layer) run() {
	for frame := range layer.eth.Packets(ethernet.EtherTypeIPv4) {
		p, err := NewPacket(bytes.NewBuffer(frame.Payload))
		if err != nil {
			continue
		}

		c := layer.channels[p.Protocol]
		if c != nil {
			c <- p
		}
	}
}
