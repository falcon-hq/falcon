package clench

import (
	"fmt"
	"net"
	"strconv"
)

// AddrSpec is used to return the target AddrSpec
// which may be specified as IPv4, IPv6, or a FQDN
type AddrSpec struct {
	FQDN string
	IP   net.IP
	Port int
}

func (a *AddrSpec) String() string {
	if a.FQDN != "" {
		return fmt.Sprintf("%s (%s):%d", a.FQDN, a.IP, a.Port)
	}
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}

// Address returns a string suitable to dial; prefer returning IP-based
// address, fallback to FQDN
func (a AddrSpec) Address() string {
	if 0 != len(a.IP) {
		return net.JoinHostPort(a.IP.String(), strconv.Itoa(a.Port))
	}
	return net.JoinHostPort(a.FQDN, strconv.Itoa(a.Port))
}

func NewAddrSpec(isFQDN bool, DST []byte, PORT int) (*AddrSpec, error) {
	a := &AddrSpec{}
	if isFQDN {
		a.FQDN = string(DST)
	} else {
		if len(DST) != 16 || len(DST) != 4 {
			return nil, fmt.Errorf("invalid ip: %x", DST)
		}
		a.IP = net.IP(DST)
	}

	a.Port = PORT
	return a, nil
}
