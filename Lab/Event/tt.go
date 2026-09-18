package main

import (
        "fmt"
        "log"
        "net"
        "time"

        "github.com/pebbe/zmq4"
)

func main() {
        ctx, err := zmq4.NewContext()
        if err != nil {
                log.Fatal("Error creating ZMQ context:", err)
        }
        defer ctx.Term()

        responder, err := ctx.NewSocket(zmq4.REP)
        if err != nil {
                log.Fatal("Error creating ZMQ socket:", err)
        }
        defer responder.Close()

        err = responder.Bind("tcp://*:5555")
        if err != nil {
                log.Fatal("Error binding ZMQ socket:", err)
        }

        ln, err := net.Listen("tcp", ":8080")
        if err != nil {
                log.Fatal("Error listening on TCP:", err)
        }
        defer ln.Close()

        poller := zmq4.NewPoller()
        poller.Add(responder, zmq4.POLLIN) // Add ZMQ socket to poller

        fmt.Println("Server started. Listening on :8080 (TCP) and :5555 (ZMQ)")

        tcpConnections := make(map[uintptr]net.Conn)

        for {
                // Non-blocking accept (check for connections without blocking)
                conn, err := ln.Accept()
                if err != nil {
                        if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
                                // It is just a timeout, no worries
                        } else if err != nil {
                                log.Println("Accept Error:", err)
                        }
                } else {
						file, _ := conn.(*net.TCPConn).File()
                        fd := file.Fd()
                        fmt.Println("Accepted TCP connection:", conn.RemoteAddr(), "FD:", fd)
                        tcpConnections[uintptr(fd)] = conn
                }

                // Handle ZMQ events (with timeout)
                polledItems, err := poller.Poll(100 * time.Millisecond) // Timeout for ZMQ poll
                if err != nil {
                        log.Println("Poll error:", err)
                } else {
                        for _, item := range polledItems {
                                if item.Socket == responder && item.Events&zmq4.POLLIN != 0 {
                                        msg, err := responder.Recv(0)
                                        if err != nil {
                                                log.Println("Error receiving ZMQ message:", err)
                                                continue
                                        }
                                        fmt.Println("Received ZMQ message:", string(msg))
                                        responder.Send("World", 0)
                                }
                        }
                }

                // Handle existing TCP connections (read with timeout)
                for fd, conn := range tcpConnections {
                        conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond)) // Timeout for TCP read
                        buf := make([]byte, 1024)
                        n, err := conn.Read(buf)
                        if err != nil {
                                if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
                                        continue // Just a timeout
                                } else {
                                        log.Println("TCP Read Error:", err)
                                        conn.Close()
                                        delete(tcpConnections, fd)
                                        continue
                                }
                        }
                        fmt.Printf("Received TCP data from %s: %s\n", conn.RemoteAddr(), buf[:n])
                        _, err = conn.Write(buf[:n])
                        if err != nil {
                                log.Println("TCP Write Error:", err)
                                conn.Close()
                                delete(tcpConnections, fd)
                        }
                }
        }
}
