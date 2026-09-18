package main
import (
	"fmt"
	"time"
	sc "centolib/screen"
	"centolib/api"
	zmq "github.com/pebbe/zmq4"
	"encoding/json"
	"os"	
)
type OrderBookEntry struct {
    Price    string `json:"p"`
    Count    string `json:"c"`
    Quantity string `json:"q"`
}

type OrderBook struct {
    Symbol      string           `json:"sy"`
    Timestamp   string           `json:"t"`
    SellOrder   string           `json:"so"`
    SellQty     string           `json:"sq"`
    BuyOrder    string           `json:"bo"`
    BuyQty      string           `json:"bq"`
    SellBook    []OrderBookEntry `json:"S"`
    BuyBook     []OrderBookEntry `json:"B"`
}
type Current struct {
    Symbol      string           `json:"sy"`
    Yyyymmdd    string           `json:"t"`
	Price       string           `json:"p"`
	Qty         string           `json:"q"`
	BuySell     string           `json:"bs"`
}
var (
	Symbol_X, Symbol_Y         int
	Curr_Qty_X, Curr_Qty_Y     int
	Curr_Prc_X, Curr_Prc_Y     int
	Curr_Time_X, Curr_Time_Y   int
	Sell_X, Sell_Y int 
	Buy_X , Buy_Y  int
	Hoga_Time_X, Hoga_Time_Y   int
)
func clearArea(x, y, width, height int) {
    // Move cursor to position (x, y)
    fmt.Printf("\033[%d;%dH", y, x)

    // Loop over the area and clear each line
    for i := 0; i < height; i++ {
        fmt.Printf("\033[K") // Clear the current line
        fmt.Print("\033[B")  // Move down to the next line
    }

    // Reset cursor to the initial position after clearing
    fmt.Printf("\033[%d;%dH", y+height, x)
}
func outStrXy(cl bool, x, y int, format string, a ...interface{}){
    sc.SAVE_CURSOR()
    sc.CURSOR_MOVE(x,y)
    if cl == true {
        sc.CLEAR_LINE_FROMCUR()
    }
    fmt.Printf(format, a...)
    sc.RESTORE_CURSOR()
}
// DspCurrentTime은 현재 시각을 화면 고정 위치에 출력한다.
func DspCurrentTime(){
	str := time.Now().Format("현재시각 : 2006년 01월 02일 - 15:04:05")
	fmt.Printf("\033[1;40H%s", str)
}
// DspSymbol은 종목명을 매수/매도 색상으로 화면에 출력한다.
func DspSymbol(symbol string, mode byte){
	clearArea(Symbol_X, Symbol_Y, 14, 1)	
	sc.ATTR_BOLD()

	switch mode{
	case 'B' : // Buy
		sc.COLOR_RED()
	case 'S' : // Sell
		sc.COLOR_BLUE()
	}
	outStrXy(false, Symbol_X, Symbol_Y, "%s", symbol)
	sc.ATTR_NONE()
	sc.COLOR_RESET()
}
// InitApplication은 화면 출력에 사용하는 좌표 변수들을 초기화한다.
func InitApplication(){
	Symbol_X  = 3   
	Symbol_Y  = 2 
	Curr_Qty_X= 21 
	Curr_Qty_Y= 3 
	Curr_Prc_X= 3  
	Curr_Prc_Y= 3 
	Curr_Time_X = 24 
	Curr_Time_Y = 2
	Sell_X= 1 
	Sell_Y= 19
	Buy_X = 41 
	Buy_Y = 21  
	Hoga_Time_X = 2
	Hoga_Time_Y= 36
}
type Packet struct{
	tr   string
	pk   string
}
// Message는 형식 문자열로 메시지를 화면 하단 고정 영역에 출력한다.
func Message(format string, a ...interface{}){
	outStrXy(true, Hoga_Time_X, Hoga_Time_Y + 1, format, a)
	sc.ATTR_NONE()
}
func displayCurr(jsonStr string){
	var curr Current
	err := json.Unmarshal([]byte(jsonStr), &curr)
	if err != nil{
		Message("%v", err)
		return
	}
	if curr.BuySell == "S" {
		DspSymbol(curr.Symbol, 'S')
	}else{
		DspSymbol(curr.Symbol, 'B')
	}
	fmt.Printf("\033[%d;%dH체결시각 : %s", 
		Curr_Time_Y, Curr_Time_X, curr.Yyyymmdd)
	fmt.Printf("\033[%d;%dH%s", 
		Curr_Prc_Y, Curr_Prc_X, curr.Price)
	fmt.Printf("\033[%d;%dH(%s)", 
		Curr_Qty_Y, Curr_Qty_X, curr.Qty)
}
func displayHoga(jsonStr string){
	var hoga OrderBook

	err := json.Unmarshal([]byte(jsonStr), &hoga)
	if err != nil{
		Message("%v", err)
		return
	}
	clearArea(Sell_X, 5, 39, 15)	 // 매도 영역
//type OrderBookEntry struct {
//    Price    string `json:"p"`
//   Count    string `json:"c"`
//  Quantity string `json:"q"`
//}
	sc.COLOR_BLUE()
	for i, ent := range hoga.SellBook{
		fmt.Printf("\033[%d;1H%14s (%3s)  %18s", 
			Sell_Y - i, ent.Quantity, ent.Count, ent.Price)
	}

	clearArea(Buy_X, Buy_Y, 39, 15)	 // 매수 영역
	sc.COLOR_RED()
	for i, ent := range hoga.BuyBook{
		fmt.Printf("\033[%d;%dH%-18s %-14s (%3s)", 
			Buy_Y + i, Buy_X, ent.Price, ent.Quantity, ent.Count)
	}
	sc.COLOR_RESET()
	outStrXy(true, Hoga_Time_X, Hoga_Time_Y, "%s", hoga.Timestamp)
}
func goDataProcess(channel chan Packet){
	for{
		select{
		case packet := <- channel :
			switch packet.tr {
				case "CURR" :
					displayCurr(packet.pk)
				case "HOGA" :
					displayHoga(packet.pk)
			}
		} // SELECT
	}// for Loop
}
func main(){
	if len(os.Args) != 2{
		fmt.Println("호가 & 현재가 수신 Process")
		fmt.Printf("Using) goHoga <coin>\n")
		os.Exit(0)
	}
	x, y, err := sc.GETWINDOWSIZE()
	if err != nil {
		fmt.Printf("GETWINDOWSIZE() : %v\n", err)
		os.Exit(0)
	}
	if x < 80 || y < 36 {
		fmt.Printf("현재 Window. 폭(%d) 높이(%d)\n", x, y)
		fmt.Printf("Window size가 적습니다.\n")
		fmt.Printf("최소 Size 폭(80) 높이(40)\n")
		os.Exit(0)
	}

	sc.CURSOR_OFF()
	defer sc.CURSOR_ON()

	InitApplication()

	sc.CLEAR_WIN()
	DspSymbol(os.Args[1], 'N')

	Message("MSG : %s", "Message TEST")

	yaml, err := api.GetPlatformEnv("PROXYENV_YAML")
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}
	// 현재가
	val, err := api.GetProxyInfor(yaml, "SISEJSON", "back")
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}
	currSub, err := zmq.NewSocket(zmq.SUB)
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}
	currSub.Connect(val)
	// 호가
	val, err = api.GetProxyInfor(yaml, "HOGAJSON", "back")
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}
	hogaSub, err := zmq.NewSocket(zmq.SUB)
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}
	hogaSub.Connect(val)

	currSub.SetSubscribe(os.Args[1])
	hogaSub.SetSubscribe(os.Args[1])

	channel := make (chan Packet, 5000)

	go goDataProcess(channel)

	poller := zmq.NewPoller()
	poller.Add(currSub, zmq.POLLIN)
	poller.Add(hogaSub, zmq.POLLIN)

	for {
		polls, _ := poller.Poll(1000 * time.Millisecond)
		for _, polled := range polls {
		switch polled.Socket{
		case currSub :
			// msg[0] : BTC.USDT
			// msg[1] : 현재가 
			msg, _ := currSub.RecvMessage(0)
			pk := Packet{
				tr : "CURR",
				pk : string(msg[1]),
			}
			channel <- pk
		case hogaSub:
			// msg[0] : BTC.USDT
			// msg[1] : 호가 String
			msg, _ := hogaSub.RecvMessage(0)
			pk := Packet{
				tr : "HOGA",
				pk : string(msg[1]),
			}
			channel <- pk
		} // SWITCH
		} // FOR (polled)LOOP
		DspCurrentTime()
	} // Main Loop
}
