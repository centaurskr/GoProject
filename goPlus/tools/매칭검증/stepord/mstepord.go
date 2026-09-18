package main
import (
	"fmt"
	"os"
	Mb "centolib/matchingboard"
	"encoding/json"
	zmq "github.com/pebbe/zmq4"
	"math/rand"
	"time"
)
func getInput(msg string)(string, bool){
	var input string

	fmt.Printf("%s ", msg)
	fmt.Scanln(&input)
	if input == "Q" || input == "q" {
		return input, false
	}
	return input, true
}
func main() {
	reqpoint := "tcp://127.0.0.1:31001"
	buysell, fg := getInput("매수:B, 매도:S - ")
	if fg == false {os.Exit(0);}
	price, fg := getInput("주문가격 ")
	if fg == false {os.Exit(0);}
	qty, fg := getInput("주문수량 ")
	if fg == false {os.Exit(0);}
	rand.Seed(time.Now().UnixNano())
	idx := rand.Intn(10)
	uid := fmt.Sprintf("%d", 42280 + idx)
	fmt.Printf("UID : %s\n", uid)

	packet := Mb.OrderJson{
		Tr       : "ORDER",
		Symbol   : "BTC.USDT",
		Channel  : "CN03",
		ItemCode : "441",
		MkCode   : "2965",    // 시장 코드 "0"
    	//UserId   : uid,       // 사용자 ID 
    	UserId   : "42281",       // 사용자 ID 
    	UserType : "00",      // 개법인구분("00":개인,...)
    	CoinCode : "2",       // Coin Code "2"
    	OrdType  : "M",       // 1 (L:지정가, M:시장가)
    	BuySell  : buysell,   // "S"(S:매도, B:매수)
    	Price    : price,     // 주문 가격  "89000.00"
    	Qty      : qty,       // 주문 수량 "0.001"
    	MakeRate : "0.0001",  // Maker 수수료 율"0.0001"
    	TakeRate : "0.0007",  // Taker 수수료 율"0.0007"
    	CallUnit : "0.01",    // 호가단위
    	MinTrade : "0.0001",  // 최소주문금액
    	//OrgNum   : " string `json:"ORGSEQ"`    // 원주문번호
		//OrgNum   : "2025022600000001001",
    	Cancel   : "0"     ,  // "0" "1"
    	UserIp   : "222.107.37.103",
    	Uuid     : "f74fef70-9124-4342-a44a-c2cd4df37922",
    	UserField:"USERFIELD", // 사용자 정의 필드 ""
	}
	stream,_ := json.Marshal(packet)
	fmt.Printf("REQ Point : %s\n", reqpoint)
	req,_ := zmq.NewSocket(zmq.REQ)
	req.Connect(reqpoint)
	req.SendMessage(stream)

	msg,_ := req.RecvMessage(0)

	fmt.Printf("REP : %s\n", msg)
}
