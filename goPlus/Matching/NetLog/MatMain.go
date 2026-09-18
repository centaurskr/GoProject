// Matching Process 
// 
// ①  FDS,SISE,HOGA,CHE를 go function으로 하는 이유
//    해당 PUB socket을 gofCoinMatching들이 공유함에 따라 경쟁조건이 발생할 
//    가능성이 있음, 내부 Proxy를 통해 각 go function으로 일원화 함
//
//                                   ┌───────────────┐
//   【PUB/SUB Proxy】               │ 🔴 goMatching │  ⇆  control Message
//            🡇                      └───────────────┘
//     ┌────────────┐  "CHANGEDATE" + yyyymmdd   🡇 
//     │DO┌─────────┴──┐      🡄       ┌─────────────┐
//     └──┤ ET┌────────┴──┐   🡆       │ Inner Proxy │
//        └───┤  BTC.USDT │           └─────────────┘
//            └───────────┘  "IPCSEQ"           🡅   IPCCURR                
//      (gofCoinMatching)    "IPCCHE"   【gofAfterFunction】🡆 "AFTER??"
//                           "IPCMKDONE"                       DBCHE
//                           "IPCCANCEL"                       DBCANCEL
//                           "IPCERR"                          DBMKDONE
//                                                             DBERR
//                           "IPCFDS"   【gofFdsFunction  】🡆 
//                                                             
//                           "IPCCURR"  【gofSiseFunction 】🡆  "[COIN]"
//                           "IPCHOGA"
//                            
package main
import (
	"centolib/api"
	Mb "centolib/matchingboard"
	Log "centolib/netlog"
	tbzmq "centolib/tbzmq"
	ctl "centolib/ctlzmq"
	"os"
	"strconv"
	"slices"
	"strings"
	"fmt"
	"runtime"
	zmq "github.com/pebbe/zmq4"
	"github.com/bwmarrin/snowflake"
	"time"
)
var (
	OrdPoint string
	// InnerProxy Front(PUB), InnerBack(SUB), control, capture point
	InnerFront, InnerBack, CtlPoint, CapPoint string
	MatchingCtlFd int // goMatchingCtl Tcp fd
	Today  int   // yyyymmdd
	Index  int   
	// AfterMatching으로 전송하는 TOPIC설정 변수(MatchingCtl에서 수신함)
	Nvalue int   
	SnowNode *snowflake.Node // 체결Unique Key생성을 위한...
	ProxyYaml string
	CoinList  []string
	Context  *zmq.Context
	InnerPub *zmq.Socket
)

// -. 일변경 check go Routine
// 일자 변경시 innerProxy를 통해 
// 모든 Matching go Routine에게 주문번호 초기화 목적으로 
// CHNAGEDATA Message를 전송한다
func goCheckDate(ctx *zmq.Context, ifront string){
	// Inner Control socket (일자변경 Signal)
	socket,_ := ctx.NewSocket(zmq.PUB)
	socket.Connect(ifront)
	defer socket.Close()

	time.Sleep(1000 * time.Millisecond) // for good zmq connection

	Log.Infor("Start 일변경 go Routine. Today[%d]", Today)
	ticker := time.NewTicker(10 * time.Second)
	for{
		select{
		case t := <-ticker.C:
			dateStr := t.Format("20060102")
			today,_ := strconv.Atoi(dateStr)
			if today != Today{
				Log.Infor("일자변경  %d 🡆 %d", Today, today)
				Today = today
				socket.SendMessage("CHANGEDATE", dateStr)
				Log.Debug("CHANGEDATE 전송 : %s", dateStr)
			}
		//default : 있으면 CPU 100% 올라감

		} // Select
	} // For Loop

}
// RunMemeGoRoutine은 환경 파일의 idx에 해당하는 코인 목록을 읽어 매칭 고루틴을 관리한다.
// 신규 코인은 고루틴을 시작하고, 제거된 코인에는 STOP 메시지를 전송한다.
func RunMemeGoRoutine(idx int)error{
	fileName, err := api.GetPlatformEnv("ORDER_LIST")
	if err != nil{
		return err
	}
	// 신규 List읽기
	coinList, err := Mb.GetCoinListByIndex(fileName, idx)
	if err != nil{return err;}

	Log.Debug("CoinList Index[%d]: (R:%d)%v, (OLD:%d):%v", idx, 
		len(coinList), coinList, len(CoinList), CoinList)

	// Coin별 매매 Go routine 생성
	if len(CoinList) == 0{ // 신규실행인 경우
		for _, coin:=range(coinList){
			resultChan := make(chan error)
			go gofCoinMatching(Context, coin, resultChan) // memefunc.go
			if err:= <-resultChan; err != nil{
				Log.Error("%s-실행 실패 : %v", coin, err)
				os.Exit(0)
			}
			Log.Infor("%s 신규 등록, gofCoinMatching() 실행", coin)
		}
		// 신규 List 치환
		CoinList = coinList
		return nil
	}
	// 수행중 변경인 경우
	// 기존의 CoinList에 있는 coin 을 
	// newList로 이동 후 CoinList에서 삭제
	// 최후 CoinList에는 삭제된 coin들만 남아 있게 됨
	// CoinList에 남아 있는 것은 "STOP" Message를 전송하여 종료하게 함
	Log.Infor("수행중 RELOAD ------")
	for _, coin := range coinList{
		if slices.Contains(CoinList, coin){ // OLD에 있다
			// OLD List 에서 삭제
			CoinList = slices.DeleteFunc(CoinList, func(c string) bool{
				return c == coin
			})
			Log.Infor("%s 있음. 유지", coin)
		}else{ // OLD에 없다
			// 신규 실행
			resultChan := make(chan error)
			go gofCoinMatching(Context, coin, resultChan) // memefunc.go
			if err:= <-resultChan; err != nil{
				Log.Error("%s-실행 실패 : %v", coin, err)
				os.Exit(0)
			}
			Log.Infor("%s 신규 등록, gofCoinMatching() 실행", coin)
		}
	}
Log.Debug("정리 후 OLD List : %v", CoinList)
	// OLD에 남아있는 matching go routine 종료
	for _, coin := range CoinList{
		InnerPub.SendMessage("STOP", coin)	
	}

	// 신규 List로 치환
	CoinList = coinList	
	return nil
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

// MatchingCtlHandler는 goMatchingCtl TCP 연결에서 N-Value 등 제어 정보를 수신하여 처리한다.
func MatchingCtlHandler(env *api.GoEvent, fd int){
	size, bytes, err := api.ReadDataStr4(fd)
	if err != nil || size <= 0{
		Log.Warn("goMatchingCtl 접속단절 : %v", err)
		env.DelTcpClient(fd)
		MatchingCtlFd = -1
		err = SendPauseToOrderPass()
		if err != nil{
			Log.Error("ORDERPASS-PAUSE 전송 실패 :%v",err)
			env.Run = false
		}
		return
	}
	// N value 수신 "NValue:1"
	parts := strings.Split(string(bytes), ":")
	Nvalue ,err = strconv.Atoi(parts[1])
	if err != nil{
		Log.Error("Nvalue 수신 후 변환 실패.수신[%s] : %v", 
			string(bytes[:]), err)
		return
	}
	Log.Infor("Index [%d] - Nvalue 수신 [%d]", Index, Nvalue)
	return
}
// SendInforToMatchingCtl은 goMatchingCtl에 현재 인덱스와 처리 중인 코인 목록을 전송한다.
func SendInforToMatchingCtl(){
	if MatchingCtlFd < 0{
		Log.Warn("goMatchingCtl과 미 접속 상태입니다.")
		return
	}
	msg := fmt.Sprintf("%d", Index)	
	for _, coin := range(CoinList){
		msg = msg + ":" + coin
	}
	_, err := api.WriteDataStr4(MatchingCtlFd, msg)
	if err != nil{
		Log.Error("goMatchingCtl로 전송실패: %v", err)
		return
	}
	Log.Infor("goMatchingCtl 전송 fd[%d] [%s] ", MatchingCtlFd, msg)
	return
}
// ConnectMatchingCtl은 goMatchingCtl TCP 서버에 접속하여 파일 디스크립터를 반환한다.
func ConnectMatchingCtl()(int, error){
	var endPoint string
	val, err := api.GetPlatformEnv("MATCHING_HOST")
	if err != nil{ return -1, err; }
	endPoint = val
	val, err = api.GetPlatformEnv("MATCHING_STREAM")
	if err != nil{ return -1, err; }
	endPoint = endPoint + val
Log.Debug("MatchingCtl Endpoint[%s]", endPoint)
	fd , err := api.OpenTcpClient(endPoint)
	if err != nil{ return -1, err; }
	return fd, nil
}
// ControlHandler는 EXTRAINFOR Proxy에서 goMatching 대상의 제어 메시지를 수신하여 처리한다.
func ControlHandler(ent *api.GoEvent, socket *zmq.Socket){
msg, err := socket.RecvMessage(0)
if err != nil{return;}
if len(msg) < 4 || msg[0] != "CONTROL" {
	Log.Warn("Control message 오류: %v", msg)
	return
}
// to Process 가 goMatching이 아니고 ALL도 아니면
if msg[2] != "goMatching" && msg[2] != "ALL"{return;}

// msg[0] - "CONTROL"
// msg[1] - from process
// msg[2] - to process
// msg[3] - command

switch msg[2] {
case "goMatching" :
	var err error
	// goMatchingCtl로 부터 "RECONNECT"이면
	if msg[1] == "goMatchingCtl"{
		switch msg[3] {
		case "RECONNECT" :
			Log.Warn("goMatchingCtl RECONNECT Message 수신")
			if MatchingCtlFd > 0{ 
				Log.Infor("goMatchingCtl 이미 접속이 있는 상태. Close한다")
				ent.DelTcpClient(MatchingCtlFd) // syscall.Close() 포함
			}
			MatchingCtlFd, err = ConnectMatchingCtl()
			if err != nil{
				Log.Error("goMatchingCtl과 접속 실패! %v", err)
				return
			}
			Log.Infor("goMatchingCtl과 접속 성공... MatchingCtlFd[%d]" , 
				MatchingCtlFd)
			SendInforToMatchingCtl()
			ent.AddTcpClient(MatchingCtlFd, MatchingCtlHandler)
		case "REPORT" : 
			Log.Warn("goMatchingCtl REPORT Message 수신")
			SendInforToMatchingCtl()
		case "RELOAD" :
			Log.Warn("goMatchingCtl RELOAD Message 수신")
			err := RunMemeGoRoutine(Index)
			if err != nil{
				Log.Error("RunMemeGoRoutine() : %v", err)
				os.Exit(0)
			}
			SendInforToMatchingCtl()
		} // switch msg[3] - goMatchingCtl message 
	}// if msg[1]- From Process(goMatchingCtl)
case "ALL":
} // switch msg[2]-To Process
return
} // Control Handler

func main(){

	if len(os.Args) != 2{
		fmt.Printf("매칭 Process Network log version\n")
		fmt.Printf("Using) goMatching <Index>.\n\tex)goMatching 1\n")
		return
	}
	err := Log.InitNetLogout(os.Args[0], os.Args[1])	
	if err != nil{
		fmt.Printf("InitNetLogout() %v\n", err)
		os.Exit(0)
	}
    defer Log.Close()

	api.UsingMaxProcess()

	// Proxy YAML path
	yamlFile ,err := api.GetPlatformEnv("PROXYENV_YAML")
	if err != nil{
		Log.Error("%v", err)
		os.Exit(0)
	}
	ProxyYaml = yamlFile
	
	// 중복실행 check
	idx, _ := strconv.Atoi(os.Args[1])
	if idx <= 0 {
		Log.Error("Index는 1부터 시작합니다.")
		os.Exit(0)
	}
	lockname := fmt.Sprintf("Matching_%d", idx)
	isrun, err := api.IsRunning(lockname)
	if isrun == true {
		Log.Error("Index %d는 이미 수행중입니다", idx)
		os.Exit(0)
	}

	Index = idx
	// 체결 Unique Key 생성을 위한 Node생성
	SnowNode, _= snowflake.NewNode(int64(Index))

	// 주문 수신 Endpoint	
	// RecvOrder --> (f)PubSubProxy(b) --> 
	val, err := api.GetProxyInfor(ProxyYaml, "BEFOREMATCHING", "back")
	if err != nil{
		Log.Error("BEFOREMATCHING:back : %v", err)
		return
	}
	OrdPoint = val

	// MatchingCtl connect Point
	val, err = api.GetPlatformEnv("MATCHING_STREAM")
	if err != nil{
		Log.Error("MATCHING_STREAM : %v", err)
		return
	}
	CtlPoint   = val

	/// 사용할 Process 개수 설정
    runtime.GOMAXPROCS(runtime.NumCPU() - 1)

	// Inner Proxy 
	pid := os.Getpid()
	InnerFront = fmt.Sprintf("inproc://IPC_FRONT-%d", pid)
	InnerBack  = fmt.Sprintf("inproc://IPC_BACK-%d", pid )
	Log.Infor("Ord (%s), Front(%s) Back(%s) Ctl(%s) ", 
		OrdPoint, InnerFront, InnerBack, CtlPoint)

	Context, err = zmq.NewContext()
	if err != nil{
		Log.Error("zmq/NewContext() : %v", err)
		os.Exit(0)
	}
	defer Context.Term()

	// Inner Proxy
	err = tbzmq.RunPubSubProxy(Context, InnerFront, InnerBack)
	if err != nil{
		Log.Error("Proxy Error : %v", err)
		return
	}

	// 시세 처리 Go routine
	go gofSiseFunction(Context)
	// 매칭 후 처리 Go Routine
	go gofAfterFunction(Context)
	// FDS 처리 Go Routine
	go gofFdsFunction(Context)

	// 오늘
	Today = api.GetNumYyyymmdd()
	Log.Infor("오늘 : %d", Today)

	// 일변경 Check go routine
	go goCheckDate(Context, InnerFront)

	// Event 초기화
	event := &api.GoEvent{}
	err = event.InitGoEvent()
	if err != nil{
		Log.Error("InitGoEvent() : %v", err)
		os.Exit(0)
	}

	// For Inner Pub Message
	InnerPub,err = Context.NewSocket(zmq.PUB)
	if err != nil{
		Log.Error("InnerPub : %v", err)
		os.Exit(0)
	}
	InnerPub.Connect(InnerFront)
	time.Sleep(1000*time.Millisecond) // for good zmq connection

	// 매매를 위한 go routien 실행(종목 당 하나의 go routine)
	err = RunMemeGoRoutine(Index)
	if err != nil{
		Log.Error("RunMemeGoRoutine() : %v", err)
		os.Exit(0)
	}
	// For Control message 
	ctlSocket, err := ctl.InitControlMessage(Context, "goMatching")
	if err != nil{
		Log.Error("InitControlMessage() : %v", err)
		os.Exit(0)
	}
	event.AddZmqClient(ctlSocket, ControlHandler)

	// For goMatchingCtl
	fd, err := ConnectMatchingCtl()
	if err != nil{
		Log.Warn("goMatchingCtl과 접속 실패: %v", err)
		MatchingCtlFd = -1
		// 접속 실패인 경우도 실행한다
	}else{
		MatchingCtlFd = fd
		event.AddTcpClient(MatchingCtlFd, MatchingCtlHandler)
		SendInforToMatchingCtl()
		Log.Infor("goMatchingCtl 접속 성공. Fd[%d], fd[%d]", MatchingCtlFd,fd)
	}
	Log.Infor("시작! Index [%d] CoinList : %v MatchingCtlFd[%d]", 
			Index, CoinList, MatchingCtlFd)
	event.GoEventLoop()
}
