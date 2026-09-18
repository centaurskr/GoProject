package main

import (
	"log"

	zmq "github.com/pebbe/zmq4"
)

func main() {
	subscriber, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		log.Fatal(err)
	}
	defer subscriber.Close()

	// Connect to the XPUB side of the proxy
	err = subscriber.Connect("tcp://127.0.0.1:5557")
	if err != nil {
		log.Fatal(err)
	}

	// Subscribe to all topics
	// subscriber.SetSubscribe("")

	for {
		// Receive messages from the publisher via the proxy
		msg, err := subscriber.RecvMessage(0)
		if err != nil {
			log.Println("Error receiving message:", err)
		}
		log.Println("Received:", msg)
	}
}

