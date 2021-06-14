package ip

import (
	"encoding/binary"

	"golang.org/x/net/ipv4"
)

type IPV4 struct {
	ipv4.Header
	Payload []byte
}

func (p *IPV4) Marshal() ([]byte, error) {
	p.Header.Version = ipv4.Version
	p.Header.Len = ipv4.HeaderLen + len(p.Options)
	p.Header.TotalLen = p.Len + len(p.Payload)
	p.Header.Checksum = 0
	headerData, err := p.Header.Marshal()
	if err != nil {
		return nil, err
	}
	p.Header.Checksum = checksum(headerData)
	binary.BigEndian.PutUint16(headerData[10:12], uint16(p.Header.Checksum))

	data := make([]byte, len(headerData)+len(p.Payload))
	copy(data, headerData)
	copy(data[len(headerData):], p.Payload)
	return data, nil
}

func checksum(msg []byte) int {
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

	return ^sum
}
