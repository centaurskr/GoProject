///
/// @brief  : Order packet 수신 TEST(Proxy 이용)
/// @file   : rcvord.go
/// @date   : 2024. 10. 28. (월) 12:00:07 KST
/// @author : Cento
///
package main
import (
    Log "centolib/logout"
    //zmq "github.com/pebbe/zmq4"
    tbzmq "centolib/tbzmq"
    Mb "centolib/matchingboard"
	"unsafe"
    //"centolib/api"
)
func main() {

	Log.InitLogger("", Log.CONSOLE_MODE);defer Log.Close()

    var sub tbzmq.SockWrapper
	err := sub.OpenSubSocket0mq("tcp://127.0.0.1:7908")
	if err != nil{Log.Error("tcp://127.0.0.1:7908");return}
	defer sub.Close()

	err = sub.SetSubscribe0mq("")
	//err = sub.SetSubscribe0mq("")
	if err != nil{Log.Error("%v", err);return}
	for {
		topic, stream, err := sub.ReceiveTopic0mq()
		if err != nil{Log.Error("%v", err);return}
	
		Log.Infor("TOPIC[%s]", topic)
		//Log.Infor("[%d] %v", len(stream), stream)
		order := (*Mb.Order)(unsafe.Pointer(&stream[0]))
		order.Dump(1)
	}	
}

