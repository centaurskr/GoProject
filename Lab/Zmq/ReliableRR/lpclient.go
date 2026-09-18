// Lazy Pirate client in GO
// Use zmq_poll to do a safe request-reply
// To run, start lpserver and then randomly kill/restart it
// Client-Side Reliability 
package main
import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"strconv"
	"time"
)

const (
	REQUEST_TIMEOUT = time.Duration(2500) * time.Millisecond
	REQUEST_RETRIES = 3
	SERVER_ENDPOINT = "tcp://localhost:5555"
)

func main() {

	context, _ := zmq.NewContext()
	defer context.Term()

	fmt.Println("I : Connection to server ....")
	client, _ := context.NewSocket(zmq.REQ)

	client.Connect(SERVER_ENDPOINT)


	for sequence, retriesLeft := 1, REQUEST_RETRIES; retriesLeft > 0; sequence ++{
		fmt.Printf("I : Sending (%d)\n", sequence)

		client.Send(strconv.Itoa(sequence), 0)


		for expectReply := true; expectReply; {
			//  Poll socket for a reply, with timeout
			poller := zmq.NewPoller()
			poller.Add(client, zmq.POLLIN)

			items, err :=poller.Poll(REQUEST_TIMEOUT)
			if  err != nil{ panic(err); }


			//  .split process server reply
			//  Here we process a server reply and exit our loop if the
			//  reply is valid. If we didn't a reply we close the client
			//  socket and resend the request. We try a number of times
			//  before finally abandoning:
			if len(items) != 0 && items[0].Socket == client && items[0].Events == zmq.POLLIN {


				//  We got a reply from the server, must match sequence
				reply, err := client.Recv(0)
				if err != nil {
					panic(err) //  Interrupted
				}
				if replyInt, err := strconv.Atoi(string(reply)); replyInt == sequence && err == nil {
					fmt.Printf("\tI: Server replied OK (%s)\n", reply)
					retriesLeft = REQUEST_RETRIES
					expectReply = false
				} else {
					fmt.Printf("E: Malformed reply from server: %s", reply)
				}
			}else if retriesLeft--; retriesLeft == 0 {
				fmt.Println("\tE: Server seems to be offline, abandoning")
				client.SetLinger(0)
				client.Close()
				break
			} else {
				fmt.Println("\tW: No response from server, Reconnect and retrying...")
				//  Old socket is confused; close it and open a new one
				client.SetLinger(0)
				client.Close()
				client, _ = context.NewSocket(zmq.REQ)
				client.Connect(SERVER_ENDPOINT)
				fmt.Printf("\tI: Resending (%d)\n", sequence)
				//  Send request again, on new socket
				client.Send(strconv.Itoa(sequence), 0)
			}
			fmt.Println("\t\tInner loop Break")
		} // for expectReply
	} // for sequence
} // main
