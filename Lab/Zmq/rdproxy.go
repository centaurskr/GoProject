//
// ROUTER & DEALER Proxy
//
//              ┌─────────────────────────────┐
// REQ          │  (ROUTER) rdproxy (DEALER)  │  REP
// tcp://*:5555 │          (PUB)              │  tcp://*:5556
//              └─────────────────────────────┘
//                       SUB tcp://*:5557
// 
package main

import (
    "fmt"
	"os"
	"time"
    "github.com/pebbe/zmq4"
	"centolib/tbzmq"
	Log "centolib/logout"
)

var (
	rCount, dCount, cCount int
)

func goMonitor(name string, s *zmq4.Socket){
	Log.Infor("%s Socket Monitor Start", name)
	for{
		event,addr, _, _ := s.RecvEvent(0)
		switch event{
			case zmq4.EVENT_ACCEPTED :
				if name == "ROUTER"{
					rCount ++
					Log.Infor("ROUTER(REQ)- Connect from %s, Count(%d)", 
						addr,rCount)
				}else if name == "DEALER"{
					dCount ++
					Log.Infor("DEALER(REP)- Connect from %s, Count(%d)",
						 addr, dCount)
				}else if name == "CAPTURE"{
					cCount ++
					Log.Infor("CAPTURE - Connect from %s, Count(%d)",
						 addr, cCount)
				}
			case  zmq4.EVENT_DISCONNECTED:
				if name == "ROUTER"{
					rCount --
					Log.Infor("ROUTER(REQ)- Disconnect from %s, Count(%d)", 
						addr,rCount)
				}else if name == "DEALER"{
					dCount --
					Log.Infor("DEALER(REP)- Disconnect from %s, Count(%d)",
						 addr, dCount)
				}else if  name == "CAPTURE"{
					cCount --
					Log.Infor("CAPTURE - Disconnect from %s, Count(%d)",
						 addr, cCount)
				}
		} // Switch
		Log.Infor("Connection Count: Rotuer-REQ(%d) Dealer-REP(%d) Capture(%d)",
			rCount, dCount, cCount)
	} // For Loop
}

// Message Capture를 위한 go Routine
func goCapture(ctx *zmq4.Context, subpoint, pubpoint string){


	s, _ := ctx.NewSocket(zmq4.SUB)
	s.Connect(subpoint)
	s.SetSubscribe("")
	defer s.Close()

	// for Capture client
	p, _ := ctx.NewSocket(zmq4.PUB)
	p.Bind(pubpoint)
	defer p.Close()

	pmon, _:= tbzmq.MakeMonitorSocket("", p)
	go goMonitor("CAPTURE", pmon)
	defer pmon.Close()

	time.Sleep(100 * time.Microsecond)

	Log.Infor("Capture Start")
	for {
		msg, _ := s.RecvMessage(0)
		//fmt.Printf("[%s]-%s\n", msg[0], msg[1])
		// 접속한 Capture Client가 없으면 보내지 않음
		if cCount <= 0{ continue;}
		p.SendMessage(msg)
	}
}
func main() {
	Log.InitLogger("", Log.CONSOLE_MODE);defer Log.Close()

	rCount = 0
	dCount = 0
	cCount = 0

	rpoint := "tcp://*:5555"
	dpoint := "tcp://*:5556"
    router, _ := zmq4.NewSocket(zmq4.ROUTER)
	dealer, _ := zmq4.NewSocket(zmq4.DEALER)
    defer router.Close()
    defer dealer.Close()
    router.Bind(rpoint)
    dealer.Bind(dpoint)

	Log.Infor("REQ/REP Proxy Start! %s(REQ) ⮕- %s(REP)", rpoint, dpoint)

	rmon,_ := tbzmq.MakeMonitorSocket("", router)
	go goMonitor("ROUTER", rmon)
	defer rmon.Close()
	dmon,_ := tbzmq.MakeMonitorSocket("", dealer)
	go goMonitor("DEALER", dmon)
	defer dmon.Close()

	// Message capture 등을 위한 내부 Proxy
	proxyFPoint := fmt.Sprintf("inproc://front_%d", os.Getpid()) // PUB
	proxyBPoint := fmt.Sprintf("inproc://back__%d", os.Getpid()) // SUB
	// Zmq context 얻기
	ctx, _ := router.Context()
	// 내부 Proxy 실행
	err := tbzmq.RunPubSubProxy(ctx, proxyFPoint, proxyBPoint)
	if err != nil{
		Log.Error("RubPubSubProxy() : %v", err)
		os.Exit(0)
	}
	// Capture를 위한 go routine
	go goCapture(ctx, proxyBPoint, "tcp://*:5557")
	Log.Infor("Capture point(SUB) : tcp://*:5557")

	// Capture 전송을 위한 socket
	capture,_ := zmq4.NewSocket(zmq4.PUB)
	capture.Connect(proxyFPoint)
	time.Sleep(100 * time.Microsecond)
	
	poller := zmq4.NewPoller()
	poller.Add(router, zmq4.POLLIN)
	poller.Add(dealer, zmq4.POLLIN)

    for {
		polls, _ := poller.Poll(-1)
		for _, polled := range polls{
			switch s := polled.Socket; s {
			case router:
        		msg, _ := router.RecvMessageBytes(0)
				capture.SendMessage("REQ", string(msg[2][:])) // [2]Message부분
        		dealer.SendMessage(msg)
			case dealer:
        		msg, _ := dealer.RecvMessageBytes(0)
				capture.SendMessage("REP", string(msg[2][:])) // [2]Message부분
        		router.SendMessage(msg)

			} // switch
		} // for (polled)
    } // main for loop
}

