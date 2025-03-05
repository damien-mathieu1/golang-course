package main

import (
	"fmt"
)

func main() {
	const N = 10
	channels := make([]chan int, N)

	for i := range channels {
		channels[i] = make(chan int)
	}

	for i := 0; i < N; i++ {
		go func(i int) {
			val := <-channels[i]
			val++
			if i+1 < N {
				channels[i+1] <- val
			} else {
				fmt.Println("Final value:", val)
			}
		}(i)
	}

	channels[0] <- 0

	fmt.Scanln()
}
