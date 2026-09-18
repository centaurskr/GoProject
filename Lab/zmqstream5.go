package main

import (
	"bufio"
	"bytes"
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

	// Use a map to store partial messages by identity
	messageBuffer := make(map[string]*bytes.Buffer)

	for {
		// Receive both identity and data
		parts, err := socket.RecvMessage(0)
		if err != nil {
			log.Fatalf("Failed to receive message: %v", err)
		}

		if len(parts) < 2 {
			continue
		}

		identity := parts[0]
		data := parts[1]

		// Initialize buffer if not exists
		if _, exists := messageBuffer[identity]; !exists {
			messageBuffer[identity] = &bytes.Buffer{}
		}

		// Append received data to the buffer
		messageBuffer[identity].WriteString(data)

		// Use bufio to read line by line
		scanner := bufio.NewScanner(messageBuffer[identity])
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("Received from [%x]: %s\n", identity, line)
		}

		// Clear buffer after processing lines
		messageBuffer[identity].Reset()
		messageBuffer[identity].WriteString(scanner.Text())
	}
}

