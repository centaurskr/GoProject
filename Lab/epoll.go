// 아직 해결하지 못하는 문제. bb
// zmq evnet가 한번만 걸린다..
//
package main

import (
	"fmt"
	"log"
	"syscall"
	"time"

	"github.com/pebbe/zmq4"
)

func main() {
	// Create a ZeroMQ SUB socket
	subscriber, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		log.Fatalf("Failed to create SUB socket: %v", err)
	}
	defer subscriber.Close()

	// Connect to the publisher
	err = subscriber.Connect("tcp://127.0.0.1:7908")
	if err != nil {
		log.Fatalf("Failed to connect to publisher: %v", err)
	}

	// Subscribe to all topics
	err = subscriber.SetSubscribe("")
	if err != nil {
		log.Fatalf("Failed to subscribe to topics: %v", err)
	}

	// Get the file descriptor for the SUB socket
	subFd, err := subscriber.GetFd()
	if err != nil {
		log.Fatalf("Failed to get SUB socket file descriptor: %v", err)
	}

	// Create a pipe for the timer
	var pipeFds [2]int
	if err := syscall.Pipe(pipeFds[:]); err != nil {
		log.Fatalf("Failed to create pipe: %v", err)
	}
	defer syscall.Close(pipeFds[0])
	defer syscall.Close(pipeFds[1])

	// Create an epoll instance
	epollFd, err := syscall.EpollCreate1(0)
	if err != nil {
		log.Fatalf("Failed to create epoll instance: %v", err)
	}
	defer syscall.Close(epollFd)

	// Add the SUB socket to epoll
	err = syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, int(subFd), 
							&syscall.EpollEvent{
								Events: syscall.EPOLLIN,
								Fd:     int32(subFd),
							})
	if err != nil {
		log.Fatalf("Failed to add SUB socket to epoll: %v", err)
	}

	// Add the pipe read end to epoll
	err = syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, pipeFds[0], 
							&syscall.EpollEvent{
								Events: syscall.EPOLLIN,
								Fd:     int32(pipeFds[0]),
							})
	if err != nil {
		log.Fatalf("Failed to add pipe to epoll: %v", err)
	}

	// Timer setup
	timerInterval := 5 * time.Second
	go func() {
		ticker := time.NewTicker(timerInterval)
		defer ticker.Stop()

		for range ticker.C {
			// Write to the pipe to signal the timer event
			_, err := syscall.Write(pipeFds[1], []byte("t"))
			if err != nil {
				log.Printf("Failed to write to pipe: %v", err)
			}
		}
	}()

	fmt.Println("Starting main loop...")

	// Buffer for epoll events
	events := make([]syscall.EpollEvent, 10)

	// Main loop
	for {

	    n, err := syscall.EpollWait(epollFd, events, -1)
	    if err != nil {
	        if err == syscall.EINTR {
	            continue // Retry on interrupt
	        }
	        log.Fatalf("EpollWait error: %v", err)
	    }
	
	    //log.Printf("....... Epoll out.......RETURN [%d]\n", n)
	
	    for i := 0; i < n; i++ {
	        fd := int(events[i].Fd)
	        log.Printf("polled fd[%d], index[%d] events[%x] EPOLLIN[%x]\n", 
	            fd, i, events[i].Events, syscall.EPOLLIN)
	
	        if events[i].Events&syscall.EPOLLIN == 0 {
	            log.Println("Skipping non-EPOLLIN event")
	            continue
	        }
	
	        switch fd {
	        case int(subFd):
	            log.Println("Handling ZeroMQ message...")
	            //msg, err := subscriber.RecvMessage(0)
				msg, err := subscriber.RecvMessage(zmq4.DONTWAIT)
	            if err != nil {
        			log.Printf("Error receiving message (non-blocking): %v", err)
        			if zmq4.AsErrno(err) == zmq4.Errno(syscall.EAGAIN) {
            			log.Println("ZeroMQ socket is not ready to read")
        			}
	            } else {
	                fmt.Printf("Received message: %v\n", msg)
	            }
	
	        case pipeFds[0]:
	            log.Println("Handling timer event...")
	            buf := make([]byte, 1)
	            n, err := syscall.Read(pipeFds[0], buf)
	            if err != nil {
	                log.Printf("Error reading from pipe: %v", err)
	            } else if n > 0 {
	                fmt.Println("Timer event triggered")
	            }
	        }
	    }
	}
}

