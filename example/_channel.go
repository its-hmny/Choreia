package main

import "fmt"

/*
// Unbuffered, synchronous comunication
ch := make(chan int)
// Buffered, asynchronous comunication till the buffer is full
ch := make(chan int, 100)

ch <- v    // Send v to channel ch.
v := <-ch  // Receive from ch, and assign value to v.
*/

var global = make(chan int, 1)

func test(callback func(int) int, channel chan int) int {
	return (callback(<-channel))
}

func void() int {
	// Does nothing
	return 1
}

func sum(s []int, c chan int) int {
	accumulator := 0
	for _, v := range s {
		accumulator += v
	}
	// Send accumulator to c
	c <- accumulator
	return 0
}

func main() {
	s := []int{7, 2, 8, -9, 4, 0}
	list := make([]int, 2)

	var channel = make(chan int)
	boundedChan := make(chan string, 10)

	go sum(s, channel)
	go sum(s[len(s)/2:], channel)

	go func(x int) {
		fmt.Println("Hello from anonymous function")
	}(3)

	if <-channel; true {
		void()
		csc := void()
	} else if false {
		void()
		csc := void()
	} else {
		void()
		csc := void()
	}

	boundedChan <- "Hello"
	<-boundedChan
	x, y := <-channel, <-channel // Receive from c

	close(channel)

	fmt.Println(x, y, x+y, csc)
}
