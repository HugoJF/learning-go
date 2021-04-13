package protocol

import (
	"bytes"
	"fmt"
	"net"
	"time"
	"valve-master-server/utils"
	"valve-master-server/valve"
)

type Query struct {
	RateLimitInterval time.Duration
	RetryDuration     time.Duration
	TimeoutDuration   time.Duration
	Servers           []string
	region            uint8
	done              chan bool
	seeds             chan string
}


func NewQuery(region string) (*Query, error) {
	code, ok := valve.Regions[region]

	if ok == false {
		return nil, fmt.Errorf(fmt.Sprintf("'%s' is not a valid region", region))
	}

	return &Query{
		RateLimitInterval: 1 * time.Second,
		RetryDuration:     5 * time.Second,
		TimeoutDuration:   30 * time.Second,
		Servers:           []string{},
		region:            code,
		seeds:             make(chan string),
		done:              make(chan bool),
	}, nil
}

//sendPacket send master server queries
func (q *Query) sendPacket(conn *net.UDPConn) {
	retry := time.NewTicker(q.RetryDuration)
	timeout := time.NewTicker(q.TimeoutDuration)
	ratelimit := time.Tick(q.RateLimitInterval)

	var seed string

	for {
		select {
		// Wait for new seed, retry or timeout
		case newSeed := <-q.seeds:
			seed = newSeed
			timeout.Reset(q.TimeoutDuration)
			break
		case <-retry.C:
			retry.Reset(q.RetryDuration)
			break
		case <-timeout.C:
			q.done <- true
			return
		}

		packetParts := [][]byte{
			valve.Header,
			{q.region},
			utils.ToZeroString(seed),
			utils.ToZeroString(""),
		}

		// Build packet
		packet := bytes.Join(packetParts, nil)

		_, err := conn.Write(packet)
		if err != nil {
			fmt.Printf("Couldn't send response %v\n", err)
			continue
		}

		<-ratelimit
	}
}

//receivePacket receives data from UDP connection
func (q *Query) receivePacket(conn *net.UDPConn) {
	buffer := make([]byte, 2048)

	for {
		_, _, err := conn.ReadFromUDP(buffer)

		if err != nil {
			fmt.Printf("Some error  %v", err)
			continue
		}

		// Check if response has correct header
		if bytes.Compare(valve.Response, buffer[0:6]) != 0 {
			fmt.Printf("Invalid response: %v", err)
			continue
		}

		// Parse buffer after header signature
		ips, last := utils.ParseIps(buffer[6:])

		// Add new IPs to list
		q.Servers = append(q.Servers, ips...)

		// Send last IP as seed for future requests
		q.seeds <- last
	}
}

func (q *Query) Start() ([]string, error) {
	ip, err := net.ResolveIPAddr("ip4", "hl2master.steampowered.com")

	if err != nil {
		return nil, err
	}

	addr := net.UDPAddr{
		Port: 27011,
		IP:   ip.IP,
	}

	conn, err := net.DialUDP("udp", nil, &addr)

	if err != nil {
		return nil, err
	}

	go q.sendPacket(conn)
	go q.receivePacket(conn)

	// First seed will always be 0.0.0.0:0
	q.seeds <- utils.ZeroIp

	// Wait for routines to end
	<-q.done

	return q.Servers, nil
}
