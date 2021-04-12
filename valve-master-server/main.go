package main

import (
	"fmt"
	"log"
	"sync"
)

func main() {
	q1, err1 := NewQuery("RestOfWorld")
	q2, err2 := NewQuery("Australia")

	if err1 != nil {
		log.Fatal("Error creating query 1", err1)
	}
	if err2 != nil {
		log.Fatal("Error creating query 2", err2)
	}

	var all []string

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		ips, _ := q1.Start()
		all = append(all, ips...)
		wg.Done()
	}()

	go func() {
		ips, _ := q2.Start()
		all = append(all, ips...)
		wg.Done()
	}()
	wg.Wait()
	fmt.Println(all)
}
