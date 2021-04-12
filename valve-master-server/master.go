package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"time"
	"valve-master-server/valve"
)

type Query struct {
	TimeoutDuration time.Duration
	GiveUpDuration  time.Duration
	Ips             []string
	region          uint8
	done            chan bool
	seeds           chan string
}

const (
	ZeroIp string = "0.0.0.0:0"
)

func NewQuery(region string) (*Query, error) {
	code, ok := valve.Regions[region]
	
	if ok == false {
		return nil, errors.New(fmt.Sprintf("'%s' is not a valid region", region))
	}

	return &Query{
		region:          code,
		TimeoutDuration: 10 * time.Second,
		GiveUpDuration:  30 * time.Second,
		Ips:             []string{},
		seeds:           make(chan string),
		done:            make(chan bool),
	}, nil
}

//sendPacket send master server queries
func (q *Query) sendPacket(conn *net.UDPConn) {
	timeout := time.NewTicker(q.TimeoutDuration)
	giveup := time.NewTicker(q.GiveUpDuration)
	ratelimit := time.Tick(time.Second)

	var seed string

	for {
		select {
		case s := <-q.seeds:
			seed = s
			giveup.Reset(q.GiveUpDuration)
			break
		case <-timeout.C:
			timeout.Reset(q.TimeoutDuration)
			break
		case <-giveup.C:
			q.done <- true
			return
		}

		fmt.Printf("Requesting with seed %s\n", seed)
		packetParts := [][]byte{
			valve.Header,
			{q.region},
			toZeroString(seed),
			toZeroString("\\dedicated\\1\\linux\\1\\empty\\1\\password\\0\\appid\\730"),
		}

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

		var last string
		for i := 6; ; i += 6 {
			ip := sliceToIp(buffer[i : i+6])

			if ip == ZeroIp {
				break
			}

			last = ip
			fmt.Printf("Received %v\n", ip)
			q.Ips = append(q.Ips, ip)
		}
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

	fmt.Println("Dialing connection")
	conn, err := net.DialUDP("udp", nil, &addr)

	if err != nil {
		return nil, err
	}

	fmt.Println("Entering loop")

	go q.sendPacket(conn)
	go q.receivePacket(conn)

	q.seeds <- ZeroIp

	<-q.done

	return q.Ips, nil
}
