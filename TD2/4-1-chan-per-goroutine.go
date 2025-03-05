package main

import (
	"fmt"
	"sync"
)

func goroutine(id int, in chan string, out chan string, wg *sync.WaitGroup, K int) {
	defer wg.Done()
	for i := 0; i < K; i++ {
		msg := <-in
		fmt.Printf("Goroutine %d received: %s\n", id, msg)
		msg = fmt.Sprintf("Message from %d", id)
		out <- msg
	}
}

func main() {
	const N = 5
	const K = 3

	var wg sync.WaitGroup
	chs := make([]chan string, N)

	for i := 0; i < N; i++ {
		chs[i] = make(chan string, 1)
	}

	for i := 0; i < N; i++ {
		wg.Add(1)
		go goroutine(i, chs[i], chs[(i+1)%N], &wg, K)
	}

	chs[0] <- "Initial message"

	wg.Wait()
}
