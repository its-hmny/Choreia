package main

import (
	"fmt"
	"math/rand"
	"time"
)

func sender(c chan string, msg string) {
	// Sends a message then waits some milliseconds
	for i := 0; ; i++ {
		c <- fmt.Sprintf("%s %d", msg, i)
		time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
	}
}

func main() {
	// Creates a shared channel and starts a Goroutine
	channel := make(chan string)
	go sender(channel, "Iteration n.")

	// Waits for the incoming messages
	for i := 0; i < 5; i++ {
		data := <-channel
		fmt.Printf("'nested' says: %q\n", data)
	}

	fmt.Println("Task completed")
}
