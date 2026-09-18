package main

import (
	"fmt"
	"log"
	"strings"

	zmq "github.com/pebbe/zmq4"
)

func main() {
	socket, err := zmq.NewSocket(zmq.STREAM)
	if err != nil {
		log.Fatalf("Failed to create socket: %v", err)
	}
	defer socket.Close()

	err = socket.Bind("tcp://*:32767")
	if err != nil {
		log.Fatalf("Failed to bind socket: %v", err)
	}

	fmt.Println("ZeroMQ STREAM Server is running...")

	// Message buffer for each client
	messageBuffer := make(map[string]string)

	for {
		// Receive the next part of the message (identity and data)
		parts, err := socket.RecvMessage(0)
		if err != nil {
			log.Fatalf("Failed to receive message: %v", err)
		}

		if len(parts) < 2 {
			continue // Skip if it's just an identity frame
		}

		identity := parts[0] // Unique identifier for each client
		data := parts[1]     // Actual received data

		// Append received data to client's buffer
		messageBuffer[identity] += data

		// Process all complete messages
		for {
			if idx := strings.Index(messageBuffer[identity], "\n"); idx != -1 {
				// Extract and print one full message
				message := messageBuffer[identity][:idx]
				fmt.Printf("Received from [%s]: %s\n", identity, message)

				// Remove processed message from buffer
				messageBuffer[identity] = messageBuffer[identity][idx+1:]
			} else {
				break // No full message left to process
			}
		}
	}
}

