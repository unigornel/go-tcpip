package ipv4

import (
	"bytes"
	"log"

	"github.ugent.be/unigornel/go-tcpip/ethernet"
)

// Layer is an IPv4 layer.
type Layer interface {
	Send() chan<- Packet
	Receive() <-chan Packet
	GetIP() Address
	Bind(*ethernet.Demux)
	Close()
}

type defaultLayer struct {
	address Address
	tx      chan Packet
	rx      chan Packet
}

// NewLayer creates a new instance of the default IPv4 layer.
func NewLayer(address Address, out chan<- ethernet.Packet) Layer {
	layer := &defaultLayer{
		address: address,
		tx:      make(chan Packet),
		rx:      make(chan Packet),
	}

	return layer
}

func (layer *defaultLayer) Send() chan<- Packet {
	return layer.tx
}

func (layer *defaultLayer) Receive() <-chan Packet {
	return layer.rx
}

func (layer *defaultLayer) GetIP() Address {
	return layer.address
}
func (layer *defaultLayer) Bind(demux *ethernet.Demux) {
	demux.SetOutput(ethernet.EtherTypeIPv4, func(packet ethernet.Packet) {
		r := bytes.NewReader(packet.Payload)
		p, err := NewPacket(r)
		if err != nil {
			log.Println("Dropping IPv4 packet:", err)
			return
		}

		layer.rx <- p
	})
}

func (layer *defaultLayer) Close() {
	close(layer.tx)
	close(layer.rx)
}

func (layer *defaultLayer) sendAll(out chan<- ethernet.Packet) {
	for _ = range layer.tx {
		panic("missing arp")
	}
}
