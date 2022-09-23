package main

import "fmt"

func sayHello(c chan string) {
	c <- "Hello world!"
}

func main() {
	channel := make(chan string, 1)
	go sayHello(channel)
	fmt.Println(<-channel)
}
