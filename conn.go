package lowsocket

import (
	"errors"
	"net"
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

type Conn struct {
	intf  *net.Interface
	proto uint16

	f    *os.File
	conn syscall.RawConn
}

const (
	ETH_P_ALL = unix.ETH_P_ALL
	ETH_P_IP  = unix.ETH_P_IP
	ETH_P_ARP = unix.ETH_P_ARP
)

var (
	ErrNotSupported = errors.New("not supported")
	// BroadcastHardwareAddress ff:ff:ff:ff:ff:ff
	BroadcastHardwareAddress = net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
)

func htons(data uint16) uint16 { return data<<8 | data>>8 }

func NewConn(ifname string, typ int, proto uint16) (*Conn, error) {

	switch typ {
	case unix.SOCK_RAW, unix.SOCK_DGRAM:
	default:
		return nil, ErrNotSupported
	}

	intf, err := net.InterfaceByName(ifname)
	if err != nil {
		return nil, err
	}

	fd, err := unix.Socket(unix.AF_PACKET, typ, 0)
	if err != nil {
		return nil, err
	}

	sa := &unix.SockaddrLinklayer{
		Protocol: htons(proto),
		Ifindex:  intf.Index,
	}
	if err := unix.Bind(fd, sa); err != nil {
		unix.Close(fd)
		return nil, err
	}

	// if err := unix.SetsockoptInt(fd, unix.SOL_PACKET, unix.PACKET_AUXDATA, 1); err != nil {
	// 	unix.Close(fd)
	// 	return nil, err
	// }

	if err := unix.SetNonblock(fd, true); err != nil {
		unix.Close(fd)
		return nil, err
	}

	f := os.NewFile(uintptr(fd), "")
	conn, err := f.SyscallConn()
	if err != nil {
		f.Close()
		return nil, err
	}

	c := &Conn{
		intf:  intf,
		proto: proto,
		f:     f,
		conn:  conn,
	}
	return c, nil
}

func (c *Conn) Close() error {
	return c.f.Close()
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.f.SetDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.f.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.f.SetWriteDeadline(t)
}

func (c *Conn) SetFilter(filter []unix.SockFilter) error {
	var err error
	cerr := c.conn.Control(func(fd uintptr) {
		if len(filter) == 0 {
			err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_DETACH_FILTER, 0)
		} else {
			err = unix.SetsockoptSockFprog(int(fd), unix.SOL_SOCKET, unix.SO_ATTACH_FILTER, &unix.SockFprog{
				Len:    uint16(len(filter)),
				Filter: &filter[0],
			})
		}
	})

	if err != nil {
		return err
	}
	return cerr
}

func (c *Conn) Read(p []byte) (n int, err error) {
	cerr := c.conn.Read(func(fd uintptr) bool {
		n, err = unix.Read(int(fd), p)
		// When the socket is in non-blocking mode, we might see EAGAIN
		// and end up here. In that case, return false to let the
		// poller wait for readiness. See the source code for
		// internal/poll.FD.RawRead for more details.
		//
		// If the socket is in blocking mode, EAGAIN should never occur.
		return err != unix.EAGAIN
	})

	if err != nil {
		return
	}
	err = cerr
	return
}

func (c *Conn) Write(p []byte) (n int, err error) {
	cerr := c.conn.Write(func(fd uintptr) bool {
		n, err = unix.Write(int(fd), p)
		return err != unix.EAGAIN
	})
	if err != nil {
		return
	}
	err = cerr
	return
}

func (c *Conn) Recvfrom(p []byte) (n int, addr net.HardwareAddr, err error) {
	var from unix.Sockaddr
	cerr := c.conn.Read(func(fd uintptr) bool {
		n, from, err = unix.Recvfrom(int(fd), p, 0)
		// When the socket is in non-blocking mode, we might see EAGAIN
		// and end up here. In that case, return false to let the
		// poller wait for readiness. See the source code for
		// internal/poll.FD.RawRead for more details.
		//
		// If the socket is in blocking mode, EAGAIN should never occur.
		return err != unix.EAGAIN
	})
	if from != nil {
		if linkAddr, ok := from.(*unix.SockaddrLinklayer); ok {
			addr = make(net.HardwareAddr, 6)
			copy(addr, linkAddr.Addr[:6])
		}
	}
	if err != nil {
		return
	}
	err = cerr
	return
}

func (c *Conn) Sendto(p []byte, to net.HardwareAddr) (err error) {
	linkAddr := &unix.SockaddrLinklayer{
		Ifindex: c.intf.Index,
		Halen:   uint8(len(to)),
	}
	if linkAddr.Halen > 0 {
		copy(linkAddr.Addr[:], to)
	}

	cerr := c.conn.Read(func(fd uintptr) bool {
		err = unix.Sendto(int(fd), p, 0, linkAddr)
		return err != unix.EAGAIN
	})
	if err != nil {
		return
	}
	err = cerr
	return
}
