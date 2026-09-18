// PUB/SUB Proxy Network Log version
//
//       ┌──────────────────────────────┐
// PUB   │  (XSUB) ReqRepProxy (XPUB)   │  SUB
//       │                control(REP)  │            
//       └──────────────────────────────┘
//                            REQ
// PubSubProxy의 Cature는 없음.
// 필요시 SUB Endpoint에서 수신....
//
// 
package main

import (
	"fmt"
	"os"
	"time"
	"centolib/tbzmq"
	"centolib/api"
	Log "centolib/netlog"
	zmq "github.com/pebbe/zmq4"
	"encoding/json"
)

var ProcInfor *api.ProxyInfor

func goMonitor(name string, s *zmq.Socket){
    Log.Infor("%s Socket Monitor Start", name)
    for{
        event,addr, _, _ := s.RecvEvent(0)
        switch event{
            case zmq.EVENT_ACCEPTED :
                if name == "FRONT"{
                    ProcInfor.FrontCount ++
                    Log.Infor("Front(PUB) - Connect from %s, Count(%d)",
                        addr,ProcInfor.FrontCount)
                }else if name == "BACK"{
                    ProcInfor.BackCount ++
                    Log.Infor("Back(SUB)- Connect from %s, Count(%d)",
                         addr, ProcInfor.BackCount)
                }
            case  zmq.EVENT_DISCONNECTED:
                if name == "FRONT"{
                    ProcInfor.FrontCount --
                    Log.Infor("Front(PUB)- Disconnect from %s, Count(%d)",
                        addr,ProcInfor.FrontCount)
                }else if name == "BACK"{
                    ProcInfor.BackCount --
                    Log.Infor("Back(SUB)- Disconnect from %s, Count(%d)",
                         addr, ProcInfor.BackCount)
                }
        } // Switch
        Log.Infor("Connection Count: Front-PUB(%d) Back-SUB(%d)",
            ProcInfor.FrontCount, 
			ProcInfor.BackCount) 
    } // For Loop
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
				jsonbyte,_ := json.Marshal(*ProcInfor)
				ctlsocket.SendMessage(string(jsonbyte))
			case "PAUSE" : 
				ProcInfor.Pause = true
			case "RUN" : 
				ProcInfor.Pause = false
			default :
				msg := fmt.Sprintf("Proxy command : INFOR, PAUSE, RUN")
				ctlsocket.SendMessage(msg)
		}
	}
}
///////////// Main Function //////////////////////////////////////////////
func main(){
	if len(os.Args) != 2{
		fmt.Printf("⛔ PUB/SUB Proxy\n")
		fmt.Printf("   Using) %s <SERVERNAME>\n", os.Args[0])
		fmt.Printf("   <ServerName>\n")
		fmt.Printf("\t$(ENV_HOME)/Proxyenv.yaml에 있는 Server name\n")
		os.Exit(0)
	}
	var err error
	err = Log.InitNetLogout(os.Args[0], os.Args[1]);
	if err != nil{
		fmt.Printf("InitNetLogout() : %v\n", err)
		os.Exit(0)
	}
	defer Log.Close()

	Log.Infor("START : %s %s. Network logout version", os.Args[0], os.Args[1])
	api.UsingMaxProcess()

	envfile := os.Getenv("ENV_HOME")
	if envfile == ""{
		Log.Error("환경변수에 ENV_HOME이 없습니다.")
		os.Exit(0)
	}
	envfile = envfile + "/" + "Proxyenv.yaml"
	ProcInfor, err = api.NewProxyInfor(envfile, os.Args[1])
	if err != nil{
		Log.Error("NewProxy(%s, %s) : %v", envfile, os.Args[1])
		os.Exit(0)
	}
	if ProcInfor.ProxyType != "PUB/SUB" {
		Log.Error("%s의 Proxy Type이 다릅니다 PUB/SUB ≠  %s", 
			os.Args[1],
			ProcInfor.ProxyType)
		os.Exit(0)
	}
	Log.Infor("🔵 ProxyEnv : %s", envfile)
	// FRONT 설정
	//front, err := zmq.NewSocket(zmq.XSUB)
	front, err := zmq.NewSocket(zmq.SUB)
	if err != nil{
		Log.Error("NewSocket(FRONT) : %v", err)
		os.Exit(0)
	}
	err = front.Bind(ProcInfor.FrontEndPoint)
	if err != nil{
		Log.Error("front.Bind(%s) : %v", ProcInfor.FrontEndPoint, err)
		os.Exit(0)
	}
	fmon, _:=tbzmq.MakeMonitorSocket(os.Args[1], front)
	defer fmon.Close()
	defer front.Close()
	front.SetSubscribe("")
	go goMonitor("FRONT", fmon)

	// BACK 설정
	//back, err := zmq.NewSocket(zmq.XPUB)
	back, err := zmq.NewSocket(zmq.PUB)
	if err != nil{
		Log.Error("NewSocket(BACK) : %v", err)
		os.Exit(0)
	}
	err = back.Bind(ProcInfor.BackEndPoint)
	if err != nil{
		Log.Error("back.Bind(%s) : %v", ProcInfor.BackEndPoint, err)
		os.Exit(0)
	}
	bmon, _:=tbzmq.MakeMonitorSocket(os.Args[1], back)
	defer bmon.Close()
	defer back.Close()
	go goMonitor("BACK",bmon)

	Log.Infor("🔵 PUB/SUB【%s】Proxy service start! %s 🡆  %s", 
		os.Args[1], ProcInfor.FrontEndPoint, ProcInfor.BackEndPoint)

	// Context 얻기
	ctx, _:= back.Context()

	// Control 설정
	go goControl(ctx, ProcInfor.ControlEndPoint)

	Log.Infor("%+v", *ProcInfor)

	poller := zmq.NewPoller()
	poller.Add(front, zmq.POLLIN)
	for {
		if ProcInfor.Pause == true{
			time.Sleep(500 * time.Millisecond)
			continue
		}
		polls , _:= poller.Poll(1000 * time.Millisecond)
		for _, polled := range polls{
			// PAUSE인경우 읽지 않고 continue
			switch s := polled.Socket;s{
			case front :
				msg, err := front.RecvMessageBytes(0)
				if err != nil{
					Log.Error("front.RecvMessage() : %v", err)
					os.Exit(0)
				}

				ProcInfor.ReadCount ++

				if ProcInfor.BackCount == 0 {break;}

				_, err = back.SendMessage(msg)
				if err != nil{
					Log.Error("back.SendMessage : %v\n\t%v", msg, err)
					os.Exit(0)
				}
				ProcInfor.WriteCount ++
			} // Swtich
		} // range polls
	} // Main Loop
}
