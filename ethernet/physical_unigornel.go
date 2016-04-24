package ethernet

type miniosNIC struct{}

// NewNIC creates a new NIC linked to the Mini-OS network.
func NewNIC() NIC {
	return new(miniosNIC)
}

func (nic *miniosNIC) Send() chan<- Packet {
	panic("not implemented")
}

func (nic *miniosNIC) Receive() <-chan Packet {
	panic("not implemented")
}

func (nic *miniosNIC) GetMAC() MAC {
	panic("not implemented")
}
