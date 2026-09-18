// Basic Reliable Queuing (Simple Pirate Pattern) 
// ┌────────┐ ┌────────┐ ┌────────┐
// │ Client │ │ Client │ │ Client │   
// │ Retry  │ │ Retry  │ │ Retry  │   
// │  REQ   │ │  REQ   │ │  REQ   │   
// └────────┘ └────────┘ └────────┘
//     │           │           │
//     ╰───────────┼───────────╯
//                🠗│
//          ┌──────┴──────────┐
//          │   ROUTER        │
//          │   Load balancer │
//          │   ROUTER        │
//          └──────┬──────────┘
//                ↑│
//     ╭───────────┼──────────╮
//     │           │          │
// ┌────────┐ ┌────────┐ ┌────────┐
// │  REQ   │ │  REQ   │ │  REQ   │   
// │ Wroker │ │ Wroker │ │ Wroker │   
// └────────┘ └────────┘ └────────┘
//  Simple Pirate broker
//  This is identical to load-balancing pattern, with no reliability
//  mechanisms. It depends on the client for recovery. Runs forever.
//

package main

import (
	zmq "github.com/pebbe/zmq4"
)

const LRU_READY = "\001"

func main() {
	context, _ := zmq.NewContext()
	defer context.Term()

	frontend, _ := context.NewSocket(zmq.ROUTER)
	defer frontend.Close()
	frontend.Bind("tcp://*:5555") //  For clients

	backend, _ := context.NewSocket(zmq.ROUTER)
	defer backend.Close()
	backend.Bind("tcp://*:5556") //  For workers

	//  Queue of available workers
	workers := make([][]byte, 0, 0)

	for {
/*****************
		items := zmq.PollItems{
			zmq.PollItem{Socket: backend, Events: zmq.POLLIN},
			zmq.PollItem{Socket: frontend, Events: zmq.POLLIN},
		}


		//  Poll frontend only if we have available workers
		if len(workers) > 0 {
			zmq.Poll(items, -1)
		} else {
			zmq.Poll(items[:1], -1)
		}
***************/
		poller := zmq.NewPoller()
		poller.Add(backend,  zmq.POLLIN)
		poller.Add(frontend, zmq.POLLIN)

		items, err := poller.Poll(-1)
		if err != nil{
			panic(err) //  Interrupted
		}
		for _, item := range items{
			if item.Socket == backend && item.Events == zmq.POLLIN {
				//  Use worker identity for load-balancing
				//msg, err := backend.RecvMultipart(0)
				msg, err := backend.RecvMessageBytes(0)
				if err != nil {
					panic(err) //  Interrupted
				}
				address := msg[0]
				workers = append(workers, address)

				//  Forward message to client if it's not a READY
				if reply := msg[2:]; string(reply[0]) != LRU_READY {
					//frontend.SendMultipart(reply, 0)
					frontend.SendMessage(reply, 0)
				}

			}else if item.Socket == frontend && item.Events == zmq.POLLIN {
				//  Get client request, route to first available worker
				//msg, err := frontend.RecvMultipart(0)
				msg, err := frontend.RecvMessageBytes(0)
				if err != nil {
					panic(err) //  Interrupted
				}
				last := workers[len(workers)-1]
				workers = workers[:len(workers)-1]
				request := append([][]byte{last, nil}, msg...)
				//backend.SendMultipart(request, 0)
				backend.SendMessage(request, 0)
			}
		}
	}
}
