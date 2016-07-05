package ipv4

import (
	"errors"

	"github.com/unigornel/go-tcpip/ethernet"
)

var (
	// ErrNoRouteToDestinationAddress is returned when there is no route
	// to the destination address.
	ErrNoRouteToDestinationAddress = errors.New("no route to destination address")
)

// Router is an IPv4 router.
type Router interface {
	// Resolve will resolve the IPv4 address to the ethernet MAC address
	// of the next hop.
	//
	// See also ErrNoRouteToDestinationAddress.
	Resolve(address Address) (ethernet.MAC, error)
}

type router struct {
	arp     ARP
	address Address
	netmask Address
	gateway *Address
}

// NewRouter creates a default router.
//
// ARP is used to resolve local addresses. Otherwise, the MAC address of
// the gateway is returned.
//
// Specifying a gateway is optional.
func NewRouter(arp ARP, address, netmask Address, gateway *Address) Router {
	return &router{
		arp:     arp,
		address: address,
		netmask: netmask,
		gateway: gateway,
	}
}

func (r *router) Resolve(address Address) (mac ethernet.MAC, err error) {
	if r.isLocal(address) {
		mac, err = r.arp.Resolve(address)
		if err == ErrARPTimeout {
			err = ErrNoRouteToDestinationAddress
		}

	} else if address.Equals(Broadcast) {
		mac = ethernet.MulticastIPv4
		mac[5] &= address[3]
		mac[4] &= address[2]
		mac[3] &= (address[1] & 0x7F)

	} else if r.gateway != nil {
		mac, err = r.arp.Resolve(*r.gateway)
		if err == ErrARPTimeout {
			err = ErrNoRouteToDestinationAddress
		}
	} else {
		err = ErrNoRouteToDestinationAddress
	}

	return
}

func (r *router) isLocal(address Address) bool {
	a := r.address.And(r.netmask)
	b := address.And(r.netmask)
	return a.Equals(b)
}
