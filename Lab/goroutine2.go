package main

import (
	"fmt"
	"time"
)

func doWork(id int, result chan<- error) {
	defer close(result) // Ensure the channel is closed when done
	// Simulate some work
	time.Sleep(1 * time.Second)
	if id%2 == 0 {
		result <- fmt.Errorf("goroutine %d failed", id)
	} else {
		result <- nil // Success
	}
}

func main() {
	for i := 1; i <= 3; i++ {
		resultChan := make(chan error)
		go doWork(i, resultChan)

		// Wait for result
		if err := <-resultChan; err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("goroutine %d succeeded\n", i)
		}
	}
}
