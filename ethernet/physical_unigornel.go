package ethernet

// #cgo LDFLAGS: -Wl,--unresolved-symbols=ignore-in-object-files
// #include <mini-os/network.h>
// #include <mini-os/types.h>
//
// extern void *memcpy(void *, const void *, size_t);
// extern void *malloc(size_t);
// extern void free(void *);
import "C"
import (
	"fmt"
	"unsafe"
)

const (
	ethHeaderSize     = 6 + 6 + 2
	maxEthPayloadSize = 1500
	maxEthPacketSize  = ethHeaderSize + maxEthPayloadSize
)

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
		var packet Packet
		data := make([]byte, maxEthPacketSize)

		i := C.receive_packet(unsafe.Pointer(&data[0]), C.int64_t(maxEthPacketSize))
		if i < ethHeaderSize {
			panic("could not receive packet")
		}

		for i, _ := range packet.Destination {
			packet.Destination[i] = data[i]
			packet.Source[i] = data[i+6]
		}

		packet.EtherType = EtherType(data[12])<<8 | EtherType(data[13])
		packet.Payload = data[14:i]

		fmt.Println("Got packet:", packet)

		select {
		case nic.rx <- packet:
		case <-nic.done:
			close(nic.rx)
			return
		}
	}
}
