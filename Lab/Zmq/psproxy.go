package main
import (
	Log "centolib/logout"
	zmq "github.com/pebbe/zmq4"
	"centolib/tbzmq"
)
func main(){
	Log.InitLogger("", Log.CONSOLE_MODE);defer Log.Close()
	Log.Infor("Proxy library test")
	ctx, _ := zmq.NewContext()
	err := tbzmq.RunPubSubProxy(ctx,
		"tcp://127.0.0.1:32768",
		"tcp://127.0.0.1:32769")
	if err != nil{
		Log.Error("RuncProxyPubSub : %v", err)
		return 
	}
	Log.Infor("tcp://127.0.0.1:32768 -> tcp://127.0.0.1:32769")
	select{}
}

