package main

import "fmt"

func dummy(channel chan string) {
	channel <- "Hello from dummy (1)" //Sends a message on the shared channel
	channel <- "Hello from dummy (2)" //Sends a message on the shared channel
}

func f(channel chan string) {
	go dummy(channel) // Spawns a new "dummy" Goroutine
	msg := <-channel  // Receives the message sent by itself
	fmt.Println(msg)
}

func main() {
	// Creates the shared channel
	channel := make(chan string)
	// Call the "f" function
	f(channel)
	// Receives something from "channel"
	msg := <-channel
	fmt.Println(msg)
}
