package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	fmt.Println("PogU")

	var wg sync.WaitGroup
	origin := make(chan string)
	for i := 1; i < 10; i++ {
		wg.Add(1)
		go loop(i, origin, &wg)
	}

	for {
		origin <- "heyo"
		time.Sleep(1 * time.Second)
	}

	wg.Wait()
}

func loop(id int, source chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	count := 1

	for {
		message := <- source
		fmt.Printf("[ID %d] Received: %s\n", id, message)
		count++
	}
}
