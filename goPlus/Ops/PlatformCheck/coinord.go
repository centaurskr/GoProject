package main

import (
	"os"
	sc "centolib/screen"
	"centolib/api"
	Mb "centolib/matchingboard"
)
func displayCoins(coins []string){
	path := os.Getenv("ORDPATH")
	i := 0
	x := 10	
	outStrXy(true,x, AppTop + i, "주문 Default Dir : %s", path)
	i++
	DrawHLine(false, 1, AppTop + i)
	i++
	for y, symbol := range coins{
		outStrXy(true, 1, y + i + AppTop, "%s", symbol)
	}
}
// CoinOrderStatus는 코인 주문 상태 정보를 화면에 출력하고 성공 여부를 반환한다.
func CoinOrderStatus()bool{
	clearArea(1, AppTop, MaxX - 1, AppBottom - AppTop)
	val, err := api.GetPlatformEnv("ORDER_LIST")	
	if err != nil{
		messageLine("%v", err)
		return false
	}
	coins, err := Mb.GetAllCoins(val)
	if err != nil{
		messageLine("%v", err)
		return false
	}
	displayCoins(coins)
	run := true
	offset := 2
	pos := 0
	MenuSelect(pos + offset, coins[pos])
	for run {
		char, err := sc.GetKey()
		if err != nil{
			messageLine("%v", err)
			os.Exit(0)
		}
		switch {
			case char == "q" || char == "Q" :
				run = false
			case char == sc.UPARR || char == sc.LTARR:
				MenuUnSelect(pos + offset, coins[pos])
				pos--
				if pos < 0{pos = len(coins) -1;}
				MenuSelect(pos + offset, coins[pos])
			case char == sc.DNARR || char == sc.RTARR:
				MenuUnSelect(pos + offset, coins[pos])
				pos++
				if pos >= len(coins){pos = 0;}
				MenuSelect(pos + offset, coins[pos])
			case char == sc.CTR_L :
				displayCoins(coins)
				MenuSelect(pos + offset, coins[pos])
			case char == sc.ENTER :
				MenuSelect(pos + offset, coins[pos])
		}
	}
	sc.CLEAR_WIN()
	sc.CURSOR_ON()
	return true
}
