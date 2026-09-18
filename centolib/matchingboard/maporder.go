///                 12345678901234567890
/// uint64 최대값 : 18446744073709551615    ( 9(20) )
// 주문번호       :  yymmdd + 0000000000000 (13자리-1조)  
//                   (yymmdd + Seq -> 9(19) )
// 접수번호       :  yymmdd+00+00000000000  (11자리-100억)
//                   (yymmdd + line no + Seq -> 9(19) )
package matchingboard
import (
	"fmt"
	//"time"
	"strconv"
	//"encoding/json"
	Dc "github.com/shopspring/decimal"
	Bd "centolib/bindecimal"
)
const (
	ORDSEQ_MAX = 9999999900000000000
	ORDSEQ_MIN = 0
	MAX_LINE_COUNT=100  // 주문 수신 Line Max Count 
)
/////////////// Local Function /////////////////////
func GetDecimal(val interface {})(Dc.Decimal, int){
	switch v := val.(type){
		case Bd.BinDecimal : return v.Decimal(), 0
		case Dc.Decimal : return v, 0
		case string : dec, _ := Dc.NewFromString(v); return dec, 0
		default : dec :=Dc.New(0,0); return dec, -1
	}
}
///////////////////////// 환율정보 /////////////////////
type EXRATE struct{
	BaseTime  string `json:"TIME"`
	Currency  string `json:"CURRENCY"`
	Rate      string `json:"RATE"`
}
///////////////////////// 접수번호 Structure ////////////
type MapHeader struct {
	Create    [60]byte  // 생성일
	Buffer    int        // Buffer count
	Count     int        // Data Count
}
type RegNumber struct {
	Index  int
	RegNum uint64    // 접수번호 yymmdd+00+00000000000  (11자리-100억)
}
///// 주문 상태 (MapHeader + OrderCondition)///////////
type OrderCondition struct{
	Symbol [32]byte   // Coin Symbol
	Group      int    // 주문 GROUP
	Status     int    // 상태 0:주문정지, 1:주문가능, 3:매도만 가능
}
//////////////////////// OrderHeader ////////////////////
type OrderHeader struct{      // Sorting key : Price
    Price    Bd.BinDecimal    // 가격
    Symbol   [32]byte         // Coin이름
    BuySell  byte              // S:Medo, B:Mesu
    BufCount int              // Map file의 Order Buffer 갯수
    OrderCount int              // Map file의 Order Index(추가)
    SumQty   Bd.BinDecimal    // 주문 총수량
}


func (p *OrderHeader)SetPrice (v interface{})int{
	val,ecode := GetDecimal(v)
	if ecode != 0 {return -1}

	p.Price.SetFromDecimal(val)
	return 0
}
func (p *OrderHeader)SetSumQty(v interface{})int{
	val,ecode := GetDecimal(v)
	if ecode != 0 {return -1}
	p.SumQty.SetFromDecimal(val)
	return 0
}
func (p *OrderHeader)SetSymbol(val string)int{
	if len(val) > len(p.Symbol){ return -1}
	copy(p.Symbol[:], val)
	return 0
}

// Map Header SumQty 증가(Bindecimal)
func (p *OrderHeader)AddSumQty(a interface {}, mode, pl int){
	p.SumQty, _ = p.SumQty.Add(a, mode, pl)
}
// Map Header SumQty 감소(Bindecimal)
func (p *OrderHeader)SubSumQty(a interface {}, mode, pl int){
	p.SumQty, _= p.SumQty.Sub(a, mode, pl)
}

func (p *OrderHeader)GetDcPrice()Dc.Decimal{return p.Price.Decimal()}
func (p *OrderHeader)GetStrPrice()string{return p.Price.String()}
func (p *OrderHeader)GetDcSumQty()Dc.Decimal{return p.SumQty.Decimal()}
func (p *OrderHeader)GetStrSumQty()string{return p.SumQty.String()}
func (p *OrderHeader)GetSymbol() string{return string(p.Symbol[:])}

//////////////// Order JSON ////////////////////////////
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
	OrgNum    string `json:"ORGSEQ"`    // 원주문번호
	Cancel    string `json:"CANCEL"`    // "0" "1"
	UserIp    string `json:"USERIP"`    // "222.107.37.103"
	Uuid      string `json:"UUID"`      // f74fef70-9124-4342-a44a-c2cd4df37922
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

/////////// Memory에 있는 주문 구조 //////////////////////////
type Order struct{           // Sorting key OrdTime
	UserId   int             // User Id
    OrdNum   uint64          // 주문번호 - 9(19)
    BuySell  byte            // 매도수구분 (S:매도, B:매수)
    OrdType  byte            // 주문종류(L:LIMIT, M:MARKET)
    OrdTime  uint64          // 주문시각, UinxNano(), 1728372731578396031(19)
    RegNum   uint64          // 접수번호 - 9(19)
    CheSeq   uint            // 최종체결번호
    Price    Bd.BinDecimal   // 주문가격
    Qty      Bd.BinDecimal   // 주문수량
    RemQty   Bd.BinDecimal   // 주문잔량
    MakeRate Bd.BinDecimal   // Maker 수수료율
    TakeRate Bd.BinDecimal   // Taker 수수료율
	Margin   Bd.BinDecimal   // 매수주문 증거금(매수인경우만)
	Fee      Bd.BinDecimal   // 매수체결 예상 수수료(매수인 경우만)
    CallUnit Bd.BinDecimal   // 호가단위
	MinTrade Bd.BinDecimal   // 최소주문금액

    OrdPacket [1024]byte     // 주문 수신 JSON
}
////////////// 내부통신 주문 Packet ///////////////////////////////
type InnerOrdPacket struct{
	Cancel    int             // 취소구분 (0:정상, 1:취소)
	OrgNum    uint64          // 원주문번호
	MemData   Order
}
////////////// Maker Taker 체결 Packet ////////////////////////////
type InnerChePacket struct{
	CheKey     uint64        // 체결 Pair Key	
	Symbol     [32]byte      // 종목 명 "BTC.USDT"
	ItemCode   int           // 종목 코드 "4441"
	MkCode     int           // 시장 코드 "0"
	CoinCode   int           // Coin Code "2"
	UserId     int           // User Id
    OrdType    byte          // 주문종류(L:LIMIT, M:MARKET)
	MakeTake   byte          // "M" Maker,  "T" Taker
	OrdNum     uint64        // 주문번호
    BuySell    byte          // 매도수구분 (S:매도, B:매수)
    CheSeq     uint          // 체결번호
    Price      Bd.BinDecimal // 체결가격
    Qty        Bd.BinDecimal // 체결수량
    RemQty     Bd.BinDecimal // 주문잔량
	Margin     Bd.BinDecimal // 주문증거금(차감용)
	Fee        Bd.BinDecimal // 체결증거금(차감용)
	CheAmt     Bd.BinDecimal // 체결 금액(체결가격X체결수량)
	CheFee     Bd.BinDecimal // 체결 수수료(매수/매도 동일)
	FeeRate    Bd.BinDecimal // 체결당시 수수료율(maker or taker)
    CheTime    uint64        // 채결시각, UinxNano(), 1728372731578396031(19)
	CntOrdNum  uint64        // 상대 주문번호	
	ExRate     Bd.BinDecimal // 환율
	UserField [60]byte       // UserField
	
    OrdPacket [1024]byte     // 주문 수신 JSON
}
type JsonChePacket struct{
   CheKey    string `json:"ck"`   // 체결 Pair Key   
   Symbol    string `json:"sy"`   // 종목 명 "BTC.USDT"
   ItemCode  string `json:"ic"`   // 종목 코드 "4441"
   MkCode    string `json:"mc"`   // 시장 코드 "0"
   CoinCode  string `json:"cc"`   // Coin Code "2"
   UserId    string `json:"ui"`   // User Id
   OrdType   string `json:"ot"`   // 주문종류(L:LIMIT, M:MARKET)
   MakeTake  string `json:"mt"`   // "M" Maker,  "T" Taker
   OrdNum    string `json:"on"`   // 주문번호
   BuySell   string `json:"bs"`   // 매도수구분 (S:매도, B:매수)
   CheSeq    string `json:"cs"`   // 체결번호
   Price     string `json:"pr"`   // 체결가격
   Qty       string `json:"qt"`   // 체결수량
   RemQty    string `json:"rq"`   // 주문잔량
   Margin    string `json:"ma"`   // 주문증거금(차감용)
   Fee       string `json:"fe"`   // 체결증거금(차감용)
   CheAmt    string `json:"ca"`   // 체결 금액(체결가격X체결수량)
   CheFee    string `json:"cf"`   // 체결 수수료(매수/매도 동일)
   FeeRate   string `json:"fr"`   // 체결당시 수수료율(maker or taker)
   CheTime   string `json:"ct"`   // 채결시각, UinxNano(),(19자리)
   CntOrdNum string `json:"con"`  // 상대 주문번호   
   ExRate    string `json:"er"`   // 환율
   UserField string `json:"uf"`   // UserField
   OrdPacket string `json:"op"`   // 주문 수신 JSON
}

func (p *Order)SetPrice     (v interface{})int{
	val,ecode := GetDecimal(v)
	if ecode != 0 {return -1}
	p.Price.SetFromDecimal(val) 
	return 0
} 
func (p *Order)SetQty       (v interface{})int{
	val,ecode := GetDecimal(v)
	if ecode != 0 {return -1}
	p.Qty.SetFromDecimal(val)
	return 0
} 
func (p *Order)SetRemQty    (v interface{})int{
	val,ecode := GetDecimal(v)
	if ecode != 0 {return -1}
	p.RemQty.SetFromDecimal(val)
	return 0
}
func (p *Order)IsZeroRemQty () bool{
	return p.RemQty.Decimal().IsZero()
}
func (p *Order)SetMakeRate  (v interface{})int{
	val,ecode := GetDecimal(v)
	if ecode != 0 {return -1}
	p.MakeRate.SetFromDecimal(val)
	return 0
}
func (p *Order)SetTakeRate  (v interface{})int{
	val,ecode := GetDecimal(v)
	if ecode != 0 {return -1}
	p.TakeRate.SetFromDecimal(val)
	return 0
}
func (p *Order)SetCallUnit  (v interface{})int{
	val,ecode := GetDecimal(v)
	if ecode != 0 {return -1}
	p.CallUnit.SetFromDecimal(val)
	return 0
}
func (p *Order)SetMinTrade  (v interface{})int{
	val,ecode := GetDecimal(v)
	if ecode != 0 {return -1}
	p.MinTrade.SetFromDecimal(val)
	return 0
}
func (p *Order)SetBuySell(bs string){
	p.BuySell = bs[0]
	/*
	if bs == "S" {p.BuySell = 'S';} // 매도
	if bs == "B" {p.BuySell = 'B';} // 매수
	*/
}  
func (p *Order)SetOrdType(bs string){
	p.OrdType = bs[0]
	/*
	if bs == "L" {p.OrdType = 'L';} // 지정가
	if bs == "M" {p.OrdType = 'M';} // 시장가
	*/
}  
func (p *Order)SetUserId(str string){
	p.UserId, _ = strconv.Atoi(str)
}

func (p *Order)GetStrPrice    ()string {return p.Price.String()   }
func (p *Order)GetStrQty      ()string {return p.Qty.String()     }
func (p *Order)GetStrRemQty   ()string {return p.RemQty.String()  }
func (p *Order)GetStrMakeRate ()string {return p.MakeRate.String()}
func (p *Order)GetStrTakeRate ()string {return p.TakeRate.String()}
func (p *Order)GetStrCallUnit ()string {return p.CallUnit.String()}

func (p *Order)GetDcPrice   ()Dc.Decimal{return p.Price.Decimal()   }
func (p *Order)GetDcQty     ()Dc.Decimal{return p.Qty.Decimal()     }
func (p *Order)GetDcRemQty  ()Dc.Decimal{return p.RemQty.Decimal()  }
func (p *Order)GetDcMakeRate()Dc.Decimal{return p.MakeRate.Decimal()}
func (p *Order)GetDcTakeRate()Dc.Decimal{return p.TakeRate.Decimal()}
func (p *Order)GetDcCallUnit()Dc.Decimal{return p.CallUnit.Decimal()}

/////////////////////////// DUMP ///////////////////////////////////
func (p *OrderHeader)Dump(){
	if p == nil {return}
fmt.Printf("=== Dump OrderHeader ==\n")
fmt.Printf("\tCOIN(%s) 가격(%s) BuySell(%c) BufCnt(%d) OrderCount(%d) 총수량(%s)\n",
	string(p.Symbol[:]), 
	p.GetStrPrice(),
	p.BuySell, 
	p.BufCount, p.OrderCount, 
	p.GetStrSumQty())
}
func (p *Order)Dump(idx int ){
	if p == nil {return}
	fmt.Printf("== Dump Order-index[%d]\n", idx)
	fmt.Printf("UserId[%d] OrdNum(%d) RegNum(%d) CheSeq(%d)\n",
		p.UserId, p.OrdNum,  p.RegNum,  p.CheSeq)

	if p.OrdNum == 0 || p.OrdNum == ORDSEQ_MAX || p.OrdNum == ORDSEQ_MIN {return}

	fmt.Printf("\tOrdType(%c) BuySell(%c) OrdTime(%d)\n",
		p.OrdType, p.BuySell , p.OrdTime )
	fmt.Printf("\tPrice(%s)   Qty(%s)   RemQty(%s)\n",
		p.Price.String() ,  p.Qty.String() , p.RemQty.String())  
	fmt.Printf("\tMargin(%s) Fee(%s)\n", p.Margin.String(), p.Fee.String())
	fmt.Printf("\tMakeRate(%s) TakeRate(%s) CallUnit(%s) MinTrade(%s)\n",
		p.MakeRate.String(), p.TakeRate.String(), p.CallUnit.String(),p.MinTrade.String())
	fmt.Printf("JSON : %s\n\n", string(p.OrdPacket[:]))
}
