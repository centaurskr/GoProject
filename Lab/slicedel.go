package main

import (
    "fmt"
    "slices"
)

func main() {
    CoinList := []string{"BTC.USDT", "ETH.USDT", "XRP.USDT"}
	newList := []string{"BTC.USDT", "XRP.USDT", "LTR.USDT", "PXB.USDT"}
    
	for _, coin := range newList {
		if slices.Contains(CoinList, coin) {
    		CoinList = slices.DeleteFunc(CoinList, func(c string) bool {
        		return c == coin
    		})
		}else{ // 없다
			fmt.Printf("%s는 OLD에 없음\n", coin)	
		}
   	} 
    fmt.Println(CoinList) // Output: [BTC.USDT XRP.USDT]
	fmt.Println("Length :", len(CoinList))
}
