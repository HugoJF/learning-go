package utils

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	ZeroIp string = "0.0.0.0:0"
)

func ParseIps(data []byte) ([]string, string) {
	var last string
	var servers []string

	for i := 0; ; i += 6 {
		ip := SliceToIp(data[i : i+6])

		if ip == ZeroIp {
			break
		}

		last = ip
		servers = append(servers, ip)
	}

	return servers, last
}

func ToZeroString(s string) []byte {
	return append([]byte(s), 0x00)
}

//SliceToIp transform byte slice to string IP
func SliceToIp(data []byte) string {
	ipBytes := data[0:4]
	portBytes := data[4:6]

	ip := net.IP(ipBytes).String()
	port := binary.BigEndian.Uint16(portBytes)

	return fmt.Sprintf("%s:%d", ip, port)
}
