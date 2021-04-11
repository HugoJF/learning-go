package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
)

var (
	header = []byte{0x31, 0xff}
	zeros  = []byte{0x00, 0x00} // TODO actually build filters (first zero is from IP)
	seeds  = make(chan string)
	ips    = make(chan string)
)

//printIps receives any parsed IP and adds to main list
func printIps() {
	var all []string
	for {
		ip := <- ips
		all = append(all, ip)
		fmt.Printf("Received %d: %s\n", len(all), ip)
	}
}

//sendQuery send master server queries
func sendQuery(conn *net.UDPConn) {
	for {
		seed := <-seeds
		fmt.Println("Seed: " + seed)

		parts := [][]byte{header, []byte(seed), zeros}

		var result []byte
		for _, p := range parts {
			result = append(result, p...)
		}

		// TODO rate limit
		_, err := conn.Write(result)
		if err != nil {
			fmt.Printf("Couldn't send response %v\n", err)
		}
	}
}

//receiveData receives data from UDP connection
func receiveData(conn *net.UDPConn) {
	p := make([]byte, 2048)

	for {
		_, _, err := conn.ReadFromUDP(p)
		if err != nil {
			fmt.Printf("Some error  %v", err)
			continue
		}
		var last string
		for i := 0; ; i += 6 {
			ip := sliceToIp(p[i : i+6])

			if ip == "0.0.0.0:0" {
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
	addr := net.UDPAddr{
		Port: 27011,
		IP:   net.ParseIP("208.64.200.52"), // FIXME: resolve from hostname
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

	seeds <- "0.0.0.0:0"

	// Wait forever
	// TODO: this should become a timeout
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
