// 주문 수신 프로세스의 JSON 처리 함수 모음.
package main

import (
	//"encoding/json"
	Mb "centolib/matchingboard"
	Bd "centolib/bindecimal"
	"strconv"
	"strings"
	"errors"
	"fmt"
)
// ValidateOrder는 주문 JSON의 정합성을 검사하고 내부 주문 패킷으로 변환한다.
//
// 처리 단계: ① 주문 JSON 유효성 검사 → ② InnerOrdPacket 변환 → ③ 증거금 계산
//
// 증거금 계산 방식 (예시: BTC/USDT)
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
func ValidateOrder(jsonpk *Mb.OrderJson)(Mb.InnerOrdPacket, error){
var innerpk    Mb.InnerOrdPacket  // for Matching

	// 🔴 𝟭 주문 Packet 유효성 검사
	if jsonpk.Tr != "ORDER" {
		jsonpk.State = "COM-IF-1"
		jsonpk.Message = fmt.Sprintf("TR 오류 ORDER ≠ %s(1)", jsonpk.Tr)
		return innerpk , errors.New(jsonpk.Message)
	}
	if jsonpk.Symbol == "" {
		jsonpk.State = "COM-IF-2"
		jsonpk.Message = "거래대상 COIN이 없습니다(2)"
		return innerpk , errors.New(jsonpk.Message)
	}
	switch jsonpk.Channel{
		case "CN01" :
		case "CN02" :
		case "CN03" :
		case "CN04" :
		default :
			jsonpk.State = "COM-IF-3"
			jsonpk.Message = fmt.Sprintf("거래 체널 오류. %s는 알수 없음",
						 jsonpk.Channel)
			return innerpk , errors.New(jsonpk.Message)
	}
	// ItemCode(MKIDX)
	if jsonpk.ItemCode == ""{
		jsonpk.State = "COM-IF-20"
		jsonpk.Message = "ItemCode(MKIDX)가 없습니다"
		return innerpk , errors.New(jsonpk.Message)
	}
	// MkCode (MKCIDX)
	if jsonpk.MkCode == ""{
		jsonpk.State = "COM-IF-21"
		jsonpk.Message = "MkCode(MKCIDX)가 없습니다"
		return innerpk , errors.New(jsonpk.Message)
	}
	// Coin Code (CIDX)
	if jsonpk.CoinCode == ""{
		jsonpk.State = "COM-IF-22"
		jsonpk.Message = "Coin Code(CIDX)가 없습니다"
		return innerpk , errors.New(jsonpk.Message)
	}
	// MkCode, ItemCode Pass : Symbol로 구함
	if jsonpk.UserId == ""{
		jsonpk.State = "COM-IF-4"
		jsonpk.Message = "USER ID가 없습니다."
		return innerpk , errors.New(jsonpk.Message)
	}
	if jsonpk.UserType == ""{
		jsonpk.State = "COM-IF-5"
		jsonpk.Message = "개인 법인 구분이 없습니다."
		return innerpk , errors.New(jsonpk.Message)
	}
	if jsonpk.OrdType != "L" && jsonpk.OrdType != "M"{
		jsonpk.State = "COM-IF-6"
		jsonpk.Message = fmt.Sprintf("주문유형 (%s)는 지원하지 않습니다.", jsonpk.OrdType)
		return innerpk , errors.New(jsonpk.Message)
	}

	switch jsonpk.BuySell {
		case "B" :
		case "S" :
		default  :
			jsonpk.State = "COM-IF-7"
			jsonpk.Message = fmt.Sprintf("매도수 구분오류[%s]", jsonpk.BuySell)
			return innerpk , errors.New(jsonpk.Message)
	}
	// 주문 상태 Check
	status := checkOrderCondition(jsonpk.Symbol)
	switch status {
		case -1:
		jsonpk.State = "COM-BO-1"
		jsonpk.Message = fmt.Sprintf("지원하지 않는 종목입니다(%s)", jsonpk.Symbol)
		return innerpk , errors.New(jsonpk.Message)
		case 0 :
		jsonpk.State = "COM-BO-2"
		jsonpk.Message = fmt.Sprintf("주문정지 상태입니다(%s)", jsonpk.Symbol)
		return innerpk , errors.New(jsonpk.Message)
		case 3 : 
			if jsonpk.BuySell == "B"{
				jsonpk.State = "COM-BO-3"
				jsonpk.Message = fmt.Sprintf("매도만 가능합니다(%s)", 
												jsonpk.Symbol)
				return innerpk , errors.New(jsonpk.Message)
			}
	}

	if jsonpk.Price == "" && jsonpk.OrdType == "L"{ // 지정가이면서 가격이 없음
		jsonpk.State = "COM-IF-8"
		jsonpk.Message = "주문 가격이 없습니다."
		return innerpk , errors.New(jsonpk.Message)
	}
	if jsonpk.Qty == "" && jsonpk.OrdType == "L"{ // 지정가이면서 수량이 없음
		jsonpk.State = "COM-IF-9"
		jsonpk.Message = "주문 수량이 없습니다."
		return innerpk , errors.New(jsonpk.Message)
	}

	if jsonpk.MakeRate == ""{
		jsonpk.State = "COM-IF-10"
		jsonpk.Message = "수수료율이 없습니다.(Maker)"
		return innerpk , errors.New(jsonpk.Message)
	}
	if jsonpk.TakeRate == ""{
		jsonpk.State = "COM-IF-11"
		jsonpk.Message = "수수료율이 없습니다.(Taker)"
		return innerpk , errors.New(jsonpk.Message)
	}
	if jsonpk.CallUnit == ""{
		jsonpk.State = "COM-IF-12"
		jsonpk.Message = "호가단위가 없습니다(CALLUNIT)"
		return innerpk , errors.New(jsonpk.Message)
	}
	if jsonpk.MinTrade == ""{
		jsonpk.State = "COM-IF-13"
		jsonpk.Message = "최소호가수량이 없습니다(MINTRADE)"
		return innerpk , errors.New(jsonpk.Message)
	}

	if jsonpk.Cancel != "0" && jsonpk.Cancel != "1" {
		jsonpk.State = "COM-IF-14" 
		jsonpk.Message = "취소구분이 없습니다.(CANCEL)"
		return innerpk , errors.New(jsonpk.Message)
	}
	if jsonpk.OrgNum == "" && jsonpk.Cancel == "1"{
		jsonpk.State = "COM-IF-15" 
		jsonpk.Message = "취소주문, 원주문번호가 없습니다.(ORGNUM)"
		return innerpk , errors.New(jsonpk.Message)
	}
	if jsonpk.UserIp == ""{
		jsonpk.State = "COM-IF-16" 
		jsonpk.Message = "사용자 IP가 없습니다"
		return innerpk , errors.New(jsonpk.Message)
	}
	if jsonpk.Uuid == ""{
		jsonpk.State = "COM-IF-17" 
		jsonpk.Message = "사용자 UUID가 없습니다"
		return innerpk , errors.New(jsonpk.Message)
	}
	// 시장가 매도 수량 오류
	if jsonpk.Qty == "" && jsonpk.OrdType =="M" && jsonpk.BuySell == "S" {
		jsonpk.State = "COM-IF-18" 
		jsonpk.Message = "시장가 매도 수량이 없습니다"
		return innerpk , errors.New(jsonpk.Message)
	}
	// 시장가 매수 가격 오류
	if jsonpk.Price == "" && jsonpk.OrdType =="M" && jsonpk.BuySell == "B" {
		jsonpk.State = "COM-IF-19" 
		jsonpk.Message = "시장가 매수 금액이 없습니다"
		return innerpk , errors.New(jsonpk.Message)

	}

	// 환율
	nameTok := strings.Split(jsonpk.Symbol, ".")
	if nameTok[1] == "KRW" {
		jsonpk.ExRate = "1.0"
	}else{
		jsonpk.ExRate = UsdExrate.Rate
	}

	jsonpk.State = "0"
	jsonpk.Message = "주문이 접수되었습니다."
	// 🔴 𝟮 내부사용 주문 packet 만들기
	// 취소구분 
	innerpk.Cancel,_ = strconv.Atoi(jsonpk.Cancel)
	if innerpk.Cancel == 1{
		num, _ := strconv.Atoi(jsonpk.OrgNum)
		innerpk.OrgNum = uint64(num)
	}else{
		innerpk.OrgNum = 0
	}
	// InnerOrdPacket.MemData 만들기
    // int           User Id
	innerpk.MemData.UserId,_   = strconv.Atoi(jsonpk.UserId)
	// int           매도수구분 (1:매도 2:매수)
	innerpk.MemData.SetBuySell(jsonpk.BuySell)       
	// int           주문종류(L:LIMIT, M:MARKET)
    innerpk.MemData.SetOrdType(jsonpk.OrdType)       

    // Bd.BinDecimal 주문가격
    innerpk.MemData.SetPrice (jsonpk.Price)
	// Bd.BinDecimal 주문수량
    innerpk.MemData.SetQty   (jsonpk.Qty)            
	// Bd.BinDecimal 주문잔량
    innerpk.MemData.SetRemQty(jsonpk.Qty)              
	// Bd.BinDecimal Maker 수수료율
    innerpk.MemData.SetMakeRate(jsonpk.MakeRate)     
	// Bd.BinDecimal Taker 수수료율
    innerpk.MemData.SetTakeRate(jsonpk.TakeRate)     
	// Bd.BinDecimal 호가단위
    innerpk.MemData.SetCallUnit(jsonpk.CallUnit)     
	// Bd.BinDecimal 최소주문수량
    innerpk.MemData.SetMinTrade(jsonpk.MinTrade)     

	var fee,margin Bd.BinDecimal
	var eno int

	// 1. 지정가(매수만 증거금 계산)
	if innerpk.MemData.OrdType == 'L' && innerpk.MemData.BuySell == 'B'{ 
		// Margin = qty X price
		margin, eno = innerpk.MemData.Price.Mul(innerpk.MemData.Qty, Bd.ROUND_UP, 8)
		if eno != 0{
			jsonpk.State = "COM-IF-20" 
			jsonpk.Message = fmt.Sprintf(
				"지정가 증거금계산 오류(using) Price(%s) Qty(%s)",
				innerpk.MemData.Price.String(), 
				innerpk.MemData.Qty.String())
			return innerpk , errors.New(jsonpk.Message)
		}
		innerpk.MemData.Margin = margin
		jsonpk.Margin = margin.String()

		// Fee   = Margin X MAX(MakeRate, TakeRate) - 올림
		rtn := Bd.DecimalCompare(innerpk.MemData.MakeRate, innerpk.MemData.TakeRate)
		if rtn == 1 {  // Make > Take
			fee, eno = innerpk.MemData.Margin.Mul(innerpk.MemData.MakeRate, Bd.ROUND_UP, 8)
		}else{         // Make <= Take
			fee, eno = innerpk.MemData.Margin.Mul(innerpk.MemData.TakeRate, Bd.ROUND_UP, 8)
		}
		if eno != 0{
			jsonpk.State = "COM-IF-21" 
			jsonpk.Message = fmt.Sprintf(
				"지정가 증거금계산 오류(fee) Price(%s) Qty(%s)",
				innerpk.MemData.Price.String(), 
				innerpk.MemData.Qty.String())
			return innerpk , errors.New(jsonpk.Message)
		}
		innerpk.MemData.Fee = fee
		jsonpk.Fee = fee.String()
	}else if innerpk.MemData.OrdType == 'M' && innerpk.MemData.BuySell == 'B' {
		// 2. 시장가 증거금 계산(매수만 계산)
		// 계산 순서는 Fee부터 계산 해야 함
		// Fee = Price X TakeRate(시장가는 Taker수수료 적용)
		//
		// ❗㉼ 시장가 주문 정보 jsonpk.Price , jsonpk.Qty
		//          │     Price        │      Qty
		// ─────────┼──────────────────┼─────────────────
		//  매수(B) │   주문가능 금액  │      0.0                
		// ─────────┼──────────────────┼─────────────────
		//  매도(S) │      0.0         │  주문가능 수량                
		// ─────────┴──────────────────┴─────────────────

		fee, eno = innerpk.MemData.Price.Mul(innerpk.MemData.TakeRate, Bd.ROUND_UP, 8)
		if eno != 0{
			jsonpk.State = "COM-IF-22" 
			jsonpk.Message = fmt.Sprintf(
				"시장가 증거금계산 오류(fee) Price(%s) TakeRate(%s)",
				innerpk.MemData.Price.String(), 
				innerpk.MemData.TakeRate.String())
			return innerpk , errors.New(jsonpk.Message)
		}
		innerpk.MemData.Fee = fee
		jsonpk.Fee = fee.String()

		// Margin = Price(주문가능 금액) - Fee
		margin, eno = innerpk.MemData.Price.Sub(fee, Bd.ROUND_UP, 8)
		if eno != 0{
			jsonpk.State = "COM-IF-23" 
			jsonpk.Message = fmt.Sprintf(
				"시장가 증거금계산 오류(margin) Price(%s) Fee(%s)",
				innerpk.MemData.Price.String(), 
				innerpk.MemData.Fee.String())
			return innerpk , errors.New(jsonpk.Message)
		}
		innerpk.MemData.Margin = margin
		jsonpk.Margin = margin.String()
	}else{ // 매도인경우
		innerpk.MemData.Margin.SetFromString("0.0")
		innerpk.MemData.Fee.SetFromString("0.0")
		jsonpk.Margin = "0.0"
		jsonpk.Fee    = "0.0"
	}
	return innerpk , nil
}
