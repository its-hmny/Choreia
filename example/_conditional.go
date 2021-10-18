package main

var channel chan int

func dummy() {}

func main() {
	channel = make(chan int)

	if <-channel; true {
		dummy()
	} else if false {
		csc := dummy()
	} else {
		<-channel
	}

	errorChan <- 1

	switch recv := <-channel; <-errorChan {
	case 0:
		dummy()
	case 1:
		<-channel
	default:
		channel <- 3
	}

	errorChan <- 1
	<-errorChan

	switch recv := <-channel; recv.(type) {
	case int:
		dummy()
	case string:
		<-channel
	default:
		channel <- 3
	}

	<-errorChan
}
