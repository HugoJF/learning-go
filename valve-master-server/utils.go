package main

import (
	"encoding/binary"
	"fmt"
	"net"
)


func toZeroString(s string) []byte {
	return append([]byte(s), 0x00)
}


//sliceToIp transform byte slice to string IP
func sliceToIp(data []byte) string {
	ipBytes := data[0:4]
	portBytes := data[4:6]

	ip := net.IP(ipBytes).String()
	port := binary.BigEndian.Uint16(portBytes)

	return fmt.Sprintf("%s:%d", ip, port)
}
