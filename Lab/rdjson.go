package main
import (
	"encoding/json"
	Mb "centolib/matchingboard"
	Bd "centolib/bindecimal"
	"bufio"
	"strconv"
	"errors"
	"fmt"
	"os"
)
// ① 주문 Packet의 정합성 검사
// ② 내부 Data형식으로 변경
// ③ 증거금계산
// ④ REQ응답 Json 생성(OrderPacket의 RegTime이후 State, Using, Fee, Message)
//
// ✓㉼ 증거금 계산 (예시: BTC/USDT)
//    해당 종목을 거래하는 경우 안전한 거래를 위해 
//   충분한 BTC 또는 USDT를 확보하기위한 계산
//   (출금 등의 자산변동을 막기 위함)
//
// 【지정가】
//  -. 매수 (USDT 확보 계산)
//   주문수량 X 주문 가격 + 수수료에 대하여  
//   거래 단위 화폐인 USDT 확보를 위한 계산 
//   using = Price X Qty
//   fee   = using X max(MakeRate, TakeRate)
//  -. 매도 (BTC 확보)
//   매도 BTC 수량에 대하여 자산 동결
//   수수료에 대한 것은 계산 하지 않음
//   (수수료는 매도 체결 후 입금되는 USDT에서 차감)
//   using, fee 계산하지 않음 
// 【시장가】
//  매수 후 USDT가 남아 있는 경우 남은 USDT 증거금 해지.
// -. 매수
//  의도 : 100 UDST만큼 BTC를 매수해!
//  입력 JSON Price부분에 100이 넘어옴(Price = 100, Qty = 0)
//  100 USDT를 모두 증거금으로 잡아야 함
//  100 USDT = 매수 금액 + 수수료
//  using = Price
//  fee = using X TakeRate(시장가는 항상 Take)
// -. 매도 
//  의도 :  10 BTC응 매도해! (Price = 0, Qty = 10)
//  입력 JSON Qty부분에  10이 넘어옴
//   using, fee 계산하지 않음 
//  
func ValidateOrder(rowdata []byte)(Mb.Order, Mb.OrderPacket, error){
var order Mb.Order
var pk Mb.OrderPacket

	err := json.Unmarshal(rowdata, &pk)
	if err != nil{
		return order, pk, err
	}
	// 🔴 𝟭 주문 Packet 유효성 검사
	if pk.Tr != "ORDER" {
		pk.State = "COM-IF-1"
		pk.Message = fmt.Sprintf("TR 오류 ORDER ≠ %s(1)", pk.Tr)
		return order , pk, errors.New(pk.Message)
	}
	if pk.Symbol == "" {
		pk.State = "COM-IF-2"
		pk.Message = "거래대상 COIN이 없습니다(2)"
		return order , pk, errors.New(pk.Message)
	}
	switch pk.Channel{
		case "CN01" :
		case "CN02" :
		case "CN03" :
		case "CN04" :
		default :
			pk.State = "COM-IF-3"
			pk.Message = fmt.Sprintf("거래 체널 오류. %s는 알수 없음", pk.Channel)
			return order , pk, errors.New(pk.Message)
	}
	// MkCode, ItemCode Pass : Symbol로 구함
	if pk.UserId == ""{
		pk.State = "COM-IF-4"
		pk.Message = "USER ID가 없습니다."
		return order , pk, errors.New(pk.Message)
	}
	if pk.UserType == ""{
		pk.State = "COM-IF-5"
		pk.Message = "개인 법인 구분이 없습니다."
		return order , pk, errors.New(pk.Message)
	}
	if pk.OrdType != "1" && pk.OrdType != "2"{
		pk.State = "COM-IF-6"
		pk.Message = fmt.Sprintf("주문유형 (%s)는 지원하지 않습니다.", pk.OrdType)
		return order , pk, errors.New(pk.Message)
	}

	switch pk.BuySell {
		case "B" :
		case "S" :
		default  :
			pk.State = "COM-IF-7"
			pk.Message = fmt.Sprintf("매도수 구분오류[%s]", pk.BuySell)
			return order ,pk,  errors.New(pk.Message)
	}

	if pk.Price == "" && pk.OrdType == "1"{ // 지정가이면서 가격이 없음
		pk.State = "COM-IF-8"
		pk.Message = "주문 가격이 없습니다."
		return order , pk, errors.New(pk.Message)
	}
	if pk.Qty == "" && pk.OrdType == "1"{ // 지정가이면서 수량이 없음
		pk.State = "COM-IF-9"
		pk.Message = "주문 수량이 없습니다."
		return order , pk, errors.New(pk.Message)
	}

	if pk.MakeRate == ""{
		pk.State = "COM-IF-10"
		pk.Message = "수수료율이 없습니다.(Maker)"
		return order , pk, errors.New(pk.Message)
	}
	if pk.TakeRate == ""{
		pk.State = "COM-IF-11"
		pk.Message = "수수료율이 없습니다.(Taker)"
		return order , pk, errors.New(pk.Message)
	}
	if pk.CallUnit == ""{
		pk.State = "COM-IF-12"
		pk.Message = "호가단위가 없습니다(CALLUNIT)"
		return order , pk, errors.New(pk.Message)
	}
	if pk.MinTrade == ""{
		pk.State = "COM-IF-13"
		pk.Message = "최소호가수량이 없습니다(MINTRADE)"
		return order , pk, errors.New(pk.Message)
	}
	if pk.Cancel == "" {
		pk.State = "COM-IF-14" 
		pk.Message = "취소구분이 없습니다.(CANCEL)"
		return order , pk, errors.New(pk.Message)
	}
	if pk.OrgNum == "" && pk.Cancel == "1"{
		pk.State = "COM-IF-15" 
		pk.Message = "취소주문, 원주문번호가 없습니다.(ORGNUM)"
		return order , pk, errors.New(pk.Message)
	}
	if pk.UserIp == ""{
		pk.State = "COM-IF-16" 
		pk.Message = "사용자 IP가 없습니다"
		return order , pk, errors.New(pk.Message)
	}
	if pk.Uuid == ""{
		pk.State = "COM-IF-17" 
		pk.Message = "사용자 UUID가 없습니다"
		return order , pk, errors.New(pk.Message)
	}
	// 시장가 매도 수량 오류
	if pk.Qty == "" && pk.OrdType =="2" && pk.BuySell == "S" {
		pk.State = "COM-IF-18" 
		pk.Message = "시장가 매도 수량이 없습니다"
		return order , pk, errors.New(pk.Message)
	}
	// 시장가 매수 가격 오류
	if pk.Price == "" && pk.OrdType =="2" && pk.BuySell == "B" {
		pk.State = "COM-IF-19" 
		pk.Message = "시장가 매수 금액이 없습니다"
		return order , pk, errors.New(pk.Message)

	}

	pk.State = "0"
	pk.Message = "주문이 접수되었습니다."
	// 🔴 𝟮 내부사용 주문 packet 만들기
    // int           User Id
	order.UserId,_   = strconv.Atoi(pk.UserId)

	// uint64        주문번호 - 9(19)
    //order.OrdNum   = pk.               

	// int           매도수구분 (1:매도 2:매수)
	order.SetBuySell(pk.BuySell)       
	// int           주문종류(1:LIMIT, 2:MARKET)
    order.OrdType,_  = strconv.Atoi(pk.OrdType)       

	// uint64        주문시각, UinxNano(), 1728372731578396031(19)
    //order.OrdTime  = pk.               
	// uint64        접수번호 - 9(19)
    //order.RegNum   = pk.               
	// uint          최종체결번호
    //order.CheSeq   = pk.               

    // Bd.BinDecimal 주문가격
    order.SetPrice (pk.Price)
	// Bd.BinDecimal 주문수량
    order.SetQty   (pk.Qty)            
	// Bd.BinDecimal 주문잔량
    order.SetRemQty(pk.Qty)              
	// Bd.BinDecimal Maker 수수료율
    order.SetMakeRate(pk.MakeRate)     
	// Bd.BinDecimal Taker 수수료율
    order.SetTakeRate(pk.TakeRate)     
	// Bd.BinDecimal 최소 호가단위(결제통화기준)
    order.SetMinTrade(pk.MinTrade)     


	var fee,using Bd.BinDecimal
	var eno int

	// 1. 지정가(매수만 증거금 계산)
	if order.OrdType == 1 && order.BuySell == 2{ 
		// Using = qty X price
		using, eno = order.Price.Mul(order.Qty, Bd.ROUND_UP, 8)
		if eno != 0{
			pk.State = "COM-IF-20" 
			pk.Message = fmt.Sprintf(
				"지정가 증거금계산 오류(using) Price(%s) Qty(%s)",
				order.Price.String(), 
				order.Qty.String())
			return order , pk, errors.New(pk.Message)
		}
		order.Using = using
		pk.Using = using.String()

		// Fee   = Using X MAX(MakeRate, TakeRate) - 올림
		rtn := Bd.DecimalCompare(order.MakeRate, order.TakeRate)
		if rtn == 1 {  // Make > Take
			fee, eno = order.Using.Mul(order.MakeRate, Bd.ROUND_UP, 8)
		}else{         // Make <= Take
			fee, eno = order.Using.Mul(order.TakeRate, Bd.ROUND_UP, 8)
		}
		if eno != 0{
			pk.State = "COM-IF-21" 
			pk.Message = fmt.Sprintf(
				"지정가 증거금계산 오류(fee) Price(%s) Qty(%s)",
				order.Price.String(), 
				order.Qty.String())
			return order , pk, errors.New(pk.Message)
		}
		order.Fee = fee
		pk.Fee = fee.String()
	}else if order.OrdType == 2 && order.BuySell == 2 {
		// 2. 시장가 증거금 계산(매수만 계산)
		// 계산 순서는 Fee부터 계산 해야 함
		// Fee = Price X TakeRate
		fee, eno = order.Price.Mul(order.TakeRate, Bd.ROUND_UP, 8)
		if eno != 0{
			pk.State = "COM-IF-22" 
			pk.Message = fmt.Sprintf(
				"시장가 증거금계산 오류(fee) Price(%s) TakeRate(%s)",
				order.Price.String(), 
				order.TakeRate.String())
			return order , pk, errors.New(pk.Message)
		}
		order.Fee = fee
		pk.Fee = fee.String()

		// Using = Price - Fee
		using, eno = order.Price.Sub(fee, Bd.ROUND_UP, 8)
		if eno != 0{
			pk.State = "COM-IF-23" 
			pk.Message = fmt.Sprintf(
				"시장가 증거금계산 오류(using) Price(%s) Fee(%s)",
				order.Price.String(), 
				order.Fee.String())
			return order , pk, errors.New(pk.Message)
		}
		order.Using = using
		pk.Using = using.String()
	}
	// [1024]byte      주문 수신 JSON
    copy(order.OrdPacket[:], rowdata)       

	return order , pk, nil
}
//////////////// Main ///////////////////////////////
func main(){
	file, err := os.Open("ord.txt")
	if err != nil{
		fmt.Printf("Error : %v\n", err)
		os.Exit(0)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		order, pk, err := ValidateOrder([]byte(line))
		if err != nil{
			fmt.Printf("ValidateOrder() : %v\n", err)
			os.Exit(0)
		}else{
			fmt.Printf("%+v\n", order)
			fmt.Printf("%+v\n", pk)
		}
	} // For Loop

}
