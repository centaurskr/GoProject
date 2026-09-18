// AfterMatch JSON 변환 함수 모음.
package main

import (
	"fmt"
	Mb "centolib/matchingboard"
	"centolib/api"
)

// MakeJsonChePacket은 내부 체결 패킷(InnerChePacket)을 JSON 직렬화 가능한 JsonChePacket으로 변환한다.
func MakeJsonChePacket(inpk *Mb.InnerChePacket) Mb.JsonChePacket{

	jpk := Mb.JsonChePacket{
		CheKey     : fmt.Sprintf("%d", inpk.CheKey),
		Symbol     : string(inpk.Symbol[:]),
		ItemCode   : fmt.Sprintf("%d", inpk.ItemCode),
		MkCode     : fmt.Sprintf("%d", inpk.MkCode  ),
		CoinCode   : fmt.Sprintf("%d", inpk.CoinCode),
		UserId     : fmt.Sprintf("%d", inpk.UserId  ),
		OrdType    : fmt.Sprintf("%c", inpk.OrdType ),
		MakeTake   : fmt.Sprintf("%c", inpk.MakeTake), 
		OrdNum     : fmt.Sprintf("%d", inpk.OrdNum  ), 
		BuySell    : fmt.Sprintf("%c", inpk.BuySell ), 
		CheSeq     : fmt.Sprintf("%d", inpk.CheSeq  ),
		Price      : inpk.Price.String(),    
		Qty        : inpk.Qty.String(),      
		RemQty     : inpk.RemQty.String(),   
		Margin     : inpk.Margin.String(),   
		Fee        : inpk.Fee.String(),      
		CheAmt     : inpk.CheAmt.String(),   
		CheFee     : inpk.CheFee.String(),   
		FeeRate    : inpk.FeeRate.String(),  
		CheTime    : fmt.Sprintf("%d", inpk.CheTime  ),
		CntOrdNum  : fmt.Sprintf("%d", inpk.CntOrdNum),
		ExRate     : inpk.ExRate.String(),   
		UserField  : api.CleanString(inpk.UserField[:]),
		OrdPacket  : api.CleanString(inpk.OrdPacket[:]),
	}
	return jpk
}
