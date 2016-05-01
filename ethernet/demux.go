package ethernet

import (
	"log"
	"sync"
)

// DemuxOutput is a function that accepts incoming Ethernet packets.
type DemuxOutput func(Packet)

// DemuxDiscard is an output function that discards the frame.
func DemuxDiscard(Packet) {}

// DemuxLog is an output function that prints the logs the frame.
func DemuxLog(packet Packet) {
	log.Println("Received an Ethernet frame:", packet)
}

// Demux will demultiplex incoming Ethernet packets.
type Demux interface {
	SetOutput(EtherType, DemuxOutput)
}

type defaultDemux struct {
	sync.RWMutex
	outputs map[EtherType]DemuxOutput
}

// NewDemux creates an Ethernet demultiplexer with a default output function.
func NewDemux(incoming <-chan Packet, defaultOutput DemuxOutput) Demux {
	demux := &defaultDemux{
		outputs: make(map[EtherType]DemuxOutput),
	}

	demux.outputs[EtherType(0)] = defaultOutput
	go demux.receiveAll(incoming)

	return demux
}

// SetOutput sets an output function for a specific EtherType.
func (demux *defaultDemux) SetOutput(etherType EtherType, output DemuxOutput) {
	if etherType.IsLength() {
		panic("must be a true EtherType, not a payload length")
	}

	demux.RWMutex.Lock()
	demux.outputs[etherType] = output
	demux.RWMutex.Unlock()
}

func (demux *defaultDemux) receiveAll(incoming <-chan Packet) {
	for p := range incoming {

		demux.RWMutex.RLock()
		output, ok := demux.outputs[p.EtherType]
		if !ok || p.EtherType.IsLength() {
			output = demux.outputs[EtherType(0)]
		}
		demux.RWMutex.RUnlock()

		go output(p)
	}
}
