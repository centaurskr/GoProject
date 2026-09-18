package main

import (
	"fmt"
	"log"
	"time"

	zmq "github.com/pebbe/zmq4"
)

func RunBasicPubSubProxy(f, b string) error {
	context, err := zmq.NewContext()
	if err != nil {
		return fmt.Errorf("failed to create zmq context: %v", err)
	}
	defer context.Term()

	// Set up frontend (XSUB) socket to receive messages from publishers
//	front, err := context.NewSocket(zmq.XSUB)
	front, err := context.NewSocket(zmq.SUB)
	if err != nil {
		return err
	}
	defer front.Close()
	err = front.Bind(f)
	if err != nil {
		return fmt.Errorf("failed to bind XSUB socket on %s: %v", f, err)
	}
	log.Printf("XSUB socket bound to %s\n", f)
	front.SetSubscribe("")

	// Set up backend (XPUB) socket to send messages to subscribers
	back, err := context.NewSocket(zmq.XPUB)
	if err != nil {
		return err
	}
	defer back.Close()
	err = back.Bind(b)
	if err != nil {
		return fmt.Errorf("failed to bind XPUB socket on %s: %v", b, err)
	}
	log.Printf("XPUB socket bound to %s\n", b)

	// Main proxy loop: receive from XSUB, send to XPUB
	go func() {
		for {
			msg, err := front.RecvMessage(0)
			if err != nil {
				log.Println("Error receiving message on XSUB:", err)
				continue
			}
			log.Println("Received message on XSUB:", msg)

			_, err = back.SendMessage(msg)
			if err != nil {
				log.Println("Error sending message to XPUB:", err)
				continue
			}
			log.Println("Forwarded message to XPUB:", msg)
		}
	}()

	// Keep the program running to maintain the proxy
	for {
		time.Sleep(1 * time.Second)
	}
}

func main() {
	err := RunBasicPubSubProxy("tcp://127.0.0.1:32768", "tcp://127.0.0.1:32769")
	if err != nil {
		log.Fatal(err)
	}
}

