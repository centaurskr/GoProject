///
/// Matching 이후 DB 정리 Process
/// file    afterMatch.go
/// date    2025. 01. 22. (수) 14:04:17 KST
package main

import (
	zmq "github.com/pebbe/zmq4"
    Log "centolib/logout"
    Mb "centolib/matchingboard"
	"encoding/json"
	"fmt"
	"os"
	"unsafe"
	"time"
	"centolib/api"
)
var (
	SubAfter *zmq.Socket
	Queue    *CircularQueue
)
/////////////// Dump정리 //////////////////////////////
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

/////// Write Queue //////////////////////////////////////////////////////////
// DBCHE : maker taker의 체결 DATA 
// DBSEQ : 주문번호 갱신, 주문상태 RG->GO
// 

func DataHandler(ent *api.GoEvent, socket *zmq.Socket){
	msg, err := socket.RecvMessageBytes(0)
	// msg[0] -> AFTER?? (Topic)
	// msg[1] -> DBCHE, DBERR or DBSEQ
	// msg[2] -> []byte data
	if err != nil{ return ;}
	for !Queue.Write(string(msg[1]), msg[2]) {
		time.Sleep(10 * time.Millisecond)
	}
	return 
}
func gofChe(data []byte){
	che := (*Mb.InnerChePacket)(unsafe.Pointer(&data[0]))	
	fmt.Printf("CHE OrdNum[%d] \n", che.OrdNum)
}
func gofSeq(data []byte){
	var seqJson Mb.OrderJson
	err := json.Unmarshal(data, &seqJson)
	if err != nil{
		fmt.Printf("SEQ: ERROR : %v\n", err)
	}else{
		fmt.Printf("SEQ : RegNUm(%s) OrdNum(%s)\n", seqJson.RegNum, seqJson.OrdNum)
	}
}
//// Read Queue ////
/***********************************
func gofReadQueue(queue *CircularQueue){
	for{
		<- queue.dataCh
		for {
			tr, data, ok := queue.Read()
			if !ok { 
				break	
			}
			//fmt.Printf("%s len(%d)\n", tr, len(data))
			switch tr {
				case "DBSEQ" : 
				 	gofSeq(data)
				case "DBCHE" :
				 	gofChe(data)
			}
		}
	}
}
********************************/
//// For queue4.go/////////////////////
func gofReadQueue(queue *CircularQueue){
	for{
		<- queue.readCh
		for {
			//Log.Infor("W/R(%d:%d)", WriteCount, ReadCount)
			qData, ok := queue.Read()
			if !ok { break; }
			for _, pk := range qData {
				switch pk.tr {
				case "DBSEQ" : 
				 	gofSeq(pk.data)
				case "DBCHE" :
				 	gofChe(pk.data)
				}
			}
		}
	}
}
/////////////////////////////////// Main //////////////////////
func main(){
	
    Log.InitLogger("", Log.CONSOLE_MODE);
    defer Log.Close()

	api.UsingMaxProcess()
/*
	WriteCount = 0
	ReadCount  = 0
*/
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

// 임시 
	SubAfter.SetSubscribe("")

	event := &api.GoEvent{}
	err = event.InitGoEvent()
	if err != nil{
		Log.Error("InitGoEvent : %v ", err)	
		os.Exit(0)
	}
	
	//Queue = NewCircularQueue(1024 * 1000)
	Queue = NewCircularQueue(1000)
	//Queue = api.NewCircularQueue(1000)

	go gofReadQueue(Queue)


	event.AddZmqClient(SubAfter, DataHandler)

	event.GoEventLoop()
}
