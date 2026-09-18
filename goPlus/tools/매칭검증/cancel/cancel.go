package main
import (
	"fmt"
	"centolib/api"
	"os"
	zmq "github.com/pebbe/zmq4"
	Mb "centolib/matchingboard"
	"encoding/json"
	"time"
)
// SendCancel은 취소 주문 메시지를 ORDERPASS Proxy에 전송하고 응답을 반환한다.
func SendCancel(
	symbol string,
	itemCode int,
	mkCode   int,
	coinCode int,
	buySell  string,
	price    float64,
	uid      int,
	orgNum   string)([]string, error){

	reqPoint := "tcp://127.0.0.1:31001"

	itemCodeStr := fmt.Sprintf("%v", itemCode)
	mkCodeStr   := fmt.Sprintf("%v", mkCode)
	coinCodeStr := fmt.Sprintf("%v", coinCode)
	priceStr    := fmt.Sprintf("%v", price)
	uidStr      := fmt.Sprintf("%v", uid)
	packet := Mb.OrderJson{
        Tr       : "ORDER",
        Symbol   : symbol,
        Channel  : "CN03",
        ItemCode : itemCodeStr,
        MkCode   : mkCodeStr,    // 시장 코드 "0"
        UserId   : uidStr,       // 사용자 ID 
        UserType : "00",      // 개법인구분("00":개인,...)
        CoinCode : coinCodeStr,       // Coin Code "2"
        OrdType  : "L",       // 1 (L:지정가, M:시장가)
        BuySell  : buySell,   // "S"(S:매도, B:매수)
        Price    : priceStr,     // 주문 가격  "89000.00"
        Qty      : "0",       // 주문 수량 "0.001"
        MakeRate : "0.0001",  // Maker 수수료 율"0.0007"
        TakeRate : "0.0007",  // Taker 수수료 율"0.0007"
        CallUnit : "0.01",    // 호가단위
        MinTrade : "0.0001",  // 최소주문금액
        OrgNum   : orgNum,    // 원주문번호
	    Cancel   : "1"     ,  // "0" "1"
        UserIp   : "222.107.37.103",
        Uuid     : "f74fef70-9124-4342-a44a-c2cd4df37922",
        UserField:"USERFIELD", // 사용자 정의 필드 ""
	}
    stream,err := json.Marshal(packet)
	if err != nil{return nil, err;}

/*
	var input string
	fmt.Printf("\n%+v\n", packet)
	fmt.Printf("전송하겠습니까? [y/n]")
	fmt.Scanln(&input)
	if input == "n" || input == "N"{return nil, nil;}

*/
    req,err := zmq.NewSocket(zmq.REQ)
	if err != nil{return nil,  err;}
    req.Connect(reqPoint)
	defer req.Close()

    req.SendMessage(stream)

    msg,_ := req.RecvMessage(0)
	return msg, nil
}
func main(){
	dbYaml, err := api.GetPlatformEnv("DBCONFIG_YAML")
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}
	dbCtx, err := api.ConnectDb(dbYaml)
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}
	defer dbCtx.Close()
	var (
		ordNum    string 
		symbol    string 
		itemCode  int 
		mkCode    int 
		coinCode  int
		userId    int
		channel   string 
		ordType   string
		buySell   string
		price     float64
	)
	sqlStr := `
	SELECT ordnum, symbol, itemcode, mkcode, 
		coincode, userid, channel, ordtype, buysell,
		price
	FROM ordertbl
	WHERE status = 'GO';`
	rows, err := dbCtx.Query(sqlStr)
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}
	defer rows.Close()

	for rows.Next() {
		err :=rows.Scan(
				&ordNum    ,
				&symbol    ,
				&itemCode  ,
				&mkCode    ,
				&coinCode  ,
				&userId    ,
				&channel   ,
				&ordType   ,
				&buySell   ,
				&price)	
		if err != nil{
			fmt.Printf("%v\n", err)
			break
		}
		msg, err := SendCancel( symbol, itemCode, mkCode, coinCode,
					buySell, price, userId, ordNum)
		if err != nil{
			fmt.Printf("%v\n", err)
			break	
		}
		fmt.Printf("ordnum[%v] symbol[%v] itemcode[%v] mkcode[%v] coincode[%v] userid[%v] channel[%v] ordtype[%v] buysell[%v] price[%v]\n\n",
				ordNum    , symbol    , itemCode  , mkCode    , coinCode  ,
				userId    , channel   , ordType   , buySell   , price )
		fmt.Printf("수신:%s\n\n", msg)
		time.Sleep(10 * time.Millisecond)
	} // For Loop

}
