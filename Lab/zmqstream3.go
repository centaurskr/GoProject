package main

import (
	"fmt"
	"log"

	zmq "github.com/pebbe/zmq4"
)

func main() {
	// Create a STREAM socket
	socket, err := zmq.NewSocket(zmq.STREAM)
	if err != nil {
		log.Fatalf("Failed to create socket: %v", err)
	}
	defer socket.Close()

	// Bind to TCP port
	err = socket.Bind("tcp://*:32767")
	if err != nil {
		log.Fatalf("Failed to bind socket: %v", err)
	}

	fmt.Println("ZeroMQ STREAM Server is running...")

	for {
		// Read first frame (identity)
		identity, err := socket.Recv(0)
		if err != nil {
			log.Fatalf("Failed to receive identity: %v", err)
		}

		// Read second frame (actual data)
		msg, err := socket.Recv(0)
		if err != nil {
			log.Fatalf("Failed to receive message: %v", err)
		}

		fmt.Printf("Received from [%s]: %s\n", identity, msg)
	}
}

