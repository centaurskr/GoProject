// REQ REP Proxy
//
//       ┌─────────────────────────────────┐
// REQ   │  (ROUTER) ReqRepProxy (DEALER)  │  REP
//       │  capture(PUB)     control(REP)  │            
//       └─────────────────────────────────┘
//              SUB            REQ
//
//
package main
import (
	"fmt"
	"os"
	"time"
	"encoding/json"
	"centolib/tbzmq"
	"centolib/api"
	Log "centolib/netlog"
	zmq "github.com/pebbe/zmq4"
)

var ProcInfor *api.ProxyInfor

func goMonitor(name string, s *zmq.Socket){
    Log.Infor("%s Socket Monitor Start", name)
    for{
        event,addr, _, _ := s.RecvEvent(0)
        switch event{
            case zmq.EVENT_ACCEPTED :
                if name == "ROUTER"{
                    ProcInfor.FrontCount ++
                    Log.Infor("ROUTER(REQ)- Connect from %s, Count(%d)",
                        addr,ProcInfor.FrontCount)
                }else if name == "DEALER"{
                    ProcInfor.BackCount ++
                    Log.Infor("DEALER(REP)- Connect from %s, Count(%d)",
                         addr, ProcInfor.BackCount)
                }else if name == "CAPTURE"{
                    ProcInfor.CaptureCount ++
                    Log.Infor("CAPTURE - Connect from %s, Count(%d)",
                         addr, ProcInfor.CaptureCount)
                }
            case  zmq.EVENT_DISCONNECTED:
                if name == "ROUTER"{
                    ProcInfor.FrontCount --
                    Log.Infor("ROUTER(REQ)- Disconnect from %s, Count(%d)",
                        addr,ProcInfor.FrontCount)
                }else if name == "DEALER"{
                    ProcInfor.BackCount --
                    Log.Infor("DEALER(REP)- Disconnect from %s, Count(%d)",
                         addr, ProcInfor.BackCount)
                }else if  name == "CAPTURE"{
                    ProcInfor.CaptureCount --
                    Log.Infor("CAPTURE - Disconnect from %s, Count(%d)",
                         addr, ProcInfor.CaptureCount)
                }
        } // Switch
        Log.Infor("Connection Count: Rotuer-REQ(%d) Dealer-REP(%d) Capture(%d)",
            ProcInfor.FrontCount, 
			ProcInfor.BackCount, 
			ProcInfor.CaptureCount)
    } // For Loop
}
// Message Capture를 위한 go Routine
func goCapture(ctx *zmq.Context, subpoint, pubpoint string){
    s, _ := ctx.NewSocket(zmq.SUB)
    s.Connect(subpoint)
    s.SetSubscribe("")
    defer s.Close()

    // for Capture client
    p, _ := ctx.NewSocket(zmq.PUB)
    p.Bind(pubpoint)
    defer p.Close()
    time.Sleep(100 * time.Microsecond)

    pmon, _:= tbzmq.MakeMonitorSocket("", p)
    go goMonitor("CAPTURE", pmon)
    defer pmon.Close()

    Log.Infor("🔵 Capture service start! (PUB/SUB) %s", pubpoint)
    for {
        msg, _ := s.RecvMessage(0)
		// Capture Client가 없으면 보내지않음
        if ProcInfor.CaptureCount <= 0{ continue;}
        p.SendMessage(msg)
    }
}
// Proxy Control을 위한 go Routine
func goControl(ctx *zmq.Context, endpoint string){
	ctlsocket, _ := ctx.NewSocket(zmq.REP)
	ctlsocket.Bind(endpoint)
	time.Sleep(1000 * time.Microsecond)

	Log.Infor("🔵 Control service start! (%s-REQ/REP)", endpoint)
	for {
		msg, _:=ctlsocket.RecvMessage(0)
		Log.Infor("control Message [%s]", msg[0])
		switch msg[0]{
			case "INFOR" :
				msg , _ := json.Marshal(*ProcInfor)
				ctlsocket.SendMessage(string(msg))
			case "PAUSE" : 
				ProcInfor.Pause = true
				ctlsocket.SendMessage("PAUSE OK")
			case "RUN" : 
				ProcInfor.Pause = false
				ctlsocket.SendMessage("RUN OK")
			default :
				msg := fmt.Sprintf("Proxy command : INFOR, PAUSE, RUN")
				ctlsocket.SendMessage(msg)
		}
	}
}
///////////// Main Function //////////////////////////////////////////////
func main(){
	if len(os.Args) != 2{
		fmt.Printf("⛔ REQ/REP Proxy Network Log version\n")
		fmt.Printf("   Using) %s <SERVERNAME>\n", os.Args[0])
		fmt.Printf("   <ServerName>\n")
		fmt.Printf("\t$(ENV_HOME)/Proxyenv.yaml에 있는 Server name\n")
		fmt.Printf("   <debug mode>\n")
		fmt.Printf("\tfile, console, all\n")
		os.Exit(0)
	}

	var err error
	err = Log.InitNetLogout(os.Args[0], os.Args[1])
	if err != nil{
		fmt.Printf("InitNetLogout() error : %v\n", err)
		os.Exit(0)
	}
	defer Log.Close()
	
	api.UsingMaxProcess()

	Log.Infor("Start %s - %s. Network logout version", os.Args[0], os.Args[1])

	envfile := os.Getenv("ENV_HOME")
	if envfile == ""{
		Log.Error("환경변수에 ENV_HOME이 없습니다.")
		os.Exit(0)
	}
	envfile = envfile + "/" + "Proxyenv.yaml"
	ProcInfor, err = api.NewProxyInfor(envfile, os.Args[1])
	if err != nil{
		Log.Error("NewProxyInfor(%s, %s) : %v", envfile, os.Args[1])
		os.Exit(0)
	}
	if ProcInfor.ProxyType != "REQ/REP" {
		Log.Error("%s의 Proxy Type이 다릅니다 REQ/REP ≠  %s", 
			os.Args[1],
			ProcInfor.ProxyType)
		os.Exit(0)
	}
	Log.Infor("🔵 ProxyEnv : %s", envfile)
	// ROUTER 설정
	router, err := zmq.NewSocket(zmq.ROUTER)
	if err != nil{
		Log.Error("NewSocket(ROUTER) : %v", err)
		os.Exit(0)
	}
	err = router.Bind(ProcInfor.FrontEndPoint)
	if err != nil{
		Log.Error("router.Bind(%s) : %v", ProcInfor.FrontEndPoint, err)
		os.Exit(0)
	}
	rmon, _:=tbzmq.MakeMonitorSocket(os.Args[1], router)
	defer rmon.Close()
	defer router.Close()
	go goMonitor("ROUTER", rmon)

	// DEALER 설정
	dealer, err := zmq.NewSocket(zmq.DEALER)
	if err != nil{
		Log.Error("NewSocket(DEALER) : %v", err)
		os.Exit(0)
	}
	err = dealer.Bind(ProcInfor.BackEndPoint)
	if err != nil{
		Log.Error("dealer.Bind(%s) : %v", ProcInfor.BackEndPoint, err)
		os.Exit(0)
	}
	dmon, _:=tbzmq.MakeMonitorSocket(os.Args[1], dealer)
	defer dmon.Close()
	defer dealer.Close()
	go goMonitor("DEALER",dmon)
	Log.Infor("🔵 REQ/REP【%s】Proxy service start! %s ⇄  %s", 
		os.Args[1], ProcInfor.FrontEndPoint, ProcInfor.BackEndPoint)

	// Context 얻기
	ctx, _:= router.Context()

	// inner proxy
	innerF := fmt.Sprintf("inproc://front_%d", os.Getpid()) // PUB
    innerB := fmt.Sprintf("inproc://back_%d", os.Getpid()) // SUB
	err = tbzmq.RunPubSubProxy(ctx, innerF, innerB)
	if err != nil{
        Log.Error("RubPubSubProxy() : %v", err)
        os.Exit(0)
    }
	// Capture를 위한 go routine
	go goCapture(ctx, innerB, ProcInfor.CaptureEndPoint)

    // Capture 전송을 위한 socket
    capture,_ := zmq.NewSocket(zmq.PUB)
    capture.Connect(innerF)
    time.Sleep(100 * time.Microsecond) // goCapture() 초기화 대기

	// Control 설정
	go goControl(ctx, ProcInfor.ControlEndPoint)

	// Poller 설정(router and dealer)
	poller := zmq.NewPoller()
	poller.Add(router, zmq.POLLIN)
	poller.Add(dealer, zmq.POLLIN)

	Log.Infor("%+v", *ProcInfor)
	// Main Loop 
    for {
		if ProcInfor.Pause {time.Sleep(1000 * time.Millisecond);continue}

        polls, _ := poller.Poll(1000*time.Millisecond)
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
