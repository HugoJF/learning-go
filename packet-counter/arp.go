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
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/lukaslueg/dumpcap"
	"github.com/manifoldco/promptui"
	"log"
	"net"
	"sync"
)

func main() {
	fmt.Println(dumpcap.VersionString())

	devices, err := dumpcap.Devices(false)
	if err != nil {
		panic(err)
	}

	var ifacesnames []string
	for _, dev := range devices {
		ifacesnames = append(ifacesnames, dev.FriendlyName)
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
		if dev.FriendlyName == ifacename {
			deviceid = dev.Name
			break
		}
	}

	if deviceid == "" {
		log.Printf("Failed to relocate device ID")
		return
	}

	// Get a list of all interfaces.
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for _, iface := range ifaces {
		wg.Add(1)
		// Start up a scan on each interface.
		go func(iface net.Interface) {
			defer wg.Done()
			if iface.Name != ifacename {
				return
			}
			if err := scan(&iface, deviceid); err != nil {
				log.Printf("interface %v: %v", iface.Name, err)
			}
		}(iface)
	}
	// Wait for all interfaces' scans to complete.  They'll try to run
	// forever, but will stop on an error, so if we get past this Wait
	// it means all attempts to write have failed.
	wg.Wait()
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
	counter := make(map[string]int)
	for {
		packet := <-in
		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		if ipLayer == nil {
			continue
		}
		ipv4 := ipLayer.(*layers.IPv4)

		counter[ipv4.SrcIP.String()]++
		// Note:  we might get some packets here that aren't responses to ones we've sent,
		// if for example someone else sends US an ARP request.  Doesn't much matter, though...
		// all information is good information :)
		log.Printf("Received %v packets from %v", counter[ipv4.SrcIP.String()], ipv4.SrcIP)
	}
}
