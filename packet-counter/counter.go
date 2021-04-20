// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// arpscan implements ARP scanning of all interfaces' local networks using
// gopacket and its subpackages.  This example shows, among other things:
//   * Generating and sending packet data
//   * Reading in packet data and interpreting it
//   * Use of the 'pcap' subpackage for reading/writing

// https://github.com/google/gopacket/blob/master/examples/arpscan/arpscan.go
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/hugojf/dumpcap"
	"github.com/manifoldco/promptui"
	"log"
	"net"
	"time"
)

var (
	ctx = context.Background()
)

func main() {
	fmt.Println(dumpcap.VersionString())

	devices, err := dumpcap.Devices(false)
	if err != nil {
		panic(err)
	}

	log.Printf("Found %d device\n", len(devices))

	var ifacesnames []string
	for _, dev := range devices {
		if dev.FriendlyName != "" {
			ifacesnames = append(ifacesnames, dev.FriendlyName)
		} else {
			ifacesnames = append(ifacesnames, dev.Name)
		}
	}

	prompt := promptui.Select{
		Label: "Select interface to sniff",
		Items: ifacesnames,
	}

	_, ifacename, err := prompt.Run()

	if err != nil {
		panic(err)
	}

	log.Printf("Running on interface: %v", ifacename)

	var deviceid string
	for _, dev := range devices {
		if dev.Name == ifacename {
			deviceid = ifacename
			break
		} else if dev.FriendlyName == ifacename {
			deviceid = dev.Name
			break
		}
	}

	if deviceid == "" {
		log.Printf("Failed to relocate device ID")
		return
	}

	done := make(chan bool)

	go func() {
		iface, err := net.InterfaceByName(ifacename)

		if err != nil {
			panic(err)
		}

		if err := scan(iface, deviceid); err != nil {
			log.Printf("interface %v: %v", iface.Name, err)
		}

		done <- true
	}()

	<-done
}

// scan scans an individual interface's local network for machines using ARP requests/replies.
//
// scan loops forever, sending packets out regularly.  It returns an error if
// it's ever unable to write a packet.
func scan(iface *net.Interface, deviceid string) error {
	// We just look for IPv4 addresses, so try to find if the interface has one.
	var addr *net.IPNet
	addresses, err := iface.Addrs()
	if err != nil {
		return err
	}

	for _, a := range addresses {
		ipnet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}

		ip4 := ipnet.IP.To4()
		if ip4 == nil {
			continue
		}

		addr = &net.IPNet{
			IP:   ip4,
			Mask: ipnet.Mask[len(ipnet.Mask)-4:],
		}
		break
	}

	// Sanity-check that the interface has a good address.
	if addr == nil {
		return errors.New("no good IP network found")
	}
	log.Printf("Using network range %v for interface %v", addr, iface.Name)

	// Open up a pcap handle for packet reads/writes.
	handle, err := pcap.OpenLive(deviceid, 65536, true, pcap.BlockForever)
	if err != nil {
		return err
	}
	defer handle.Close()

	read(handle)

	return nil
}

// read watches a handle for incoming ARP responses we might care about, and prints them.
//
// read loops until 'stop' is closed.
func read(handle *pcap.Handle) {
	src := gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)
	in := src.Packets()

	red := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	for {
		packet := <-in
		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		if ipLayer == nil {
			continue
		}
		ipv4 := ipLayer.(*layers.IPv4)
		ip := ipv4.SrcIP.String()

		_, err := red.Get(ctx, ip).Result()

		// Assert count is set to something
		switch {
		case err == redis.Nil:
			red.Set(ctx, ip, 0, 30 * time.Second)
		case err != nil:
			continue
		}

		// Increment IP count
		val, err := red.Incr(ctx, ip).Result()

		if err != nil {
			continue
		}

		// Get all IPs tracked
		keys, err := red.Keys(ctx, "*").Result()

		if err != nil {
			continue
		}

		// Note:  we might get some packets here that aren't responses to ones we've sent,
		// if for example someone else sends US an ARP request.  Doesn't much matter, though...
		// all information is good information :)
		log.Printf("Received %v packets with source %v. Tracking %v IPs", val, ipv4.SrcIP, len(keys))
	}
}
