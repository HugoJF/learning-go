package main

import (
	"fmt"
	"time"
)

func main() {
	tick := time.Tick(5 * time.Second)
	fucku := make(chan bool)

	go func() {
		for {
			fmt.Printf("Waiting for ticks %v\n", time.Now())
			<-tick
			fmt.Printf("Ticked %v\n", time.Now())
			time.Sleep(6 * time.Second)
			fmt.Printf("Slept %v\n", time.Now())
		}
	}()

	<-fucku
}
