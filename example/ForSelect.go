package main

import (
	"fmt"
	"math/rand"
)

func worker(channel chan int) {
	for {
		channel <- rand.Int()
	}
}

func main() {
	// Creates the channels
	chanA, chanB := make(chan int, 10), make(chan int, 10)

	// Starts the worker processes
	go worker(chanA)
	go worker(chanB)

	for { // Receives from both channels indefinitely
		select {
		case <-chanA:
			fmt.Println("Received from channel A")
		case <-chanB:
			fmt.Println("Received from channel B")
		default:
			fmt.Println("No message available")
		}
	}
}
