// MEMCOIN을 구성하고 구조를 Dumpout한다
//
package main
import (
	"fmt"
	Mb "centolib/matchingboard"
	"os"
)

func main(){
	if len(os.Args) != 2{
		fmt.Printf("주문(MEMCOIN) Dump Program\n")
		fmt.Printf("Using) goDumpCoin <coin>\n")	
		os.Exit(0)
	}
	ordhome := os.Getenv("ORDPATH")
	if ordhome == "" {
		fmt.Printf("ORDPATH 환경변수가 없습니다\n")
		os.Exit(0)
	}
	memcoin := &Mb.MEMCOIN{}

	err := memcoin.LoadMEMCOIN(ordhome, os.Args[1])
	if err != nil{
		fmt.Printf("LoadMEMCOIN(%s,%s)\n\t%v\n", ordhome, os.Args[1], err)
		os.Exit(0)
	}
	memcoin.Dump()
}

