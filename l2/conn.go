package l2

import (
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

func NewConn(ifname string, proto uint16) (*Conn, error) {
	lc, err := lowsocket.NewConn(ifname, unix.SOCK_RAW, proto)
	if err != nil {
		return nil, err
	}
	c := &Conn{lc}
	return c, nil
}

func (c *Conn) Sendto(p []byte) (err error) {
	return c.Conn.Sendto(p, nil)
}
