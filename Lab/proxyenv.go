package main
import (
	Log "centolib/logout"
	"centolib/api"
	"os"
	"fmt"
)

func main(){
	if len(os.Args) != 2{
		fmt.Println("Proxy Env Test")	
		fmt.Println("PROXY 이름이 필요합니다.")
		fmt.Println("환경 file은 Platform.env")
		os.Exit(0)
	}
	Log.InitLogger(os.Args[1], Log.CONSOLE_MODE);defer Log.Close()
	envfile, err := api.GetEnvValue("Platform.env", "PROXYENV_YAML")
	if err != nil{
		Log.Error("GetEnvValue(Platform.env) : %v", err)
		os.Exit(0)
	}
	Log.Infor("Envfile : %s, PROXY NAME : %s", envfile, os.Args[1])
	proxy, err := api.NewProxy(envfile, os.Args[1])
	Log.Infor("%+w", proxy)
}
