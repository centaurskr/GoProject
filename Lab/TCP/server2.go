package main

import (
    "fmt"
    "log"
    "net"
    "syscall"
)

func main() {
    port := ":32767"

    // Create TCP listener
    listener, err := net.Listen("tcp", port)
    if err != nil {
        log.Fatal("Error starting server:", err)
    }
    defer listener.Close()

    // Get file descriptor of listener
    tcpListener := listener.(*net.TCPListener)
    listenerFD, err := tcpListener.File()
    if err != nil {
        log.Fatal("Error getting listener FD:", err)
    }

    // Create epoll instance
    epollFD, err := syscall.EpollCreate1(0)
    if err != nil {
        log.Fatal("Error creating epoll:", err)
    }

    // Add listener FD to epoll
    event := &syscall.EpollEvent{
        Events: syscall.EPOLLIN, // Listen for input (new connections)
        Fd:     int32(listenerFD.Fd()),
    }
    if err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, int(listenerFD.Fd()), event); err != nil {
        log.Fatal("Error adding listener FD to epoll:", err)
    }

    // Client connections map
    connections := make(map[int]net.Conn)

    log.Println("Server is listening on", port)

    for {
        // Wait for events (up to 100 events at a time)
        events := make([]syscall.EpollEvent, 100)
        n, err := syscall.EpollWait(epollFD, events, -1)
        if err != nil {
            log.Println("Epoll wait error:", err)
            continue
        }

        // Process events
        for i := 0; i < n; i++ {
            fd := int(events[i].Fd)

            if fd == int(listenerFD.Fd()) {
                // New client connection
                conn, err := listener.Accept()
                if err != nil {
                    log.Println("Accept error:", err)
                    continue
                }

                // Get connection file descriptor
                tcpConn := conn.(*net.TCPConn)
                connFD, err := tcpConn.File()
                if err != nil {
                    log.Println("Error getting connection FD:", err)
                    conn.Close()
                    continue
                }

                // Add new client to epoll
                connEvent := &syscall.EpollEvent{
                    Events: syscall.EPOLLIN, // Monitor for incoming data
                    Fd:     int32(connFD.Fd()),
                }
                if err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, int(connFD.Fd()), connEvent); err != nil {
                    log.Println("Error adding client FD to epoll:", err)
                    conn.Close()
                    continue
                }

                // Store the connection
                connections[int(connFD.Fd())] = conn
                log.Println("New client connected:", conn.RemoteAddr())

            } else {
                // Data from client
                if conn, ok := connections[fd]; ok {
                    handleClient(conn, fd, epollFD, connections)
                }
            }
        }
    }
}

// Handle client communication
func handleClient(conn net.Conn, fd int, epollFD int, connections map[int]net.Conn) {
    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        log.Println("Client disconnected:", conn.RemoteAddr())

        // 🔥 Remove from epoll
        if err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_DEL, fd, nil); err != nil {
            log.Println("Error removing client FD from epoll:", err)
        }

        // 🔥 Close connection and delete from map
        conn.Close()
        delete(connections, fd)
        return
    }

    fmt.Println("Received:", string(buffer[:n]))
    conn.Write([]byte("Message received\n"))
}

