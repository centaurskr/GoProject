//
// Lab : Matching을 위한 Coin을 Memory에 Loading
//
package main
import (
	Log "centolib/logout"
	Mb "centolib/matchingboard"
	//"centolib/api"
	"os"
)

func main(){
	Log.InitLogger("", Log.CONSOLE_MODE);defer Log.Close()

	if len(os.Args) != 2{
		Log.Error("MEMCOIN Load test")
		Log.Infor("%s <coin>", os.Args[0])
		os.Exit(1)
	}
	
	ordhome := os.Getenv("ORDPATH")
	if ordhome == "" {
		Log.Error("환경변수에 ORDPATH가 없습니다")
		os.Exit(1)
	}

	memcoin := &Mb.MEMCOIN{}
	memcoin.LoadMEMCOIN(ordhome, os.Args[1])
	
	memcoin.Dump()
}
