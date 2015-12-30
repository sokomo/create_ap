package main

import (
	"fmt"
	"net"
)

type IPNet struct {
	*net.IPNet
}

func parseIPv4(s string) (*IPNet, error) {
	// First try to parse it as CIDR format.
	ip, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		// If IP is not in CIDR format, parse it as a normal one.
		ip = net.ParseIP(s)
		if ip == nil {
			return nil, fmt.Errorf("Invalid IPv4 address (%s)", s)
		}
	}

	// Make sure that it's an IPv4
	if ip = ip.To4(); ip == nil {
		return nil, fmt.Errorf("Invalid IPv4 address (%s)", s)
	}

	newIPNet := &net.IPNet{IP: ip}
	if ipnet != nil {
		// IP was an CIDR format
		newIPNet.Mask = ipnet.Mask
	} else {
		newIPNet.Mask = ip.DefaultMask()
	}

	return &IPNet{newIPNet}, nil
}

func (ipnet *IPNet) network() *IPNet {
	ip := ipnet.IP.Mask(ipnet.Mask)

	return &IPNet{
		&net.IPNet{
			IP:   net.IPv4(ip[0], ip[1], ip[2], ip[3]).To4(),
			Mask: ipnet.Mask,
		},
	}
}

func (ipnet *IPNet) broadcast() net.IP {
	network := ipnet.network()
	broadcast := make(net.IP, net.IPv4len)

	for i := range broadcast {
		broadcast[i] = network.IP[i] | ^ipnet.Mask[i]
	}

	return broadcast
}

func (ipnet *IPNet) hostmin() net.IP {
	ip := ipnet.network().IP

	// increase by 1
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}

	return ip
}

func (ipnet *IPNet) hostmax() net.IP {
	ip := ipnet.broadcast()

	// decrease by 1
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]--
		if ip[i] != 255 {
			break
		}
	}

	return ip
}

func (ipnet *IPNet) netmask() net.IP {
	ip := ipnet.Mask
	return net.IPv4(ip[0], ip[1], ip[2], ip[3]).To4()
}
