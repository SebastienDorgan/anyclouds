package iputils

import (
	"encoding/binary"
	"net"
)

//Itou converts an IP v4 into an uint32
func Itou(ip *net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}

//Utoi converts a uint32 into an IP v4
func Utoi(u uint32) *net.IP {
	ip := net.IPv4(0, 0, 0, 0)
	binary.BigEndian.PutUint32(ip[12:16], u)
	return &ip
}

//NextIP gives the next IP v4 of a given IP v4
func NextIP(ip *net.IP) *net.IP {
	return Utoi(Itou(ip) + 1)
}

//PreviousIP gives the previous IP v4 of a given IP v4
func PreviousIP(ip *net.IP) *net.IP {
	return Utoi(Itou(ip) - 1)
}

//IPAddressRange a contigues range of IP address
type IPAddressRange struct {
	FirstIP net.IP
	LastIP  net.IP
}

//GetRange computes the IP address range of a given CIDR
func GetRange(cidr string) (*IPAddressRange, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	first := NextIP(&ip)
	var last *net.IP
	for last = first; ipnet.Contains(*NextIP(last)); {
		last = NextIP(last)
	}
	return &IPAddressRange{
		FirstIP: *first,
		LastIP:  *PreviousIP(last),
	}, nil
}
