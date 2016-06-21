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
		data := p.Bytes()
		C.send_packet(unsafe.Pointer(&data[0]), C.int64_t(len(data)))
	}
}

func (nic *miniosNIC) receiveAll() {
	for {
		data := make([]byte, MaxPacketSize)

		i := C.receive_packet(unsafe.Pointer(&data[0]), C.int64_t(MaxPacketSize))
		if i < 0 {
			panic("could not receive packet")
		}

		packet, err := PacketFromBytes(data[:i])
		if err != nil {
			panic(err)
		}

		fmt.Println("Got packet:", packet)

		select {
		case nic.rx <- packet:
		case <-nic.done:
			close(nic.rx)
			return
		}
	}
}
