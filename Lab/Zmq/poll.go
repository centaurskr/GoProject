package main

import (
	"log"
	//"os"
	"syscall"
	"time"

	zmq "github.com/pebbe/zmq4"
	"centolib/tbzmq"
)

// Helper function to create a read-write pipe for timer events
func createTimerPipe() (readFD, writeFD int, err error) {
	// Create a pipe
	pipe := make([]int, 2)
	err = syscall.Pipe(pipe)
	if err != nil {
		return -1, -1, err
	}
	return pipe[0], pipe[1], nil
}

func main() {
	context, err := zmq.NewContext()
	if err != nil {
		log.Fatalf("Failed to create ZMQ context: %v", err)
	}
	defer context.Term()

	// Setup REP socket
	repSocket, err := context.NewSocket(zmq.REP)
	if err != nil {
		log.Fatalf("Failed to create REP socket: %v", err)
	}
	defer repSocket.Close()

	err = repSocket.Bind("tcp://*:5555")
	if err != nil {
		log.Fatalf("Failed to bind REP socket: %v", err)
	}

	// Create a pipe for timer events
	readFD, writeFD, err := createTimerPipe()
	if err != nil {
		log.Fatalf("Failed to create timer pipe: %v", err)
	}
	defer syscall.Close(readFD)
	defer syscall.Close(writeFD)

	// Create ZeroMQ poller
	poller := zmq.NewPoller()
	poller.Add(repSocket, zmq.POLLIN) // Add REP socket to poller
	tbzmq.AddFD(poller, readFD, zmq.POLLIN)  // Add timer pipe's read file descriptor to poller

	// Start the ticker in a goroutine
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			// Write a single byte to the write end of the pipe
			_, err := syscall.Write(writeFD, []byte{1})
			if err != nil {
				log.Printf("Failed to write to timer pipe: %v", err)
				return
			}
		}
	}()

	log.Println("Starting main loop...")

	for {
		// Poll for events
		polledSockets, err := poller.Poll(-1) // Wait indefinitely for events
		if err != nil {
			log.Fatalf("Error while polling sockets: %v", err)
		}

		for _, polled := range polledSockets {
			switch s := polled.Socket; s {
			case repSocket:
				// Handle incoming message on REP socket
				msg, err := repSocket.RecvMessage(0)
				if err == nil {
					log.Printf("Received message on REP: %v", msg)
					// Send a response
					err = repSocket.SendMessage("Response to: " + msg[0])
					if err != nil {
						log.Printf("Failed to send response: %v", err)
					}
				} else {
					log.Printf("Error receiving message on REP: %v", err)
				}

			default:
				if polled.Fd == readFD {
					// Handle timer event
					buf := make([]byte, 1)
					_, err := syscall.Read(readFD, buf) // Read the timer "signal"
					if err == nil {
						log.Println("Timer tick - perform periodic task here")
					} else {
						log.Printf("Error reading timer event: %v", err)
					}
				}
			}
		}
	}
}

