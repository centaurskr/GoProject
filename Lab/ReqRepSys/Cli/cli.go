//
// TCP Client
//
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 2{
		fmt.Printf("TCP Client Program\n")
		fmt.Printf("Stdin -> TCP Send -> TCP Recv\n")
		fmt.Printf("Using) %s <connet port>\n", os.Args[0])
		os.Exit(0)
	}
	// Server address and port
	server := "127.0.0.1:"+os.Args[1]

	// Connect to the server
	conn, err := net.Dial("tcp", server)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to server at", server)

	// Input loop
	for {
		// Read user input
		fmt.Print("Enter message: ")
		reader := bufio.NewReader(os.Stdin)
		message, _ := reader.ReadString('\n')

		// Send message to server
		_, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Error sending message:", err)
			return
		}

		// Receive response from server
		reply := make([]byte, 1024)
		n, err := conn.Read(reply)
		if err != nil {
			fmt.Println("Error reading response:", err)
			return
		}

		fmt.Println("Server reply:", string(reply[:n]))

		// Exit condition
		if message == "exit\n" {
			fmt.Println("Exiting...")
			break
		}
	}
}

