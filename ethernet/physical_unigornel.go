package ethernet

// #cgo LDFLAGS: -Wl,--unresolved-symbols=ignore-in-object-files
// #include <mini-os/network.h>
// #include <mini-os/types.h>
//
// extern void *memcpy(void *, const void *, size_t);
// extern void *malloc(size_t);
// extern void free(void *);
import "C"

type miniosNIC struct {
	tx chan Packet
	rx chan Packet
}

// NewNIC creates a new NIC linked to the Mini-OS network.
func NewNIC() NIC {
	nic := new(miniosNIC)

	nic.tx = make(chan Packet)
	nic.rx = make(chan Packet)
	go nic.sendAll()

	return nic
}

func (nic *miniosNIC) Close() {
	if nic.tx == nil || nic.rx == nil {
		panic("NIC not started")
	}

	close(nic.tx)
	close(nic.rx)
}

func (nic *miniosNIC) Send() chan<- Packet {
	return nic.tx
}

func (nic *miniosNIC) Receive() <-chan Packet {
	panic("not implemented")
}

func (nic *miniosNIC) GetMAC() MAC {
	panic("not implemented")
}

func (nic *miniosNIC) sendAll() {
}
