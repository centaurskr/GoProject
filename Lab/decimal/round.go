///
/// @brief  : Big Decimal round test
/// @file   : round.go
/// @date   : 2024. 09. 13. (금) 10:21:40 KST
/// @author : Cento
///

package main

import (
	"fmt"
	Dc "github.com/shopspring/decimal"
)

func main() {
	price, _ := Dc.NewFromString("57980.5678")
	qty ,  _ := Dc.NewFromString("0.123456785")
	fmt.Println("NewFromString(57980.5678)  :", price)
	fmt.Println("NewFromString(0.123456785) :", qty)

	fmt.Println("\nDivisionPrecision = 3 설정 후 /1, Defualt 16")
	fmt.Println("DivisionPrecision 은 나누기에 만 적용 됨")
	Dc.DivisionPrecision = 3
	price = price.Div(Dc.NewFromFloat(1))
	qty   = qty.Div(Dc.NewFromFloat(1))
	
	fmt.Println("PRICE :", price.String())
	fmt.Println("QTY   :", qty.String())

	fmt.Println("\nDivisionPrecision = 16으로 복구")
	Dc.DivisionPrecision = 16
	d1 := Dc.NewFromFloat(2).Div(Dc.NewFromFloat(3))
	fmt.Println(" 2.0 / 3.0 = ", d1.String())

	fmt.Println("\nRound test")
	fmt.Println("                   12345678901234567")
	fmt.Println("원본          : ", d1             ) 
	fmt.Println("ROUND(8)      : ", d1.Round(8)    , "   - 4사 5입")
	fmt.Println("ROUND DOWN(8) : ", d1.RoundDown(8), "   - 강제 절사")
	fmt.Println("ROUND UP  (8) : ", d1.RoundUp(8)  , "   - 강제 올림")
}
