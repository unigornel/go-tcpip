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

const maxEthPayloadSize = 1500

type miniosNIC struct {
	tx   chan Packet
	rx   chan Packet
	done chan struct{}
}

// NewNIC creates a new NIC linked to the Mini-OS network.
func NewNIC() NIC {
	nic := new(miniosNIC)

	nic.tx = make(chan Packet)
	nic.rx = make(chan Packet)
	nic.done = make(chan struct{})

	go nic.sendAll()
	go nic.receiveAll()

	return nic
}

func (nic *miniosNIC) Close() {
	close(nic.tx)
	close(nic.done)
}

func (nic *miniosNIC) Send() chan<- Packet {
	return nic.tx
}

func (nic *miniosNIC) Receive() <-chan Packet {
	return nic.rx
}

func (nic *miniosNIC) GetMAC() MAC {
	buffer := (*C.uchar)(C.malloc(C.size_t(6)))
	defer C.free(unsafe.Pointer(buffer))
	C.get_mac_address(buffer)

	s := []byte(C.GoStringN((*C.char)(unsafe.Pointer(buffer)), 6))
	var mac MAC
	for i := 0; i < 6; i++ {
		mac[i] = s[i]
	}
	return mac
}

func (nic *miniosNIC) sendAll() {
	for p := range nic.tx {
		var packet C.struct_eth_packet
		for i := 0; i < 6; i++ {
			packet.destination[i] = C.uchar(p.Destination[i])
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

func (nic *miniosNIC) receiveAll() {
	for {
		var packet C.struct_eth_packet
		packet.payload = (*C.uchar)(C.malloc(C.size_t(maxEthPayloadSize)))
		packet.payload_length = C.uint(maxEthPayloadSize)

		i := C.receive_packet(&packet)
		if i != 0 {
			panic("could not receive packet")
		}

		var p Packet
		for i := 0; i < 6; i++ {
			p.Source[i] = byte(packet.source[i])
			p.Destination[i] = byte(packet.destination[i])
		}
		p.EtherType = EtherType(packet.ether_type)
		p.Payload = make([]byte, int(packet.payload_length))
		C.memcpy(unsafe.Pointer(&p.Payload[0]), unsafe.Pointer(packet.payload), C.size_t(packet.payload_length))
		C.free(unsafe.Pointer(packet.payload))

		select {
		case nic.rx <- p:
		case <-nic.done:
			close(nic.rx)
			return
		}
	}
}
