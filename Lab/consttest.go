package main
import (
	Log "centolib/logout"
	Mb "centolib/matchingboard"
	//"reflect"
	"fmt"
)

func checkBuySell(mode Mb.BuySell){
	var path, mapname string
	Log.Infor("Mode is : %s", Mb.BuySell(mode))
	path = "/HOME/"
	mapname = fmt.Sprintf("%s%s.%04d", path,  Mb.BuySell(mode), 0)
	Log.Debug("Mapname : %s", mapname)
}
func main(){
	Log.InitLogger("")
	defer Log.Close()

	checkBuySell(Mb.MEDO)
	checkBuySell(Mb.MESU)
}
