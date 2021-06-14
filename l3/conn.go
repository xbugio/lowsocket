package l3

import (
	"net"

	"github.com/xbugio/lowsocket"
	"golang.org/x/sys/unix"
)

type Conn struct {
	*lowsocket.Conn
}

const (
	ETH_P_ALL = lowsocket.ETH_P_ALL
	ETH_P_IP  = lowsocket.ETH_P_IP
	ETH_P_ARP = lowsocket.ETH_P_ARP
)

var (
	BroadcastHardwareAddress = lowsocket.BroadcastHardwareAddress
)

func NewConn(ifname string, proto uint16) (*Conn, error) {
	lc, err := lowsocket.NewConn(ifname, unix.SOCK_DGRAM, proto)
	if err != nil {
		return nil, err
	}
	c := &Conn{lc}
	return c, nil
}

func (c *Conn) Write(p []byte, to net.HardwareAddr) error {
	return c.Conn.Sendto(p, to)
}
