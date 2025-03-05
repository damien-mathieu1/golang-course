package main

import (
	"fmt"
	"sync"
)

func goroutine(id int, ch chan string, wg *sync.WaitGroup, K int) {
	defer wg.Done()
	for i := 0; i < K; i++ {
		msg := <-ch
		fmt.Printf("Goroutine %d received: %s\n", id, msg)
		msg = fmt.Sprintf("Message from %d", id)
		ch <- msg
	}
}

func main() {
	const N = 5
	const K = 3

	var wg sync.WaitGroup
	ch := make(chan string, 1)

	for i := 0; i < N; i++ {
		wg.Add(1)
		go goroutine(i, ch, &wg, K)
	}

	ch <- "Initial message"

	wg.Wait()
}
