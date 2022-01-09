package main

import (
	"fmt"
	"math/rand"
	"time"
)

type payload struct {
	data      int
	timestamp int64
}

func worker(incoming chan int, outgoing chan payload) {
	for msg := range incoming {
		// Sends back the results on the out channel
		outgoing <- payload{msg + 1, time.Now().Unix()}
	}
}

func main() {
	// Creates the channels
	in, out := make(chan int, 10), make(chan payload, 10)

	// Starts the worker processes
	go worker(in, out)
	go worker(in, out)

	// Infinite loop
	for {
		in <- rand.Int()
		res := <-out
		fmt.Printf("Received %d at %d \n", res.data, res.timestamp)
	}
}
