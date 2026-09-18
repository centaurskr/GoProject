package main

import (
	"log"
	"time"

	zmq "github.com/pebbe/zmq4"
)

func main() {
	publisher, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		log.Fatal(err)
	}
	defer publisher.Close()

	// Connect to the XSUB side of the proxy
	err = publisher.Connect("tcp://127.0.0.1:5555")
	if err != nil {
		log.Fatal(err)
	}

	for {
		// Send a message to subscribers
		_, err := publisher.Send("Hello", 0)
		if err != nil {
			log.Println("Error sending message:", err)
		}
		time.Sleep(1 * time.Second)
	}
}

