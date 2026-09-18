// Matching System control process
//
//                      
//  "AFTER" + idx        N-Number       "ORDCTL"
// 【goAfterMatch】        ↑            "CONTROL"(SUB)
//             ↘  ┌────────────────────┐       ┌─────────────┐
// tcp-connect    │  🔴 goMatchingCtl  │  ⇆    │  EXTRAINFOR │
//             ↗  └───────────┬────────┘       └─────────────┘
// 【goMatching】             │ "CONTROL" + "MATCHINGCTL"(PUB)
// "MATCHING"+idx +Symbol    ↓│      G   S
//                    ┌─────────┬────┬───┐   
//                    │ BTC.USDT│  1 │ 1 │  G : Matching Group
//                    ├─────────┼────┼───┤  S (상태) 1 : 정상
//                    │ ETH.USDT│  2 │ 0 │           0 : 주문중지
//                    ├─────────┴────┴───┘           3 : 매도만 가능
//                    └─────────┴────┴───┘
//                    ~/Order/ORDERCOND.MAP 
//
// ①  N-Number
//    Running 중인 goAfterMatch의 개수
//    goMatching과 goAfterMatch 사이의 Topic을 결정하기 위한 번호로 
//   사용되며 이는 부하 분산의 효과와 같은 User Id는 최대한 같은 goAfterMatch
//   가 처리를 담당하여 Data의 일관성 및 Data의 역전(주문 보다 먼저인 체결)
//   을 막기 위함
// ②  Topic의 결정 방법 
//    "AFTER"+ UserId % N-Number
//
package main
import (
	zmq "github.com/pebbe/zmq4"
	Log "centolib/logout"
	ctl "centolib/ctlzmq"
	Mb "centolib/matchingboard"
	"centolib/api"
	"strings"
	"strconv"
	"fmt"
	"time"
	"os"
)
var (
	Nvalue int 
	MatchingCli map[int]int  // fd : group Number
	AfterCli    map[int]int  // fd : Index
)

// Using은 프로세스 사용법을 표준 출력에 출력한다.
func Using(){
	fmt.Printf("Matching System Control  Process\n")
	fmt.Printf("USING) %s <debug mode>\n", os.Args[0])
	fmt.Printf("  <debug mode>\n")
	fmt.Printf("\tfile, console, all\n")
}
// 신규상장 처리 
func newListProcess(symbol string)error{

	// 신규 코인의 주문정보를 만든다
	err := MakeNewCoinOrder(symbol)
	if err != nil{return err;}

	ordName, err := api.GetPlatformEnv("ORDER_LIST")
	if err != nil{ return err;}
	Log.Infor("ORDER_LIST file name : %s", ordName)	

	err = Mb.AddNewCoin(ordName, symbol)
	if err != nil{ return err;}

	Log.Infor("신규상장코인(%s) 정보 생성 완료", symbol)

	// Pause 전송
	err = SendPauseToOrderPass()
	if err != nil{ return err;}

	Log.Infor("ORDERPASS : PAUSE 전송")

	err = LoadCoin()
	if err != nil{ return err;}
	err = SendRunToOrderPass()
	if err != nil{ return err;}
	Log.Infor("ORDERPASS : RESUME 전송")

	ctl.SendControlMessage("goRecvOrder", "REATTACH")
	ctl.SendControlMessage("goMatching", "RELOAD")
	ctl.SendControlMessage("goAfterMatch", "RECONNECT")
	ctl.SendControlMessage("goMatching",  "REPORT")

	Log.Infor("Control Message : to [goRecvOrder]  cmd[REATTACH]")
	Log.Infor("Control Message : to [goMatching]   cmd[RECONNECT]")
	Log.Infor("Control Message : to [goMatching]   cmd[RELOAD]")
	Log.Infor("Control Message : to [goAfterMatch] cmd[RECONNECT]")

	Nvalue = 0
	return nil
}
////////// Control ////////////////////////////////////
func controlHandler(ent *api.GoEvent, socket *zmq.Socket){
	msg, err := socket.RecvMessage(0)
	if err != nil{
		Log.Error("Control msg : %v", err)
		return
	}
	if len(msg) < 2 { return ;}
Log.Infor("%v", msg)
	//
	// msg[0] : "ORDCTL" or "CONTROL"
	// CONTROL인경우 
	//     msg[1] : from 
	//     msg[2] : to
	//     msg[3] : commend
	// ORDCTL인 경우
	//     msg[1] : MESSAGE
	//
	switch msg[0] {
	case "CONTROL" : // 내부 Control Mesasge
		if len(msg) < 4 {
			Log.Debug("Control Message 오류 : %v", msg)
			return
		}
		if msg[2] != "goMatchingCtl"{ 
			Log.Debug("내것아님 : %v", msg)
			return
		}

	case "ORDCTL" : // 외부에서 주문 Control Message
		// NEWLIST:BTC.USDT
		parts := strings.Split(msg[1], ":")
		if parts[0] != "NEWLIST" { 
			Log.Warn("NEWLIST가 아닙니다. [%s]", parts[0])
			return
		}
		if len(parts) < 2{
			Log.Warn("ORDCTL Message 오류! msg:%s", msg[1])
			return
		}
		err := newListProcess(parts[1])
		if err != nil{
			Log.Error("%v", err)
			ent.Run = false
			return
		}
	default :
		return
	}
	return
}
/////////////////// matching Client Handler ///////////////////////
func matchingClientHandler(ent *api.GoEvent, fd int){
	size, msg, err := api.ReadDataStr4(fd)
	if err != nil{
		Log.Infor("Matching Client Read : %v", err)
		return
	}
	if size <= 0{
		Log.Infor("goMatching Connection Close")
		ent.DelTcpClient(fd)
		// MatchingCli 삭제
		group := MatchingCli[fd]
		// 주문 상태 변경( 주문 불가)
		ChangeGroupStatus(group, 0)
		delete(MatchingCli, fd)
		return
	}
	// msg->1:ETC.USDT:BTC.USDT
	parts := strings.Split(string(msg), ":")
	Log.Infor("From goMatching Cleint :[%v]", parts)
	group, err := strconv.Atoi(parts[0])
	if err != nil{
		Log.Infor("goMatching id 오류. [%s]는 group index로 변환 불가",
					parts[0])
		return
	}
	// Group내 기존 COIN을 모두 주문불가로 변경
	ChangeGroupStatus(group, 0)
	// 신규List의 Coin 상태 변경 
	parts = parts[1:]
	for _, coin := range(parts){
		err := ChangeOrderStatus(coin, group, 1)
		if err != nil{
			Log.Error("%v", err)
			break
		}
	}
	// MatchingCli 등록
	MatchingCli[fd] = group
	return
}
/////////////////// Matcing Connect Handler ////////////////////////
func matchingConnectHandler(ent *api.GoEvent, fd int){
	newfd, ip, err := api.WaitConnectRaw(fd)
	if err != nil{
		Log.Error("goMatching Connect : %v", err)
		return
	}
	Log.Infor("goMatching. connect from [%s]. fd[%d] newfd[%d]", ip, fd, newfd)
	ent.AddTcpClient(newfd, matchingClientHandler)
	packet := fmt.Sprintf("NValue:%d", Nvalue)
	sz, err := api.WriteDataStr4(newfd, packet)		
	if sz <= 0 || err != nil{
		Log.Error("NValue전송 오류 Size[%d]. newfd[%d]:%v ", newfd, sz, err)
		ent.Run = false
	}else{
		Log.Infor("NValue 전송. %s", packet)
	}
	return
}
// SendNvalueToMatching은 실행 중인 모든 goMatching 프로세스에 현재 N-Value를 전송한다.
func SendNvalueToMatching(){
	// goMatching으로 Nvalue 전송
	packet := fmt.Sprintf("NValue:%d", Nvalue)
	for mFd, group :=range MatchingCli{
		sz, err := api.WriteDataStr4(mFd, packet)		
		if sz <=0 || err != nil{
			Log.Error("SZ[%d] NValue 전송 오류. fd[%d] group[%d] : %v", 
					sz, mFd, group, err)
			return
		}
	}
	return
}
// SendTopicToAfterMatch는 실행 중인 모든 goAfterMatch 프로세스에 수신 TOPIC을 전송한다.
func SendTopicToAfterMatch(){
	i := 0
	for mFd, _ := range AfterCli {
		msg := fmt.Sprintf("AFTER%02d", i)
		_, err := api.WriteDataStr4(mFd, msg)
		if err != nil{
			Log.Error("Topic 전송 실패. mFd[%d] : %v", mFd, err)
			continue
		}
		Log.Infor("mFD[%d] - Topic(%s)", mFd, msg)
		i++
	}
}
/////////////////// AfterMatching Client Handler //////////////////
func afterMatchClientHandler(ent *api.GoEvent, fd int){
	size, msg, err := api.ReadDataStr4(fd)
	if err != nil{
		Log.Infor("goAfterMatch Client Read : %v", err)
		return
	}
	if size <= 0{ // Connection Closed
		ent.DelTcpClient(fd)
		delete(AfterCli, fd)
		// goAfterMatch가 하나도 없으면 ORDERPASS PAUSE
		Nvalue = len(AfterCli)
		if Nvalue == 0{
			err := SendPauseToOrderPass()
			if err != nil{
				Log.Error("SendPauseToOrderPass() : %v", err)
				ent.Run = false  // 종료
				return
			}
			Log.Warn("goAfterMatch가 모두 종료 상태입니다.ORDERPASS-PAUSE")
			return
		}
		Log.Warn("goAfterMatch Connection Close.NValue(%d)", Nvalue)
		SendNvalueToMatching()  // goMatching에 Nvalue 전송
		SendTopicToAfterMatch() // goAfterMatch에 Topic 전송
		return
	}
	
	// msg->AFTERMATCH:01
	parts := strings.Split(string(msg), ":")
	index, err := strconv.Atoi(parts[1])
	if err != nil{
		Log.Infor("goAfterMatch id 오류. [%s]는 index로 변환 불가",
					parts[1])
		return
	}
	// Nvalue 변경 
	AfterCli[fd] = index
	Nvalue = len(AfterCli)
	Log.Infor("From goAfterMatch Cleint :[%v] Nvalue[%d]", parts, Nvalue)

	SendNvalueToMatching()  // goMatching에 Nvalue 전송
	SendTopicToAfterMatch() // goAfterMatch에 Topic 전송

	err = SendRunToOrderPass()
	if err != nil{
		Log.Error("SendRunToOrderPass() : %v", err)
		ent.Run = false  // 종료
		return
	}
	Log.Infor("NValue(%d) 전송 OK. ORDERPASS RUN 전송 OK.", Nvalue)
	return
}
/////////////////// AfterMatching Connect Handler //////////////////
func afterMatchConnectHandler(ent *api.GoEvent, fd int){
	newfd, ip, err := api.WaitConnectRaw(fd)
	if err != nil{
		Log.Error("goAfterMatch Connect : %v", err)
		return
	}
	Log.Infor("goAfterMatch. connect from [%s]", ip)
	ent.AddTcpClient(newfd, afterMatchClientHandler)
	return
}
// SendPauseToOrderPass는 ORDERPASS Proxy에 PAUSE 메시지를 전송하여 주문 수신을 일시 중단시킨다.
func SendPauseToOrderPass()error{
	file, err:=api.GetPlatformEnv("PROXYENV_YAML")	
	if err != nil{return err;}
	val, err := api.GetProxyInfor(file, "ORDERPASS", "control")
	if err != nil{return err;}
	socket, err := zmq.NewSocket(zmq.REQ)
	if err != nil{return err; }
	defer socket.Close()

	err = socket.Connect(val)
	if err != nil{return err;}

	_, err = socket.SendMessage("PAUSE")
	if err != nil{return err;}
	msg, _:= socket.RecvMessage(0)
	Log.Warn("ORDERPASS에 PAUSE 전송 후 수신 : %s", string(msg[0]))
	return nil
}

// SendRunToOrderPass는 ORDERPASS Proxy에 RUN 메시지를 전송하여 주문 수신을 재개시킨다.
func SendRunToOrderPass()error{
	file, err:=api.GetPlatformEnv("PROXYENV_YAML")	
	if err != nil{return err;}
	val, err := api.GetProxyInfor(file, "ORDERPASS", "control")
	if err != nil{return err;}
	socket, err := zmq.NewSocket(zmq.REQ)
	if err != nil{return err; }
	defer socket.Close()

	err = socket.Connect(val)
	if err != nil{return err;}

	_, err = socket.SendMessage("RUN")
	if err != nil{return err;}
	msg, _:= socket.RecvMessage(0)
	Log.Warn("ORDERPASS에 RUN 전송 후 수신 : %s", string(msg[0]))

	return nil
}
/////////MAIN FUNCTION ////////////////////////////////////
func main(){
	if len(os.Args) != 2{
		Using()
		os.Exit(0)
	}

    switch os.Args[1] {
        case "file" :
        Log.InitLogger("", Log.FILE_MODE);
        case "console" :
        Log.InitLogger("", Log.CONSOLE_MODE);
        case "all" :
        Log.InitLogger("", Log.ALL_MODE);
        default :
        Log.InitLogger("", Log.CONSOLE_MODE);
    }
    defer Log.Close()

	// 중복실행 Check
	run, err := api.IsRunning("MATCHINGCTL")
	if err != nil{
		Log.Error("IsRunning() : %v", err)
		os.Exit(0)
	}
	if run {
		Log.Error("MATCHINGCTL은 실행 중 입니다.")
		os.Exit(0)
	}

	// Max Process사용
	api.UsingMaxProcess()
	
	// Event 초기화
	event := api.GoEvent{}
	err = event.InitGoEvent()
	if err != nil{
		Log.Error("InitGoEvent : %v", err)
		os.Exit(0)	
	}

	// 
	// Control Message
	// SetSubscribe("CONTROL") 포함
	context, _:= zmq.NewContext()
	ctlSub,err := ctl.InitControlMessage(context, "goMatchingCtl")
	if ctlSub == nil{ os.Exit(0) }
	defer ctl.CloseControlMessage()
	defer ctlSub.Close()
	ctlSub.SetSubscribe("ORDCTL")
	time.Sleep(1000 * time.Millisecond) // for good zmq connection	

	event.AddZmqClient(ctlSub, controlHandler)

	// goMatching을 위한 TCP 준비
    val, err := api.GetPlatformEnv("MATCHING_STREAM")
    if err != nil{
        Log.Error("%v", err)
        os.Exit(0)
    }
	sfd, err := api.OpenTcpServerRaw(val)
	if err != nil{
		Log.Error("%v", err)
		os.Exit(0)
	}
	event.AddTcpClient(sfd, matchingConnectHandler)
	Log.Infor("TCP_infor for goMatching : %s", val)

	// AfterMatching을 위한 TCP 준비
    val, err = api.GetPlatformEnv("AFTERMAT_STREAM")
    if err != nil{
        Log.Error("%v", err)
        os.Exit(0)
    }
	sfd, err = api.OpenTcpServerRaw(val)
	event.AddTcpClient(sfd, afterMatchConnectHandler)
	Log.Infor("TCP_infor for goAfterMatch : %s", val)

	err = LoadCoin()
	if err != nil{
		Log.Error("LoadCoin() : %v", err)
		os.Exit(0)
	}
	MatchingCli = make(map[int]int)	
	AfterCli    = make(map[int]int)
	Nvalue      = 0

	// ORDERPASS Proxy PAUSE 전송
	err = SendPauseToOrderPass()	
	if err != nil{
		Log.Error("SendPasueToOrderPass() : %v", err)
		os.Exit(0)
	}
	ctl.SendControlMessage("goRecvOrder", "REATTACH")
	ctl.SendControlMessage("goMatching",  "RECONNECT")
	ctl.SendControlMessage("goAfterMatch", "RECONNECT")
	Log.Infor("Control Message : to [goRecvOrder]  cmd[REATTACH]")
	Log.Infor("Control Message : to [goMatching]   cmd[RECONNECT]")
	Log.Infor("Control Message : to [goAfterMatch] cmd[RECONNECT]")

	Log.Infor("Matching System Control  Process Start- ")
	event.GoEventLoop() // RUN
}
