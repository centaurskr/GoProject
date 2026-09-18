// Data 를 처리하는 go routine
//
package main
import (
    Log "centolib/logout"
    Mb "centolib/matchingboard"
	"centolib/api"
	"os"
	"strconv"
	"encoding/json"
	"unsafe"
	"fmt"
)
// Channel Message Format
// type Message struct{ 
//    tr string
//    msg [][]byte
// }
// tr -> DBCHE, DBERR or DBSEQ

func goDataProcess(queue *CircularQueue){
for{
	<-queue.readCh
for{
	qData, ok := queue.Read()
	if !ok { break; }

	for _, pk := range qData{

	switch pk.tr {
	case "DBCHE" :
		// 수신 : DBCHE + InnerChePakcet
		// 체결 내역 전송 TOPIC : USER??
		// -. 주문 정보 Update
		// 체결수량, 잔량이 없을경우 GO->CM으로 변경

		// 체결내역 Insert
		err:= DbInsertChe(DbContext, 
				(*Mb.InnerChePacket)(unsafe.Pointer(&pk.data[0])))
		if err != nil{ 
			Log.Error("DbInsertChe() :%v:%v ",err, 
				(*Mb.InnerChePacket)(unsafe.Pointer(&pk.data[0])))
			//os.Exit(0)
		}
		err = DbUpdateOrder(DbContext, 
				(*Mb.InnerChePacket)(unsafe.Pointer(&pk.data[0])))
		if err != nil {
			Log.Error("DbUpdateOrder() :%v:%v ",err, 
				(*Mb.InnerChePacket)(unsafe.Pointer(&pk.data[0])))
			//os.Exit(0)
		}
		chepk := (*Mb.InnerChePacket)(unsafe.Pointer(&pk.data[0]))
		topicStr := fmt.Sprintf("USER%02d", chepk.UserId % 10)	

		// InnnerChePacket 을 JSON(JsonChePacket)으로 변경 
		jsonStr,err := json.Marshal( MakeJsonChePacket(chepk))
		if err != nil {
			Log.Error("Marshal(%v) : %v", err, chepk)
		}else{
			PubReal.SendMessage(topicStr, "DBCHE",  jsonStr)
		}
	case "DBSEQ" :
		// 수신 : DBSEQ + OrderJson
		// 주문번호 갱신, 주문상태 RG->GO
		// 주문번호갱신 전송 TOPIC : USER??
		var pkjson Mb.OrderJson
		err := json.Unmarshal(api.CleanByte(pk.data[:]), &pkjson)	
		if err != nil{
			Log.Error("DBSEQ : Unmarshal()%v : %s",
				err, api.CleanString(pk.data[:]))
			//os.Exit(0)
		}
		err = DbUpdateOrdNum(DbContext, pkjson)
		if err != nil {
			Log.Error("DbUpdateOrdNum() :%v:%s ",err, 
				api.CleanString(pk.data[:]))
			//os.Exit(0)
		}
		uid, _:= strconv.Atoi(pkjson.UserId)
		topicStr := fmt.Sprintf("USER%02d", uid % 10)
		PubReal.SendMessage(topicStr, "DBSEQ", pk.data)
	case "DBCANCEL" :
		// 수신 : DBCANCEL + OrderJson
		// pkjson.Margin, pkjson.Fee에는 주문증거금 해소금액이 설정 되어 있음(matching에서)
		var pkjson Mb.OrderJson
		err := json.Unmarshal(api.CleanByte(pk.data[:]), &pkjson)	
		if err != nil{
			Log.Error("DBCANCEL : Unmarshal()%v : %s",
				err, api.CleanString(pk.data[:]))
			os.Exit(0)
		}
		err = DbCancelOrder(DbContext, pkjson)
		if err != nil {
			Log.Error("DbCancelOrder() :%v:%s ",err, 
				api.CleanString(pk.data[:]))
			// 이경우 사용자에게는 Error가 전달되고(아래에서)
			// DB에는 Error 표시가 안된 상태임....
		}
		uid, _:= strconv.Atoi(pkjson.UserId)
		topicStr := fmt.Sprintf("USER%02d", uid % 10)
		PubReal.SendMessage(topicStr, "DBCANCEL", pk.data)
	case "DBERR" :
		// 수신 : DBERR + OrderJson
		var pkjson Mb.OrderJson
		err := json.Unmarshal(api.CleanByte(pk.data[:]), &pkjson)	
		if err != nil{
			Log.Error("DBERR : Unmarshal()%v : %s",
				err, api.CleanString(pk.data[:]))
			os.Exit(0)
		}
		err = DbUpdateError(DbContext, pkjson)
		if err != nil {
			Log.Error("DbUpdateError() :%v:%s ",err, 
				api.CleanString(pk.data[:]))
			// 이경우 사용자에게는 Error가 전달되고(아래에서)
			// DB에는 Error 표시가 안된 상태임....
		}
		uid, _:= strconv.Atoi(pkjson.UserId)
		topicStr := fmt.Sprintf("USER%02d", uid % 10)
		PubReal.SendMessage(topicStr, "DBERR", pk.data)
	case "MKDONE": // 시장가주문 완료
		// 수신 MKDONE + OrderJson
		var pkjson Mb.OrderJson
		err := json.Unmarshal(api.CleanByte(pk.data[:]), &pkjson)	
		if err != nil{
			Log.Error("MKDONE : Unmarshal()%v : %s",
				err, api.CleanString(pk.data[:]))
			os.Exit(0)
		}
		err = DbUpdateMkDone(DbContext, pkjson)
		if err != nil {
			Log.Error("DbUpdateMkDone() :%v:%s ",err, 
				api.CleanString(pk.data[:]))
			// 이경우 사용자에게는 Error가 전달되고(아래에서)
			// DB에는 Error 표시가 안된 상태임....
		}
Log.Debug("MKDONE : %+v", pkjson)
		uid, _:= strconv.Atoi(pkjson.UserId)
		topicStr := fmt.Sprintf("USER%02d", uid % 10)
		PubReal.SendMessage(topicStr, "MKDONE", pk.data)

} // Switch
}// for range
}// Queue channel
} // For Loop
} // Main Loop
