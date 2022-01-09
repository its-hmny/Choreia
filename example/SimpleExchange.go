package main

import (
	"fmt"
	"time"
)

func responder(channel chan int) {
	channel <- 0
}

func main() {
	// Creates the channels
	chanA, chanB := make(chan int), make(chan int)

	// Starts the worker processes
	go responder(chanA)
	go responder(chanB)

	// Little delay
	time.Sleep(time.Second * 1)

	// Select from both channels
	select {
	case <-chanA:
		fmt.Println("Received from channel A")
	case <-chanB:
		fmt.Println("Received from channel B")
	default:
		fmt.Println("Both A and B doesn't have messages")
	}
}
