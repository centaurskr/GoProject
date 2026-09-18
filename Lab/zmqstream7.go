package main

import (
	"encoding/binary"
	"fmt"
	"log"
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

	// Create the poller
	poller := zmq.NewPoller()

	// Add the stream socket to the poller
	poller.Add(socket, zmq.POLLIN)

	// Buffer to track data per client
	messageBuffer := make(map[string][]byte)

	for {
		// Wait for events (use Pollin to check for incoming messages)
		polled, err := poller.Poll(-1) // Block indefinitely, waiting for events
		if err != nil {
			log.Fatalf("Poll failed: %v", err)
		}

		// Iterate through the sockets that are ready
		for _, poll := range polled {
			if socket == poll.Socket {
				// Receive both identity and data
				parts, err := socket.RecvMessage(0)
				if err != nil {
					log.Fatalf("Failed to receive message: %v", err)
				}
				if len(parts) < 2 { continue; }

				identity := parts[0]
				data := []byte(parts[1])

				if _, exist := messageBuffer[identity];!exist{
					messageBuffer[identity] = []byte{}
					fmt.Printf("New Connect [%x]\n", identity)
					continue
				}

				if len(data) == 0{
					fmt.Printf("Client disconnected: [%x]\n", identity)
    				delete(messageBuffer, identity) // Cleanup buffer
    				continue
				}

				// Append new data to the client's buffer
				messageBuffer[identity] = append(messageBuffer[identity], data...)

				// Process messages from buffer
				for len(messageBuffer[identity]) >= 4 {
					// Read 4-byte length header
					msgLen := binary.BigEndian.Uint32(messageBuffer[identity][:4])
					if len(messageBuffer[identity]) < int(4+msgLen) {
						break // Wait for full message
					}

					// Extract message
					msg := string(messageBuffer[identity][4 : 4+msgLen])
					fmt.Printf("             Received from [%x]: %s\n", identity, msg)

				    responseSize := make([]byte, 4)
                    binary.BigEndian.PutUint32(responseSize, uint32(len(msg)))

                    response := append(responseSize, msg...) // Prepend size
                    socket.SendMessage(identity, response)

					// Remove processed message from buffer
					messageBuffer[identity] = messageBuffer[identity][4+msgLen:]
				}
			}
		}
	}
}

