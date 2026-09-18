package main

import (
	"fmt"
	"time"
)

type Context struct{
	Symbol string
	Channel chan string
}

func worker(id int, ctx Context) {
	for msg := range ctx.Channel {
		fmt.Printf("Worker %d received: %s\n", id, msg)
	}
	fmt.Printf("Worker %d exiting\n", id)
}

func main() {
	workerChannels := make(map[string]Context)

	// Create channels and start goroutines
	for i := 0; i < 5; i++ {
		ctx := Context{
			Symbol : fmt.Sprintf("%d", i),
			Channel : make(chan string),
		}
		workerChannels[ctx.Symbol] = ctx
		go worker(i, ctx)
	}

	// Send messages to each worker
	for i, worker := range(workerChannels){
		worker.Channel <- fmt.Sprintf("Hello from main to worker %s", i)
	}

	// Add and Delete go routine
	newWorker := make(map[string]Context)
	for i:= 3; i < 8; i++{
		symbol := fmt.Sprintf("%d", i)
		item, exist := workerChannels[symbol]
		if exist {
			fmt.Printf("%s는 있습니다. [%s]\n", symbol, item.Symbol)

			newWorker[symbol] = item
			delete(workerChannels, symbol)
		}else{
			ctx := Context{
				Symbol : symbol,
				Channel : make(chan string),
			}
			newWorker[symbol]= ctx
			go worker(i, ctx)
		}
	}

	// Clean worker
	for _, worker := range(workerChannels){
		close(worker.Channel)
	}
	workerChannels = newWorker

	// Close channels after some time to stop workers
	time.Sleep(2 * time.Second)
	for _, worker := range(workerChannels){
		fmt.Printf("Close [%s]\n", worker.Symbol)
		close(worker.Channel)
	}
	time.Sleep(1 * time.Second) // Allow goroutines to finish
	fmt.Println("All workers stopped")
}

