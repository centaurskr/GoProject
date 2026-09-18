package main
import ( 
	"fmt"
	"strings"
)

func main(){
	Exrate := map[string]interface{}{
		"USDT" : "0.1",
		"KRW"  : "1.0",
	}
	coin := "BTC.USDT"
	slice := strings.Split(coin, ".")
	fmt.Println(slice[0], slice[1])
	currency := slice[1]
	Exrate[currency] = "1234.56"
	for key, value := range Exrate{
		fmt.Println(key, value)
	}
}
