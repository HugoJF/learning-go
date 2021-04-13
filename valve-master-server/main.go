package main

import (
	"fmt"
	"log"
	"valve-master-server/protocol"
)

func main() {
	q1, err1 := protocol.NewQuery("RestOfWorld")
	//q2, err2 := NewQuery("Australia")

	if err1 != nil {
		log.Fatal("Error creating query 1", err1)
	}

	all, err := q1.Start()

	if err != nil {
		log.Fatal("Failed to query servers", err)
	}

	fmt.Println(all)
}
