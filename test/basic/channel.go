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

func sum(s []int, c chan int) {
	accumulator := 0
	for _, v := range s {
		accumulator += v
	}
	// Send accumulator to c
	c <- accumulator
}

func main() {
	s := []int{7, 2, 8, -9, 4, 0}

	c := make(chan int)
	d := make(chan string)

	go sum(s[:len(s)/2], c)
	go sum(s[len(s)/2:], c)
	x, y := <-c, <-c // Receive from c

	close(c)

	fmt.Println(x, y, x+y)
}
