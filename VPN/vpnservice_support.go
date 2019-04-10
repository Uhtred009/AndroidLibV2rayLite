package VPN

import (
	"errors"
	"log"
	"net"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

type vpnProtectedDialer struct {
	vp *VPNSupport
}

func (sDialer *vpnProtectedDialer) Dial(network, Address string) (net.Conn, error) {
	if strings.HasPrefix(network, "tcp") {

		var addr *net.TCPAddr
		var err error

		addr, haveaddr := sDialer.vp.prepareddomain.tcpprepared[Address]

		if haveaddr == false {
			log.Println("Not Using Prepared: TCP,", Address)
			addr, err = net.ResolveTCPAddr(network, Address)
		} else {
			log.Println("Using Prepared: TCP,", Address)
		}

		if err != nil {
			return nil, err
		}
		fd, err := unix.Socket(unix.AF_INET6, unix.SOCK_STREAM, unix.IPPROTO_TCP)
		if err != nil {
			return nil, err
		}

		//Protect socket fd!
		//log.Println("Protecting Sock:", fd)
		sDialer.vp.VpnSupportSet.Protect(fd)

		sa := new(unix.SockaddrInet6)
		sa.Port = addr.Port
		sa.ZoneId = uint32(zoneToInt(addr.Zone))
		//fmt.Println(addr.IP.To16())
		copy(sa.Addr[:], addr.IP.To16())
		//fmt.Println(sa.Addr)
		err = unix.Connect(fd, sa)
		if err != nil {
			return nil, err
		}

		file := os.NewFile(uintptr(fd), "Socket")
		conn, err := net.FileConn(file)
		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	if strings.HasPrefix(network, "udp") {

		var addr *net.UDPAddr
		var err error

		addr, haveaddr := sDialer.vp.prepareddomain.udpprepared[Address]

		if haveaddr == false {
			log.Println("Not Using Prepared: UDP,", Address)
			addr, err = net.ResolveUDPAddr(network, Address)
		} else {
			log.Println("Using Prepared: UDP,", Address)
		}

		if err != nil {
			return nil, err
		}
		fd, err := unix.Socket(unix.AF_INET6, unix.SOCK_DGRAM, unix.IPPROTO_UDP)
		if err != nil {
			return nil, err
		}

		//Protect socket fd!
		//log.Println("Protecting Sock:", fd)
		sDialer.vp.VpnSupportSet.Protect(fd)

		sa := new(unix.SockaddrInet6)
		sa.Port = addr.Port
		sa.ZoneId = uint32(zoneToInt(addr.Zone))
		//fmt.Println(addr.IP.To16())
		copy(sa.Addr[:], addr.IP.To16())
		//fmt.Println(sa.Addr)
		err = unix.Connect(fd, sa)
		if err != nil {
			return nil, err
		}

		file := os.NewFile(uintptr(fd), "Socket")
		conn, err := net.FileConn(file)
		if err != nil {
			return nil, err
		}

		return conn, nil

	}
	return nil, errors.New("Pto udf")
}

// Bigger than we need, not too big to worry about overflow
const big = 0xFFFFFF

// Decimal to integer starting at &s[i0].
// Returns number, new offset, success.
func dtoi(s string, i0 int) (n int, i int, ok bool) {
	n = 0
	for i = i0; i < len(s) && '0' <= s[i] && s[i] <= '9'; i++ {
		n = n*10 + int(s[i]-'0')
		if n >= big {
			return 0, i, false
		}
	}
	if i == i0 {
		return 0, i, false
	}
	return n, i, true
}

func zoneToInt(zone string) int {
	if zone == "" {
		return 0
	}
	if ifi, err := net.InterfaceByName(zone); err == nil {
		return ifi.Index
	}
	n, _, _ := dtoi(zone, 0)
	return n
}
