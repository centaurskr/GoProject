package main

import (
    zmq "github.com/pebbe/zmq4"
    "fmt"
    "time"
)

func startProxy() {
    // Create frontend (SUB) and backend (PUB) sockets
    frontend, _ := zmq.NewSocket(zmq.XSUB)
    backend, _ := zmq.NewSocket(zmq.XPUB)

    defer frontend.Close()
    defer backend.Close()

    // Bind both frontend and backend to inproc:// transport
    frontend.Bind("inproc://frontend")
    backend.Bind("inproc://backend")

    //frontend.Bind("tcp://localhost:32786")
    //backend.Bind ("tcp://localhost:32787")

	time.Sleep(100 * time.Microsecond)

    // Proxy between the frontend and backend
    fmt.Println("Starting proxy...")
    zmq.Proxy(frontend, backend, nil)
}

func startPublisher() {
    // This thread will connect to the proxy's backend and send messages
    publisher, _ := zmq.NewSocket(zmq.PUB)
    defer publisher.Close()
    publisher.Connect("inproc://frontend")
    //publisher.Connect("tcp://localhost:32786")

	time.Sleep(100 * time.Microsecond)

	fmt.Println("Start Publisher")

    for i := 0; i < 10; i++ {
        msg := fmt.Sprintf("Message %d", i)
        fmt.Println("Publisher sending:", msg)
        publisher.SendMessage(msg)
        time.Sleep(time.Second)
    }
}

func startSubscriber() {
    // This thread will connect to the proxy's frontend and receive messages
    subscriber, _ := zmq.NewSocket(zmq.SUB)
    defer subscriber.Close()
    subscriber.Connect("inproc://backend")
    //subscriber.Connect("tcp://localhost:32787")
    subscriber.SetSubscribe("") // Subscribe to all messages
	
	fmt.Println("Start Subscribe")

    for {
        msg, _ := subscriber.Recv(0)
        fmt.Println("Subscriber received:", msg)
    }
}

func main() {
    // Start proxy in a separate thread
    go startProxy()

    time.Sleep(time.Second) // Allow proxy to start

    // Start the publisher and subscriber in separate threads
    go startPublisher()
    go startSubscriber()

    // Run indefinitely
    select {}
}

