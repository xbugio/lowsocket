package udp

import (
	"encoding/binary"
	"net"
)

type UDP struct {
	SrcIP    net.IP
	DstIP    net.IP
	SrcPort  uint16
	DstPort  uint16
	Len      uint16
	Checksum uint16
	Payload  []byte
}

func (p *UDP) Marshal() []byte {

	p.Checksum = 0
	p.Len = uint16(8 + len(p.Payload))

	// 20 fake header + 8 udp header + payloadlen
	data := make([]byte, 12+8+len(p.Payload))

	// fake header
	copy(data[0:4], p.SrcIP.To4())
	copy(data[4:8], p.DstIP.To4())
	data[8], data[9] = 0x00, 0x11
	binary.BigEndian.PutUint16(data[10:12], p.Len)

	// udp header
	binary.BigEndian.PutUint16(data[12:14], p.SrcPort)
	binary.BigEndian.PutUint16(data[14:16], p.DstPort)
	binary.BigEndian.PutUint16(data[16:18], p.Len)
	binary.BigEndian.PutUint16(data[18:20], p.Checksum)
	copy(data[20:], p.Payload)

	p.Checksum = checksum(data)
	binary.BigEndian.PutUint16(data[18:20], p.Checksum)

	return data
}

func checksum(msg []byte) uint16 {
	sum := 0
	msgLen := len(msg)
	if msgLen%2 != 0 {
		return 0
	}

	for i := 0; i < msgLen; i += 2 {
		sum += int(binary.BigEndian.Uint16(msg[i : i+1]))
	}

	for sum > 0xffff {
		sum = (sum >> 16) + (sum & 0xffff)
	}

	return uint16(^sum)
}
