package ipv4

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"

	"github.ugent.be/unigornel/go-tcpip/ethernet"
)

// ARP represents an ARP layer that can convert IPv4 addresses to Ethernet
// addresses.
type ARP interface {
	Bind(ethernet.Demux)
	Resolve(address Address) (ethernet.MAC, error)
}

// ARPOperation is a type of ARP packet.
type ARPOperation uint16

const (
	// ARPRequest for ARP requests.
	ARPRequest = 1
	// ARPReply for ARP replies.
	ARPReply = 2
)

// ARPHardwareType is the network protocol of the ARP packet.
type ARPHardwareType uint16

const (
	// ARPHardwareEthernet as lower layer.
	ARPHardwareEthernet = 1
)

// ARPProtocolType is the internetwork protocol of the ARP packet.
type ARPProtocolType uint16

const (
	// ARPProtocolIPv4 as upper layer.
	ARPProtocolIPv4 = 0x800
)

// ARPPacket represents an ARP packet
type ARPPacket struct {
	HardwareType          ARPHardwareType
	ProtocolType          ARPProtocolType
	HardwareAddressLength uint8
	ProtocolAddressLength uint8
	Operation             ARPOperation
	SenderHardwareAddress ethernet.MAC
	SenderProtocolAddress Address
	TargetHardwareAddress ethernet.MAC
	TargetProtocolAddress Address
}

// NewARPPacket reads an ARP packet from a reader.
func NewARPPacket(r io.Reader) (ARPPacket, error) {
	var p ARPPacket
	err := binary.Read(r, binary.BigEndian, &p)
	return p, err
}

// NewARPRequest creates an ARP request packet.
func NewARPRequest(senderMAC ethernet.MAC, senderIP, targetIP Address) ARPPacket {
	return ARPPacket{
		HardwareType:          ARPHardwareEthernet,
		ProtocolType:          ARPProtocolIPv4,
		HardwareAddressLength: 6,
		ProtocolAddressLength: 4,
		Operation:             ARPRequest,
		SenderHardwareAddress: senderMAC,
		SenderProtocolAddress: senderIP,
		TargetProtocolAddress: targetIP,
	}
}

func NewARPReply(senderMAC, targetMAC ethernet.MAC, senderIP, targetIP Address) ARPPacket {
	return ARPPacket{
		HardwareType:          ARPHardwareEthernet,
		ProtocolType:          ARPProtocolIPv4,
		HardwareAddressLength: 6,
		ProtocolAddressLength: 4,
		Operation:             ARPReply,
		SenderHardwareAddress: senderMAC,
		TargetHardwareAddress: targetMAC,
		SenderProtocolAddress: senderIP,
		TargetProtocolAddress: targetIP,
	}
}

// Write an ARP packet to a writer.
func (p ARPPacket) Write(w io.Writer) error {
	return binary.Write(w, binary.BigEndian, &p)
}

type pendingARPRequest struct {
	gotReply chan struct{}
	timeout  chan struct{}
}

type defaultARP struct {
	sourceMAC     ethernet.MAC
	sourceIP      Address
	tx            chan<- ethernet.Packet
	cache         *cache.Cache
	queryInterval time.Duration
	timeout       int

	requestsLock sync.RWMutex
	requests     map[Address]*pendingARPRequest
}

const (
	// DefaultARPExpiration is the default expiration for entries in the ARP table.
	DefaultARPExpiration = 4 * time.Hour

	// DefaultARPCleanupInterval is the cleanup interval for the cache.
	DefaultARPCleanupInterval = DefaultARPExpiration

	// DefaultARPQueryInterval is the default query interval to use when sending
	// ARP requests.
	DefaultARPQueryInterval = 1 * time.Second

	// DefaultARPTimeout is the default number of ARP requests after which to
	// give up on an ARP request.
	DefaultARPTimeout = 3
)

var (
	// ErrARPTimeout occurs when no ARP reply is received for an ARP request.
	ErrARPTimeout = errors.New("ARP request timeout")
)

// NewARP will create a default ARP interface with the default configuration.
func NewARP(mac ethernet.MAC, ip Address, tx chan<- ethernet.Packet) ARP {
	return NewCustomARP(
		mac, ip, tx,
		DefaultARPExpiration,
		DefaultARPCleanupInterval,
		DefaultARPQueryInterval,
		DefaultARPTimeout,
	)
}

// NewCustomARP will create a default ARP interface with a custom configuration.
func NewCustomARP(mac ethernet.MAC, ip Address, tx chan<- ethernet.Packet, expiration, cleanupInterval, queryInterval time.Duration, timeout int) ARP {
	return &defaultARP{
		sourceMAC:     mac,
		sourceIP:      ip,
		tx:            tx,
		cache:         cache.New(expiration, cleanupInterval),
		queryInterval: queryInterval,
		timeout:       timeout,
		requests:      make(map[Address]*pendingARPRequest),
	}
}

func (arp *defaultARP) Bind(demux ethernet.Demux) {
	demux.SetOutput(ethernet.EtherTypeARP, func(packet ethernet.Packet) {
		r := bytes.NewReader(packet.Payload)
		p, err := NewARPPacket(r)
		if err != nil {
			log.Println("Dropping ARP packet:", err)
			return
		}

		switch p.Operation {
		case ARPRequest:
			go arp.handleRequest(p)
		case ARPReply:
			go arp.handleReply(p)
		default:
			log.Println("Dropping ARP packet with unknown operation:", p)
		}
	})
}

func (arp *defaultARP) Resolve(address Address) (ethernet.MAC, error) {
	item, _ := arp.cache.Get(address.String())
	if item != nil {
		return item.(ethernet.MAC), nil
	}
	return arp.arpResolve(address)
}

func (arp *defaultARP) arpResolve(address Address) (mac ethernet.MAC, err error) {
	arp.requestsLock.RLock()
	pending, ok := arp.requests[address]
	arp.requestsLock.RUnlock()

	if !ok {
		arp.requestsLock.Lock()
		pending, ok = arp.requests[address]
		if !ok {
			pending = &pendingARPRequest{
				gotReply: make(chan struct{}),
				timeout:  make(chan struct{}),
			}
			arp.requests[address] = pending
			go arp.sendARPRequestAndNotify(pending, address)
		}
		arp.requestsLock.Unlock()
	}

	select {
	case <-pending.gotReply:
	case <-pending.timeout:
	}

	if !ok {
		arp.requestsLock.Lock()
		delete(arp.requests, address)
		arp.requestsLock.Unlock()
	}

	item, _ := arp.cache.Get(address.String())
	if item == nil {
		err = ErrARPTimeout
	}
	mac = item.(ethernet.MAC)
	return
}

func (arp *defaultARP) sendARPRequestAndNotify(pending *pendingARPRequest, address Address) {
	p := ethernet.Packet{
		Destination: ethernet.Broadcast,
		EtherType:   ethernet.EtherTypeARP,
	}
	p.WritePayload(NewARPRequest(arp.sourceMAC, arp.sourceIP, address))

	flag := false
	for i := 0; i < arp.timeout; i++ {
		arp.tx <- p
		select {
		case <-time.After(arp.queryInterval):
			continue
		case <-pending.gotReply:
			flag = true
			break
		}
	}

	if !flag {
		close(pending.timeout)
	}
}

func (arp *defaultARP) handleRequest(request ARPPacket) {
	if bytes.Equal(request.TargetProtocolAddress.Bytes(), arp.sourceIP.Bytes()) {
		reply := NewARPReply(
			arp.sourceMAC,
			request.SenderHardwareAddress,
			arp.sourceIP,
			request.SenderProtocolAddress,
		)
		p := ethernet.Packet{
			Destination: request.SenderHardwareAddress,
			EtherType:   ethernet.EtherTypeARP,
		}
		p.WritePayload(reply)
		arp.tx <- p
	}
}

func (arp *defaultARP) handleReply(request ARPPacket) {
	arp.requestsLock.Lock()
	defer arp.requestsLock.Unlock()

	ip := request.SenderProtocolAddress
	mac := request.SenderHardwareAddress
	pending, ok := arp.requests[ip]
	if ok {
		arp.cache.Set(ip.String(), mac, cache.DefaultExpiration)
		if pending.gotReply != nil {
			close(pending.gotReply)
			pending.gotReply = nil
		}
	}
}
