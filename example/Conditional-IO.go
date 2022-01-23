package main

import (
	"fmt"
	"math/rand"
)

func getRandomNumber(q chan int) {
	q <- rand.Int()
}

func main() {
	A, B, C, D := make(chan int), make(chan int), make(chan int), make(chan int)

	go getRandomNumber(A)
	go getRandomNumber(B)
	go getRandomNumber(C)

	if true {
		resA, resB := <-A, <-B
		fmt.Printf("Received from A & B: %d %d\n", resA, resB)
	}

	if true {
		resB, resC := <-B, <-C
		fmt.Printf("Received from B & C: %d %d\n", resB, resC)
	}

	go getRandomNumber(D)
	<-D
}
