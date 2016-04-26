package ethernet

// #cgo LDFLAGS: -Wl,--unresolved-symbols=ignore-in-object-files
// #include <mini-os/network.h>
// #include <mini-os/types.h>
//
// extern void *memcpy(void *, const void *, size_t);
// extern void *malloc(size_t);
// extern void free(void *);
import "C"
import "unsafe"

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
	for p := range nic.tx {
		var packet C.struct_eth_packet
		for i := 0; i < 6; i++ {
			packet.destination[i] = C.char(p.Destination[i])
		}
		packet.ether_type = C.uint16_t(p.EtherType)
		packet.payload_length = C.uint(len(p.Payload))
		if len(p.Payload) == 0 {
			packet.payload = nil
		} else {
			packet.payload = (*C.uchar)(C.malloc(C.size_t(len(p.Payload))))
			defer C.free(unsafe.Pointer(packet.payload))
			C.memcpy(unsafe.Pointer(packet.payload), unsafe.Pointer(&p.Payload[0]), C.size_t(len(p.Payload)))
		}

		C.send_packet(&packet)
	}
}
