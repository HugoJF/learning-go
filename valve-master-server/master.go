package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

var (
	header   = []byte{0x31}
	response = []byte{0xff, 0xff, 0xff, 0xff, 0x66, 0x0a}
	seeds    = make(chan string)
	ips      = make(chan string)
)

const (
	ZeroIp string = "0.0.0.0:0"
	// regions
	UsEastCoast  byte = 0x00
	UsWestCoast  byte = 0x01
	SouthAmerica byte = 0x02
	Europe       byte = 0x03
	Asia         byte = 0x04
	Australia    byte = 0x05
	MiddleEast   byte = 0x06
	Africa       byte = 0x07
	RestOfWorld  byte = 0xff
)

//printIps receives any parsed IP and adds to main list
func printIps() {
	var all []string
	for {
		ip := <-ips
		all = append(all, ip)
		fmt.Printf("Received %d: %s\n", len(all), ip)
	}
}

func ToZeroString(s string) []byte {
	return append([]byte(s), 0x00)
}

//sendQuery send master server queries
func sendQuery(conn *net.UDPConn) {
	ratelimit := time.Tick(time.Second)
	for {
		seed := <-seeds

		packetParts := [][]byte{
			header,
			{SouthAmerica},
			ToZeroString(seed),
			ToZeroString("\\dedicated\\1\\linux\\1\\empty\\1\\password\\0\\appid\\730"),
		}

		var packet []byte
		for _, part := range packetParts {
			packet = append(packet, part...)
		}

		_, err := conn.Write(packet)
		if err != nil {
			fmt.Printf("Couldn't send response %v\n", err)
		}

		<- ratelimit
	}
}

//receiveData receives data from UDP connection
func receiveData(conn *net.UDPConn) {
	buffer := make([]byte, 2048)

	for {
		_, _, err := conn.ReadFromUDP(buffer)

		if err != nil {
			fmt.Printf("Some error  %v", err)
			continue
		}

		// Check if response has correct header
		if bytes.Compare(response, buffer[0:6]) != 0 {
			fmt.Printf("Invalid response: %v", err)
			continue
		}

		var last string
		for i := 6; ; i += 6 {
			ip := sliceToIp(buffer[i : i+6])

			if ip == ZeroIp {
				break
			}
			ips <- ip
			last = ip
		}
		seeds <- last
	}
}

//sliceToIp transform byte slice to string IP
func sliceToIp(data []byte) string {
	ipBytes := data[0:4]
	portBytes := data[4:6]

	ip := net.IP(ipBytes).String()
	port := binary.BigEndian.Uint16(portBytes)

	return fmt.Sprintf("%s:%d", ip, port)
}

//main kickstart everything
func main() {
	ip, err := net.ResolveIPAddr("ip4", "hl2master.steampowered.com")

	if err != nil {
		log.Panic(err)
	}

	addr := net.UDPAddr{
		Port: 27011,
		IP:   ip.IP, // FIXME: resolve from hostname
	}

	fmt.Println("Dialing connection")
	conn, err := net.DialUDP("udp", nil, &addr)

	if err != nil {
		panic(err)
	}

	fmt.Println("Entering loop")

	go sendQuery(conn)
	go receiveData(conn)
	go printIps()

	seeds <- ZeroIp

	// Wait forever
	// TODO: this should become a timeout
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
