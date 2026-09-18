///
/// @brief  : GOLANG Decimal 연산 TEST
/// @file   : calc.go
/// @date   : 2024. 09. 13. (금) 11:36:27 KST
/// @author : Cento
///
package main

import (
	"fmt"
	"reflect"
	Dc "github.com/shopspring/decimal"
)

func main(){
// Add, Sub, Mul, Div, Mod, QuoRem
// DivRound
// Copy
// IntPart, NumDigits
// Min, Max, Avg
//
	strprc := "987654321.123456789"
	strqty := "0.987654387"
	prc, _ := Dc.NewFromString(strprc)
	qty, _ := Dc.NewFromString(strqty)

	fmt.Printf("\nSOURCE : prc %s, qty : %s\n", strprc, strqty)
	fmt.Printf("TypeOf : %s\n", reflect.TypeOf(prc))
	fmt.Printf("ADD : %s + %s = %v\n", strprc, strqty, prc.Add(qty))
	fmt.Printf("SUB : %s - %s = %v\n", strprc, strqty, prc.Sub(qty))
	fmt.Printf("SUB : %s - %s = %v\n", strqty, strprc, qty.Sub(prc))
	fmt.Printf("MUL : %s X %s = %v\n", strprc, strqty, prc.Mul(qty))
	fmt.Printf("DIV : %s / %s = %v\n", strprc, strqty, prc.Div(qty))

	fmt.Println(" ")
	fmt.Println("반올림나누기 TEST (4死5入)                                  1234567890")
	fmt.Printf("DivRound(8) : %s / %s = %v\n", strprc, strqty, prc.DivRound(qty, 8))
	fmt.Printf("DivRound(9) : %s / %s = %v\n", strprc, strqty, prc.DivRound(qty, 9))
	prc1 := prc.Copy()
	fmt.Printf("Copy : %v\n", prc1)
	fmt.Printf("IntPart(정수부분) %v => %v\n", prc1, prc1.IntPart())
	fmt.Println("                                      12345678901234567890")
	fmt.Printf("Numdigits(숫자 자리수 소수점 미 포함) %v => %v\n", prc1, prc1.NumDigits())

	strprc = "8.5"
	strqty = "0.3"
	prc, _ = Dc.NewFromString(strprc)
	qty, _ = Dc.NewFromString(strqty)
	fmt.Println("======= 몫과 나머지 계산 QuoRem()=================")
	quot, rem := prc.QuoRem(qty, 0)
	fmt.Printf("%v.QuoRem(%v, 0) 소수점 이하 없이=> Quo : %v, Rem : %v\n", prc, qty, quot, rem)
	quot, rem = prc.QuoRem(qty, 1)
	fmt.Printf("%v.QuoRem(%v, 1) 소수점이하 1자리 => Quo : %v, Rem : %v\n", prc, qty, quot, rem)
	quot, rem = prc.QuoRem(qty, 8)
	fmt.Printf("%v.QuoRem(%v, 8) 소수점 이하 8자리 => Quo : %v, Rem : %v\n", prc, qty, quot, rem)

	fmt.Printf("Convert to String %v : %s\n", prc, prc.String())

	fmt.Println("=======버림,올림,반올림 TEST ==============")
	strprc = "987654321.123456789"
	prc, _ = Dc.NewFromString(strprc)
	
	fmt.Printf("\n올림          %v - 수수료 처리\n", prc.String())
	fmt.Printf("RoundUp(-2) : %v\n", prc.RoundUp(-2))
	fmt.Printf("RoundUp(-1) : %v\n", prc.RoundUp(-1))
	fmt.Printf("RoundUp(0)  : %v\n", prc.RoundUp(0))
	fmt.Printf("RoundUp(1)  : %v\n", prc.RoundUp(1))
	fmt.Printf("RoundUp(3)  : %v\n", prc.RoundUp(3))
	fmt.Printf("RoundUp(4)  : %v\n", prc.RoundUp(4))
	fmt.Printf("RoundUp(5)  : %v\n", prc.RoundUp(5))
	fmt.Printf("RoundUp(6)  : %v\n", prc.RoundUp(6))

	fmt.Printf("\n버림            %v - 이자 처리\n", prc.String())
	fmt.Printf("RoundDown(-2) : %v\n", prc.RoundDown(-2))
	fmt.Printf("RoundDown(-1) : %v\n", prc.RoundDown(-1))
	fmt.Printf("RoundDown(0)  : %v\n", prc.RoundDown(0))
	fmt.Printf("RoundDown(1)  : %v\n", prc.RoundDown(1))
	fmt.Printf("RoundDown(3)  : %v\n", prc.RoundDown(3))
	fmt.Printf("RoundDown(5)  : %v\n", prc.RoundDown(5))
	fmt.Printf("RoundDown(6)  : %v\n", prc.RoundDown(6))
	fmt.Printf("RoundDown(7)  : %v\n", prc.RoundDown(7))

	fmt.Printf("\n반올림      %v- 일반계산\n", prc.String())
	fmt.Printf("Round(-2) : %v\n", prc.Round(-2))
	fmt.Printf("Round(-1) : %v\n", prc.Round(-1))
	fmt.Printf("Round(0)  : %v\n", prc.Round(0))
	fmt.Printf("Round(1)  : %v\n", prc.Round(1))
	fmt.Printf("Round(3)  : %v\n", prc.Round(3))
	fmt.Printf("Round(4)  : %v\n", prc.Round(4))
	fmt.Printf("Round(5)  : %v\n", prc.Round(5))
	fmt.Printf("Round(6)  : %v\n", prc.Round(6))

	
	strprc = "-987654321.123456789"
	prc, _ = Dc.NewFromString(strprc)
	prc1, _ = Dc.NewFromString("987654321.123456789")
	fmt.Printf("\n+무한대로 올림       %v, %v\n", prc.String(), prc1.String())
	fmt.Printf("RoundCeil(-2) :       %v, %v\n", prc.RoundCeil(-2), prc1.RoundCeil(-2))
	fmt.Printf("RoundCeil(-1) :       %v, %v\n", prc.RoundCeil(-1), prc1.RoundCeil(-1))
	fmt.Printf("RoundCeil(0)  :       %v, %v\n", prc.RoundCeil(0),  prc1.RoundCeil(0))
	fmt.Printf("RoundCeil(1)  :       %v, %v\n", prc.RoundCeil(1),  prc1.RoundCeil(1))
	fmt.Printf("RoundCeil(3)  :       %v, %v\n", prc.RoundCeil(3),  prc1.RoundCeil(3))
	fmt.Printf("RoundCeil(4)  :       %v, %v\n", prc.RoundCeil(4),  prc1.RoundCeil(4))
	fmt.Printf("RoundCeil(5)  :       %v, %v\n", prc.RoundCeil(5),  prc1.RoundCeil(5))
	fmt.Printf("RoundCeil(6)  :       %v, %v\n", prc.RoundCeil(6),  prc1.RoundCeil(6))
	fmt.Printf("RoundCeil(7)  :       %v, %v\n", prc.RoundCeil(7),  prc1.RoundCeil(7))
	fmt.Println("숫자 관점에서 음수는 버림, 양수는 올림\n")

	fmt.Printf("\n-무한대로 올림         %v, %v\n", prc.String(), prc1.String())
	fmt.Printf("RoundFloor(-2) :       %v, %v\n", prc.RoundFloor(-2), prc1.RoundFloor(-2))
	fmt.Printf("RoundFloor(-1) :       %v, %v\n", prc.RoundFloor(-1), prc1.RoundFloor(-1))
	fmt.Printf("RoundFloor(0)  :       %v, %v\n", prc.RoundFloor(0),  prc1.RoundFloor(0))
	fmt.Printf("RoundFloor(1)  :       %v, %v\n", prc.RoundFloor(1),  prc1.RoundFloor(1))
	fmt.Printf("RoundFloor(3)  :       %v, %v\n", prc.RoundFloor(3),  prc1.RoundFloor(3))
	fmt.Printf("RoundFloor(4)  :       %v, %v\n", prc.RoundFloor(4),  prc1.RoundFloor(4))
	fmt.Printf("RoundFloor(5)  :       %v, %v\n", prc.RoundFloor(5),  prc1.RoundFloor(5))
	fmt.Printf("RoundFloor(6)  :       %v, %v\n", prc.RoundFloor(6),  prc1.RoundFloor(6))
	fmt.Printf("RoundFloor(7)  :       %v, %v\n", prc.RoundFloor(7),  prc1.RoundFloor(7))
	fmt.Println("숫자 관점에서 음수올림, 양수는 버림\n")
}
