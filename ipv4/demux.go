package ipv4

import (
	"log"
	"sync"
)

// DemuxOutput is a function that accepts incoming IPv4 packets.
type DemuxOutput func(Packet)

// DemuxDiscard is an output function that discards the packet.
func DemuxDiscard(Packet) {}

// DemuxLog is an output function that prints the logs the packet.
func DemuxLog(packet Packet) {
	log.Println("Received an IPv4 packet:", packet)
}

// Demux will demultiplex incoming IPv4 packets.
type Demux interface {
	SetOutput(Protocol, DemuxOutput)
}

type defaultDemux struct {
	sync.RWMutex
	outputs map[Protocol]DemuxOutput
}

// NewDemux creates an IPv4 demultiplexer with a default output function.
func NewDemux(incoming <-chan Packet, defaultOutput DemuxOutput) Demux {
	demux := &defaultDemux{
		outputs: make(map[Protocol]DemuxOutput),
	}

	// IPv4 protocol number 255 is officially reserved, will is it.
	demux.outputs[255] = defaultOutput

	go demux.receiveAll(incoming)

	return demux
}

// SetOutput sets an output function for a specific IPv4 protocol.
func (demux *defaultDemux) SetOutput(protocol Protocol, output DemuxOutput) {
	demux.RWMutex.Lock()
	demux.outputs[protocol] = output
	demux.RWMutex.Unlock()
}

func (demux *defaultDemux) receiveAll(incoming <-chan Packet) {
	for p := range incoming {
		demux.RWMutex.RLock()
		output, ok := demux.outputs[p.Protocol]
		if !ok {
			output = demux.outputs[255]
		}
		demux.RWMutex.RUnlock()

		go output(p)
	}
}
