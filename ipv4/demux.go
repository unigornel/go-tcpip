package ipv4

import "sync"

// DemuxOutput is a function that accepts incoming IPv4 packets.
type DemuxOutput func(Packet)

// Demux will demultiplex incoming IPv4 packets.
type Demux struct {
	sync.RWMutex
	outputs map[Protocol]DemuxOutput
}

// NewDemux creates an IPv4 demultiplexer with a default output function.
func NewDemux(incoming <-chan Packet, defaultOutput DemuxOutput) *Demux {
	demux := &Demux{
		outputs: make(map[Protocol]DemuxOutput),
	}

	// IPv4 protocol number 255 is officially reserved, will is it.
	demux.outputs[255] = defaultOutput

	return demux
}

// SetOutput sets an output function for a specific IPv4 protocol.
func (demux *Demux) SetOutput(protocol Protocol, output DemuxOutput) {
	demux.RWMutex.Lock()
	demux.outputs[protocol] = output
	demux.RWMutex.Unlock()
}

func (demux *Demux) receiveAll(incoming <-chan Packet) {
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
