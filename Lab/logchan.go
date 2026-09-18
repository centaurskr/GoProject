package main

import (
    "fmt"
    "log"
    //"sync"
    "time"
)

func logger(logChan <-chan string) {
    for msg := range logChan {
        log.Println(msg) // Sequential logging
    }
}

func worker(id int, logChan chan<- string) {
    for i := 0; i < 5; i++ {
        //logChan <- fmt.Sprintf("Worker %d - Log %d", id, i)
        fmt.Printf("Worker %d - Log %d\n", id, i)
        time.Sleep(100 * time.Millisecond)
    }
}

func main() {
    logChan := make(chan string, 1000)

    // Start the logger goroutine
    go logger(logChan)

    // Start workers
	for{
    	for	 i := 0; i < 5; i++ {
        	go worker(i, logChan)
    	}
		logChan <- fmt.Sprintf("Main Sleep-----------\n")
		time.Sleep(10 * time.Second)
	}
    close(logChan)
}

