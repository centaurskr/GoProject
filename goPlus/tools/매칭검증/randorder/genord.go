// 주문 생성기 (genord.go)
package main
import (
	"math/rand"
	"time"
	Log "centolib/logout"
	"os"
	"encoding/json"
    "io/ioutil"
	"fmt"
	//Bd "centolib/bindecimal"
    zmq "github.com/pebbe/zmq4"
    //Mb "centolib/matchingboard"
    //"centolib/api"
)
const (
	UID_BASE = 42280
)
type Coin struct {
    Symbol     string    `json:"name"`
	ItemCode   int       `json:"itemcode"`
	CoinCode   int       `json:"coincode"`
	MkCode     int       `json:"mkcode"`
    Price      float64   `json:"price"`
    Qty        float64   `json:"qty"`
} 
type OrdData struct {
	UserId     int 
    Symbol     string    `json:"name"`
	ItemCode   int       `json:"itemcode"`
	CoinCode   int       `json:"coincode"`
	MkCode     int       `json:"mkcode"`
    Price      float64   `json:"price"`
    Qty        float64   `json:"qty"`
} 

type OrderJson struct{
	Tr        string `json:"TR"`        // "ORDER"
	Symbol    string `json:"COIN"`      // 종목 명 "BTC.USDT"
	Channel   string `json:"CHANNEL"`   // 거래 체널 구분  "CN02"
	ItemCode  string `json:"MKIDX"`     // 종목 코드 "4441"
	MkCode    string `json:"MKCIDX"`    // 시장 코드 "0"
	UserId    string `json:"MIDX"`      // 사용자 ID "42284"
	UserType  string `json:"USERTYPE"`  // 개법인구분("00":개인,...)
	CoinCode  string `json:"CIDX"`      // Coin Code "2"
	OrdType   string `json:"ORDTYPE"`   // 1 (1:지정가, 2:시장가)
	BuySell   string `json:"TYPE"`      // "S"(S:매도, B:매수)
	Price     string `json:"PRC"`       // 주문 가격  "89000.00"
	Qty       string `json:"QTY"`       // 주문 수량 "0.001"
	MakeRate  string `json:"M_FEE"`     // Maker 수수료 율"0.0007"
	TakeRate  string `json:"T_FEE"`     // Taker 수수료 율"0.0007"  
	CallUnit  string `json:"CALLUNIT"`  // 호가단위
	MinTrade  string `json:"MINTRADE"`  // 최소주문금액
	OrgNum    string `json:"ORGSEQ"`    // 주문번호
	Cancel    string `json:"CANCEL"`    // "0" "1"
	UserIp    string `json:"USERIP"`    // "222.107.37.103"
	Uuid      string `json:"UUID"`      // "f74fef70-9124-4342-a44a-c2cd4df37922"
	UserField string `json:"USERFIELD"` // 사용자 정의 필드 ""
}
type RtnJson struct{
	Tr        string `json:"TR"`        // "ORDER"
	Symbol    string `json:"COIN"`      // 종목 명 "BTC.USDT"
	Channel   string `json:"CHANNEL"`   // 거래 체널 구분  "CN02"
	ItemCode  string `json:"MKIDX"`     // 종목 코드 "4441"
	MkCode    string `json:"MKCIDX"`    // 시장 코드 "0"
	UserId    string `json:"MIDX"`      // 사용자 ID "42284"
	UserType  string `json:"USERTYPE"`  // 개법인구분("00":개인,...)
	CoinCode  string `json:"CIDX"`      // Coin Code "2"
	OrdType   string `json:"ORDTYPE"`   // 1 (1:지정가, 2:시장가)
	BuySell   string `json:"TYPE"`      // "S"(S:매도, B:매수)
	Price     string `json:"PRC"`       // 주문 가격  "89000.00"
	Qty       string `json:"QTY"`       // 주문 수량 "0.001"
	MakeRate  string `json:"M_FEE"`     // Maker 수수료 율"0.0007"
	TakeRate  string `json:"T_FEE"`     // Taker 수수료 율"0.0007"  
	CallUnit  string `json:"CALLUNIT"`  // 호가단위
	MinTrade  string `json:"MINTRADE"`  // 최소주문금액
	OrgNum    string `json:"ORGSEQ"`    // 주문번호
	Cancel    string `json:"CANCEL"`    // "0" "1"
	UserIp    string `json:"USERIP"`    // "222.107.37.103"
	Uuid      string `json:"UUID"`      // "f74fef70-9124-4342-a44a-c2cd4df37922"
	UserField string `json:"USERFIELD"` // 사용자 정의 필드 ""

	ExRate    string `json:"EXRATE"`    // USD 환율
	RegNum    string `json:"REGNUM"`    // 접수번호 ""
	RegTime   string `json:"REGTIME"`   // 접수시각 TimeStamp""
	OrdNum    string `json:"ORDNUM"`    // 주문번호
	OrdTime   string `json:"ORDTIME"`   // 주문시각 TimeStamp
	State     string `json:"STATE"`     // 상태코드  0
	Margin    string `json:"O_USING"`   // 주문 증거금 "89.0000"
	Fee       string `json:"O_FEE"`     // 체결 예상 증거금 "0.0"
	Message   string `json:"MESSAGE"`   // 처리 Message ""
}

var CoinList []Coin
var Orders   []OrdData
var PubSocket *zmq.Socket

func getBuySell()int{
	rand.Seed(time.Now().UnixNano())
	idx := rand.Intn(2)
	return idx
}
func getUid()int{
	rand.Seed(time.Now().UnixNano())
	usid := rand.Intn(10)
	return UID_BASE+usid
}
func getCoin()int{
	rand.Seed(time.Now().UnixNano())
	idx := rand.Intn(len(CoinList))
	return idx
}
func getPrice(prc float64)float64{
    rand.Seed(time.Now().UnixNano())
    fluctuation := 0.001
	change := 1 + (fluctuation * (rand.Float64()*2 - 1))
	return prc *change
}
func getQty(qty float64)float64{
    rand.Seed(time.Now().UnixNano())
    fluctuation := 0.5
	change := 1 + (fluctuation * (rand.Float64()*2 - 1))
	return qty *change
}
func readCoin()error{
	file ,err := os.Open("coins.json")
	if err != nil{return err;}
	defer file.Close()

	value, err := ioutil.ReadAll(file)
	if err != nil{return err;}

	if err := json.Unmarshal(value, &CoinList); err != nil{ return err; }
/*
	for _, coin := range CoinList {
		Log.Infor("%v", coin)
	}
*/
	return nil
}
func makeOrdJson(sock *zmq.Socket, idx int)int {
	mode := getBuySell() + 1 // 1:매도, 2:매수
	var mstr string
	if mode == 1{ mstr = "S";}else{ mstr = "B";}

	stord := OrderJson{
    	Tr       : "ORDER",  // "ORDER"
    	Symbol   : Orders[idx].Symbol,  // 종목 명 "BTC.USDT"
    	Channel  : "CN02",  // 거래 체널 구분  "CN02"
    	ItemCode : fmt.Sprintf("%d",Orders[idx].ItemCode),  // 종목 코드 "4441"
    	MkCode   : fmt.Sprintf("%d",Orders[idx].MkCode),  // 시장 코드 "0"
    	UserId   : fmt.Sprintf("%d",Orders[idx].UserId),  // 사용자 ID "42284"
    	UserType : "00",  // 개법인구분("00":개인,...)
    	CoinCode : fmt.Sprintf("%d", Orders[idx].CoinCode),  // Coin Code "2"
    	OrdType  : "L",  // L (L:지정가, M:시장가)
    	BuySell  : mstr,  // "S"(S:매도, B:매수)
    	Price    : fmt.Sprintf("%f", getPrice(Orders[idx].Price)),  // 주문 가격  "89000.00"
    	Qty      : fmt.Sprintf("%f", getQty(Orders[idx].Qty)),  // 주문 수량 "0.001"
    	MakeRate : "0.0007",  // Maker 수수료 율"0.0007"
    	TakeRate : "0.0007",  // Taker 수수료 율"0.0007"  
    	CallUnit : "0.01",  // 호가단위
    	MinTrade : "0.0001",  // 최소주문금액
    	OrgNum   : "",  // 주문번호
    	Cancel   : "0",  // "0" "1"
    	UserIp   : "222.107.37.103",
    	Uuid     : "f74fef70-9124-4342-a44a-c2cd4df37922",
    	UserField: "사용자 Data", 
	}
	jstr, err := json.Marshal(stord)
	if err != nil{
		fmt.Printf("Error Marshaling JSON : %v", err)
		return 0;
	}
	fmt.Printf("%s\n", jstr)

	sock.SendMessage(jstr)
	
	var rtnOrd RtnJson
	msg, _:=sock.RecvMessage(0)
	err = json.Unmarshal([]byte(msg[0]), &rtnOrd)	
	if err != nil{ return 0;}
	
	if rtnOrd.State != "0" {
		fmt.Printf("%v\n", msg)
		return 0
	}else{
		fmt.Printf("접수번호[%s] 상태 [%s]\n\n", rtnOrd.RegNum, rtnOrd.State)
	}
	return 1
}
// count 개 주문정보 만들기
func makeOrder(count int){
	for i:= 0; i< count; i++{
		idx := getCoin() // Random Index
		userid := getUid()
		ord := OrdData{
			UserId   : userid,
			Symbol   : CoinList[idx].Symbol,
			Price    : getPrice(CoinList[idx].Price),
			Qty      : getQty(CoinList[idx].Qty),
			ItemCode : CoinList[idx].ItemCode,
			CoinCode : CoinList[idx].CoinCode,
			MkCode   : CoinList[idx].MkCode,
		}
		Orders = append(Orders, ord)
	}
}
func main(){

	Log.InitLogger("", Log.CONSOLE_MODE)
	defer Log.Close()

	if len(os.Args) != 2 {
		Log.Error("Using) genord <REQ endpoint>")
		return
	}

	sock, err := zmq.NewSocket(zmq.REQ) 
	if err != nil {Log.Error("%v", err);return}
	defer sock.Close()
	err = sock.Connect(os.Args[1])
	if err != nil{
		Log.Error("ReqSocket.Connect(%s) : %v", os.Args[1], err)
		return
	}
	time.Sleep(100 * time.Microsecond)

	err = readCoin()
	if err != nil {
		Log.Error("%v", err)
		return
	}
	makeOrder(5000)  // 주문 5000개
	for i := 1; i <= 10000; i++{
		rand.Seed(time.Now().UnixNano())
		idx := rand.Intn(len(Orders))
		rtn := makeOrdJson(sock, idx)
		if rtn == 0 {break;}
/*
		if i % 10 == 0{
			time.Sleep(time.Millisecond * 20)
		}
*/
		//time.Sleep(time.Millisecond * 1000) 
		//time.Sleep(time.Millisecond * 5000) // 5000/1000
	}
}
