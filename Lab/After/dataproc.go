///
/// Data 를 처리하는 go routine
///
package main
import (
    Log "centolib/logout"
    Mb "centolib/matchingboard"
	"centolib/api"
	"os"
	"time"
	"encoding/json"
	//"unsafe"
	//"fmt"
)
// Channel Message Format
/// type Message struct{ 
///    tr string
///    msg [][]byte
/// }
/// tr -> DBCHE, DBERR or DBSEQ

func goDataProcess(queue *CircularQueue){
for{
	tr, msg := queue.Read()
	if tr == "" && msg == nil {
		time.Sleep(10 * time.Microsecond) // Spin briefly if queue is empty
		continue
	}
	switch tr {
	case "DBCHE" :
		// 수신 : DBCHE + InnerChePakcet
		// 체결 내역 전송 TOPIC : USER??
		// -. 주문 정보 Update
		// 체결수량, 잔량이 없을경우 GO->CM으로 변경

	case "DBSEQ" :
		// 수신 : DBSEQ + OrderJson
		// 주문번호 갱신, 주문상태 RG->GO
		// 주문번호갱신 전송 TOPIC : USER??
		var pkjson Mb.OrderJson
		err := json.Unmarshal(msg, &pkjson)	
		if err != nil{
			Log.Error("DBSEQ : Unmarshal()%v : %s",
				err, api.CleanString(msg[:]))
			os.Exit(0)
		}
Log.Debug("DBSEQ OrdNum[%s] - DataProcess", pkjson.OrdNum)
	case "DBCANCEL" :
		// 수신 : DBCANCEL + OrderJson
		// pkjson.Margin, pkjson.Fee에는 주문증거금 해소금액이 설정 되어 있음(matching에서)
		var pkjson Mb.OrderJson
		err := json.Unmarshal(msg, &pkjson)	
		if err != nil{
			Log.Error("DBCANCEL : Unmarshal()%v : %s",
				err, api.CleanString(msg[:]))
			os.Exit(0)
		}
	case "DBERR" :
		// 수신 : DBERR + OrderJson
		var pkjson Mb.OrderJson
		err := json.Unmarshal(msg, &pkjson)	
		if err != nil{
			Log.Error("DBERR : Unmarshal()%v : %s",
				err, api.CleanString(msg[:]))
			os.Exit(0)
		}
	case "MKDONE": // 시장가주문 완료
		// 수신 MKDONE + OrderJson
		var pkjson Mb.OrderJson
		err := json.Unmarshal(msg, &pkjson)	
		if err != nil{
			Log.Error("MKDONE : Unmarshal()%v : %s",
				err, api.CleanString(msg[:]))
			os.Exit(0)
		}
Log.Debug("MKDONE : %+v", pkjson)

} // Switch
} // For Loop
} // Main Loop
