package ethernet

// NIC is a general network interface controller.
//
// A NIC implements all methods needed to send and receive Ethernet packets or
// to setup a network connectivity.
type NIC interface {
	Send() chan<- Packet
	Receive() <-chan Packet
	GetMAC() MAC
}
