package main

import (
	"fmt"
	"math/rand"
	"time"
)

func ouvrier(id int, numbers <-chan int, results chan<- int) {
	for num := range numbers {
		fmt.Printf("Ouvrier %d a reçu le nombre %d\n", id, num)
		results <- num * num
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	const M = 5
	numbers := make(chan int, M)
	results := make(chan int, M)

	for i := 1; i <= M; i++ {
		go ouvrier(i, numbers, results)
	}

	for i := 0; i < M; i++ {
		num := rand.Intn(100)
		fmt.Printf("Maître envoie le nombre %d\n", num)
		numbers <- num
	}
	close(numbers)

	for i := 0; i < M; i++ {
		result := <-results
		fmt.Printf("Maître a reçu le résultat %d\n", result)
	}
}
