// Matching 이후 DB 정리 Process
//
//                    ┌─────────────────┐
//                    │ ordertbl,chetbl │          【goMatchingCtl】
//           "DBERR"  └─────────────────┘                  │
//           "DBCHE"          🡅  ② Insert(chetbl)          │ 
//           "DBSEQ"             ③ Update(ordertbl)        │ ✓zmq.STREM connect
//               ①    ┌─────────────────┐ ─────────────────┘
//   【AFTERCHE】🡆    │ 🔴goAfterMatch  │ 🡄  【EXTRAINFOR Proxy】ControlMessage
//                    └─────────────────┘            "goAfterMatch"+"RECONNECT"
//               channel(5000) 🡇      ④  DBCHE->USER?? + DBCHE + (체결DATA)
//                     ┌───────────────┐  DBSEQ->USER?? + DBSEQ + (주문번호)
//                     │ goDataProcess │
//                     └───────────────┘
//                          🡇                  
//                     ┌──────────┐           
//                     │ REALJSON │            시세(BTC.USDT)
//                     │ SISEJSON │            호가(BTC.USDT)
//                     └──────────┘
// ✓ zmq.STREM connect
// AFTERCHE Proxy로 부터 수신할 TOPIC을 goMatchingCtl에서 받기 위한 
// zmq.STREAM(TCP) Connection임.
//
package main

import (
	zmq "github.com/pebbe/zmq4"
    Log "centolib/netlog"
    Mb "centolib/matchingboard"
	ctl "centolib/ctlzmq"
	"centolib/api"
	"strconv"
	"os"
	"time"
	"fmt"
	"database/sql"
)
var (
	Index    int
	MatchingCtlFd int
	PubReal   *zmq.Socket
	SubAfter *zmq.Socket
	DbContext *sql.DB
	SubTopic   string
	Queue     *CircularQueue
)
// Dump는 InnerChePacket의 모든 필드를 표준 출력으로 덤프한다.
func Dump( pk *Mb.InnerChePacket){
	fmt.Printf(`
-----------DBCHE----------------
CheKey     : %d
Symbol     : %s
ItemCode   : %d
MkCode     : %d
CoinCode   : %d
UserId     : %d
OrdType    : %s
MakeTake   : %s
OrdNum     : %d
BuySell    : %s
CheSeq     : %d
Price      : %s
Qty        : %s
RemQty     : %s
Margin     : %s
Fee        : %s
CheAmt     : %s
CheFee     : %s
FeeRate    : %s
CheTime    : %d
CntOrdNum  : %d
ExRate     : %s
UserField  : %s
JSON : %s
`,
    	pk.CheKey           , // 체결 Pair Key   
    	api.CleanString(pk.Symbol[:]), // 종목 명 "BTC.USDT"
    	pk.ItemCode         , // 종목 코드 "4441"
    	pk.MkCode           , // 시장 코드 "0"
    	pk.CoinCode         , // Coin Code "2"
    	pk.UserId           , // User Id
		string(pk.OrdType)  , // 주문종류(L,M)
    	string(pk.MakeTake) , // "M" Maker,  "T" Taker
    	pk.OrdNum           , // 주문번호
    	string(pk.BuySell)  , // 매도수구분 (S:매도, B:매수)
    	pk.CheSeq           , // 체결번호
    	pk.Price.String()   , // 체결가격
    	pk.Qty.String()     , // 체결수량
    	pk.RemQty.String()  , // 주문잔량
    	pk.Margin.String()  , // 주문증거금(변제용)
    	pk.Fee.String()     , // 체결증거금(변제용)
    	pk.CheAmt.String()  , // 체결 금액(체결가격X체결수량)
    	pk.CheFee.String()  , // 체결 수수료(매수/매도 동일)
		pk.FeeRate.String() , // 체결당시 수수료
    	pk.CheTime          , // 채결시각, UinxNano(), 1728372731578396031(19)
    	pk.CntOrdNum        , // 상대 주문번호   
    	pk.ExRate.String()  , // 환율
		api.CleanString(pk.UserField[:]), 
		api.CleanString(pk.OrdPacket[:]),
	)
    

}
// ConnectMatchingCtl은 goMatchingCtl TCP 서버에 접속하여 파일 디스크립터를 반환한다.
func ConnectMatchingCtl()(int, error){
	var endPoint string
	val, err := api.GetPlatformEnv("MATCHING_HOST")
	if err != nil {return -1, err;}
	endPoint = val
	val, err = api.GetPlatformEnv("AFTERMAT_STREAM")
	if err != nil {return -1, err;}
	endPoint = endPoint + val // 192.168.0.12:123123
Log.Debug("MatchingCtl EndPoint : %s", endPoint)	
	fd , err := api.OpenTcpClient(endPoint)
	if err != nil {return -1, err;}
	return fd, nil
}

// ControlHandler는 EXTRAINFOR Proxy에서 수신하는 제어 메시지를 처리한다.
func ControlHandler(ent *api.GoEvent, socket *zmq.Socket){
	msg, err := socket.RecvMessage(0)
	if err != nil{return;}
	// msg[0] - "CONTROL"
	// msg[1] - from Process
	// msg[2] - to Process
	// msg[3] - command
	if msg[0] != "CONTROL" || len(msg) < 4{
		Log.Warn("Control message 오류 : %v", msg)
		return 
	}
	if msg[2] != "goAfterMatch" && msg[2] != "ALL"{return;}
	switch msg[2] {
	case "goAfterMatch" :
		// goMatchingCtl로 부터 "RECONNECT" 이면 
		if msg[1] == "goMatchingCtl" && msg[3] == "RECONNECT"{
			if MatchingCtlFd > 0{
				Log.Warn("goMatchingCtl, 이미 접속 상태. close한다")
				ent.DelTcpClient(MatchingCtlFd)
			}
			MatchingCtlFd, err = ConnectMatchingCtl()
			if err != nil{
				Log.Error("goMatchingCtl과 접속 실패! : %v", err)
				return
			}
			Log.Infor("goMatchingCtl과 재 접속 성공..Fd[%d]", MatchingCtlFd)
			SendInforToMatchingCtl()
			ent.AddTcpClient(MatchingCtlFd, MatchingCtlHandler)
		}
	case "ALL":

	}

	return 
}
// SendInforToMatchingCtl은 goMatchingCtl에 현재 프로세스 인덱스 정보를 전송한다.
func SendInforToMatchingCtl(){
	if MatchingCtlFd < 0{
		Log.Warn("goMatchingCtl과 미접속 상태입니다.")
		return
	}
	msg := fmt.Sprintf("AFTERMATCH:%d", Index)
	_, err := api.WriteDataStr4(MatchingCtlFd, msg)
	if err != nil{
		Log.Error("goMatchingCtl로 전송 실패 : %v", err)
		return
	}
	Log.Infor("goMatchingCtl 전송 fd[%d] [%s]", MatchingCtlFd, msg)
	return
}

// MatchingCtlHandler는 goMatchingCtl TCP 연결에서 수신 토픽을 처리한다.
func MatchingCtlHandler(ent *api.GoEvent, fd int){
	size, bytes, err := api.ReadDataStr4(fd)
	if err != nil || size <= 0{
		Log.Warn("goMatchingCtl 접속 단절 : %v", err)
		ent.DelTcpClient(fd)
		MatchingCtlFd = -1
		return
	}
	// Topic 수신
	// Unsubscribe()
	// Subscribe()
	if SubTopic != "" {
		Log.Infor("🔴Subscribe Topic 변경 [%s] 🡆 [%s]", SubTopic, string(bytes))
		SubAfter.SetUnsubscribe(SubTopic)
		SubTopic = string(bytes)
		SubAfter.SetSubscribe(SubTopic)
	}else if SubTopic == ""{
		SubTopic = string(bytes)
		SubAfter.SetSubscribe(SubTopic)
		Log.Infor("🔴Subscribe Topic 수신 [%s]", SubTopic)
	}else if SubTopic == string(bytes){
		Log.Infor("🔴Subscribe 변경없음.Topic 수신 [%s]", string(bytes))
	}
	return
}
// DataHandler는 AFTERMATCHING Proxy에서 수신한 데이터를 CircularQueue에 기록한다.
// 수신 타입: DBCHE(체결), DBSEQ(주문번호 갱신), DBERR(오류), DBCANCEL(취소), MKDONE(시장가 완료)
func DataHandler(ent *api.GoEvent, socket *zmq.Socket){
	msg, err := socket.RecvMessageBytes(0)
	// msg[0] -> AFTER?? (Topic)
	// msg[1] -> DBCHE, DBERR or DBSEQ
	// msg[2] -> []byte data
	if err != nil{ return ; }
	for !Queue.Write(string(msg[1]), msg[2]){
		Log.Warn("Wait Writing Queue ")
		time.Sleep(10 * time.Microsecond)
	}
	return 
}
/////////////////////////////////// Main //////////////////////
func main(){
	
	if len(os.Args) != 2{
		fmt.Printf("주문 후 처리 (DB 정리, 체결정리등\n")
		fmt.Printf("Using) goAfterMatch <Index>\n")
		os.Exit(0)
	}
	err := Log.InitNetLogout(os.Args[0], os.Args[1])
	if err != nil{
		fmt.Printf("InitNetLogout() : %v\n", err)
	}
    defer Log.Close()

	api.UsingMaxProcess()

    // 중복실행 check
    Index, _ := strconv.Atoi(os.Args[1])
    lockname := fmt.Sprintf("AfterMatch_%d", Index)
    isrun, err := api.IsRunning(lockname)
    if isrun == true {
        Log.Error("Index %d는 이미 수행중입니다", Index)
        return
    }

    // 매칭 후 Data 수신 Proxy 환경설정
    yamlFile ,err := api.GetPlatformEnv("PROXYENV_YAML")
    if err != nil{
        Log.Error("%v", err)
        os.Exit(0)
    }
	val, err := api.GetProxyInfor(yamlFile, "AFTERMATCHING", "back")
	if err != nil{
		Log.Error("%s에 AFTERMATCHING:back이 없습니다", yamlFile)
		os.Exit(0)
	}

	ctx , _ := zmq.NewContext()

	sub, err := ctx.NewSocket(zmq.SUB)
	if err != nil{
		Log.Error("Sub socket : %v", err)
		os.Exit(0)
	}
	SubAfter = sub
	SubAfter.Connect(val)
	defer SubAfter.Close()
	// Subscribe는 MatchingCtl에서 수신 후 설정

	// 처리후 송신 대상 Proxy 환경설정
	val, err = api.GetProxyInfor(yamlFile, "REALJSON", "front")
	if err != nil{
		Log.Error("%s에 REALJSON:front이 없습니다", yamlFile)
		os.Exit(0)
	}
	PubReal, err = ctx.NewSocket(zmq.PUB)
	if err != nil{
		Log.Error("PubReal : %v", err)
		os.Exit(0)
	}
	PubReal.Connect(val)
	defer PubReal.Close()

	// Control Message
    ctlSub,err := ctl.InitControlMessage(ctx, "goAfterMatch")
    if err != nil{ 
		Log.Error("InitControlMessage() : %v", err)
		os.Exit(0) 
	}
    defer ctlSub.Close()
	time.Sleep(100*time.Microsecond) // for good zmq connection

    // Db Connect
    dbyaml,err := api.GetPlatformEnv("DBCONFIG_YAML")
    if err != nil{
        Log.Error("%s", err)
        os.Exit(0)
    }
    DbContext, err = api.ConnectDb(dbyaml)
    if err  != nil{
        Log.Error("ConnectDb(%s):%v", dbyaml, err)
        os.Exit(0)
    }

	MatchingCtlFd ,err = ConnectMatchingCtl()
	if err != nil{
		Log.Warn("ConnectMatchingctl() : %v", err)
		// 접속 실패도 수행한다
	}
	SendInforToMatchingCtl()

	event := &api.GoEvent{}
	err = event.InitGoEvent()
	if err != nil{
		Log.Error("InitGoEvent : %v ", err)	
		os.Exit(0)
	}

	Queue = NewCircularQueue(10000)

	go goDataProcess(Queue)

	Log.Infor("Start Index [%d]", Index)

	event.AddZmqClient(SubAfter, DataHandler)
	event.AddZmqClient(ctlSub, ControlHandler)
	event.AddTcpClient(MatchingCtlFd, MatchingCtlHandler)

	event.GoEventLoop()
}
