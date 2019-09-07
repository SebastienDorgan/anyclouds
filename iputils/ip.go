package iputils

import (
	"encoding/binary"
	"net"
)

func Itou(ip *net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}

func Utoi(u uint32) *net.IP {
	ip := net.IPv4(0, 0, 0, 0)
	binary.BigEndian.PutUint32(ip[12:16], u)
	return &ip
}

func NextIP(ip *net.IP) *net.IP {
	return Utoi(Itou(ip) + 1)
}

func PreviousIP(ip *net.IP) *net.IP {
	return Utoi(Itou(ip) - 1)
}

type AddressRange struct {
	FirstIP net.IP
	LastIP  net.IP
}

func GetRange(cidr string) (*AddressRange, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	first := NextIP(&ip)
	var last *net.IP
	for last = first; ipnet.Contains(*NextIP(last)); {
		last = NextIP(last)
	}
	return &AddressRange{
		FirstIP: *first,
		LastIP:  *PreviousIP(last),
	}, nil
}
