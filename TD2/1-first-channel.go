package main

import (
	"fmt"
)

func main() {
	messageChannel := make(chan string)

	go func() {
		messageChannel <- "Bonjour, Goroutine !"
	}()

	message := <-messageChannel
	fmt.Println(message)
}
