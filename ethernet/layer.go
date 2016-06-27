package ethernet

// Layer is the ethernet receive layer
type Layer interface {
	Packets(t EtherType) <-chan Packet
	Send(t Packet) error
}

// NewLayer will receive packets from a NIC.
func NewLayer(nic NIC) Layer {
	return &layer{
		mac:      nic.GetMAC(),
		nic:      nic,
		channels: make(map[EtherType]chan Packet),
	}
}

type layer struct {
	nic      NIC
	mac      MAC
	channels map[EtherType]chan Packet
}

func (layer *layer) Packets(t EtherType) <-chan Packet {
	c, ok := layer.channels[t]
	if !ok {
		c = make(chan Packet)
		layer.channels[t] = c
	}
	return c
}

func (layer *layer) Send(packet Packet) error {
	packet.Source = layer.mac
	layer.nic.Send() <- packet
	return nil
}

func (layer *layer) run() {
	for p := range layer.nic.Receive() {
		c := layer.channels[p.EtherType]

		if c != nil {
			c <- p
		}
	}
}
