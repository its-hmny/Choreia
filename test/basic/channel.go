package main

import "fmt"

/*
ch := make(chan int) // Unbuffered, synchronous comunication
ch := make(chan int, 100) // Buffered, asynchronous comunication till the buffer is full

ch <- v    // Send v to channel ch.
v := <-ch  // Receive from ch, and assign value to v.
*/

func sum(s []int, c chan int) {
	sum := 0
	for _, v := range s {
		sum += v
	}
	c <- sum // send sum to c
}

func main() {
	s := []int{7, 2, 8, -9, 4, 0}

	c := make(chan int)
	go sum(s[:len(s)/2], c)
	go sum(s[len(s)/2:], c)
	x, y := <-c, <-c // receive from c

	fmt.Println(x, y, x+y)
}
