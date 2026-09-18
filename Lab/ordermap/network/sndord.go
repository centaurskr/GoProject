///
/// @brief   Order packet 전송 TEST
/// @file    sndord.go
/// @date    2024. 10. 28. (월) 09:52:24 KST
/// @author  Cento
///
package main
import (
	Log "centolib/logout"
	tbzmq "centolib/tbzmq"
	Mb "centolib/matchingboard"
	"centolib/api"
)
func main() {
	
	var pub tbzmq.SockWrapper	
	
	Log.InitLogger("",Log.CONSOLE_MODE);defer Log.Close()
	
	err := pub.OpenPubSocket0mq("tcp://127.0.0.1:7905")
	if err != nil{Log.Error("%v", err);return}
	defer pub.Close()
	

	jsonStr  :=  `{"COIN":"BTC.USDT", "OCTYPE":"1", "BUYSELL":"1", "PRICE":"123.0402", "QTY":"0.12345", "MRATE":"0.001", "TRATE":"0.001", "MTRADE":"0.0001"}`
	//for {
		order := Mb.Order{}
		order.FillFromStream([]byte(jsonStr))

		order.OrdSeq = order.OrdTime

	//Log.Infor("%v", order)

		stream := api.GetBytesOfStruct(&order)
	//Log.Debug("[%d] %v", len(stream), stream)
	
		sz, err := pub.SendTopic0mq("ORDER", stream)
		if err != nil{return}
		Log.Debug("Pub Size : %d", sz)
	//}

}
