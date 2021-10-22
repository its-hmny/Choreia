package main

func main() {
	for i := make(chan int, 5); <-i == 5; i <- 5 {
		<-tick
	}

	<-boom

	for i := 0; i <= 5; i++ {
		tick <- 6
	}

	<-boom

	myChan := make(chan int, 5)
	for _, i := range myChan {
		<-tick
		for _, i := range list {
			<-tick
		}
	}

	<-boom
}
