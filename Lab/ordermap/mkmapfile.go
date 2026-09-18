///
/// @brief   matchingboard package test
/// @file    mkmapfile.go
/// @date    2024. 10. 14. (월) 12:13:24 KST
/// @author  Cento
///
package main
import (
	//"fmt"
	//"centolib/api"
	Mb "centolib/matchingboard"
	Log "centolib/logout"
	//"unsafe"
	//Log "centolib/logout"
	//"sort"
)

var Coin Mb.MEMCOIN

func main(){
	Log.InitLogger("", Log.CONSOLE_MODE);defer Log.Close()

	// Make
	Coin,err := Mb.NewMEMCOIN(".", "BTC.USDT")	
	if err != nil{
		Log.Error( "%v", err)
		return
	}
	_, err = Coin.NewHoga("123.0402", Mb.MESU)
	if err != nil{
		Log.Error("%v", err)
		return
	}
	// find hoga
	phoga := Coin.FindHoga(Mb.MESU, "123.0402")
	if phoga == nil{
		Log.Error("MESU 호가 없음 : 123.0402")
	}else{
		Log.Infor("MESU 호가 찾음")
		phoga.Dump()
	}

	// Order 추가방법(Memory 추가, MapFile 증가)
	// Order 추가 후 삭제의 필요성
	//Log.Infor( "%+v", Coin)
	jsonStr  :=  `{"OCTYPE":"1", "BUYSELL":"2", "PRICE":"123.0402", "QTY":"0.12345", "MRATE":"0.001", "TRATE":"0.001", "MTRADE":"0.0001"}`
	jsonStr1 :=  `{"OCTYPE":"1", "BUYSELL":"2", "PRICE":"123.0402", "QTY":"0.98765", "MRATE":"0.002", "TRATE":"0.002", "MTRADE":"0.0002"}`

	// 주문번호 채번
	Coin.OrdSeq ++

	// 실 상황에서는 Network으로 받아야 됨
	// -. 실 상황 코드
	// var sw tbzmq.SockWrapper
	// sw.OpenSubSocket0mq(ctx, endpoint)
	// err := sub.SetSubscribe0mq("BTC.USDT")
	// topic, stream, err := sw.ReceiveTopic0mq()
	// order := (*Mb.Order)(unsafe.Pointer(&stream[0]))
	// phoga = Coin.FindHoga(BuySell(order.BuySell), order.Price)
	// err = phoga.InsertOrder(order)
	order,_ := Mb.MakeOrderFromStream([]byte(jsonStr))
	order.OrdSeq = Coin.OrdSeq
	order.RegNum = order.OrdTime
	err = phoga.InsertOrder(&order)
	if err != nil{
		Log.Error("%v", err)
	}

	// 주문번호 채번
	Coin.OrdSeq ++
	order,_ = Mb.MakeOrderFromStream([]byte(jsonStr1))
	order.OrdSeq = Coin.OrdSeq
	order.RegNum = order.OrdTime
	err = phoga.InsertOrder(&order)
	if err != nil{
		Log.Error("%v", err)
	}

	if phoga.IsWindow == false {phoga.InWindow()}

	Coin.Dump()
	phoga.Dump()
	phoga.Hd.Dump()

	phoga,err = Coin.NewHoga("124.0402", Mb.MESU)
	if err != nil{ Log.Error("%v", err); return }
	jsonStr1 =  `{"OCTYPE":"1", "BUYSELL":"2", "PRICE":"124.0402", "QTY":"0.00123", "MRATE":"0.0015", "TRATE":"0.0015", "MTRADE":"0.0001"}`
	for i:= 0; i < 20; i++{
		Coin.OrdSeq ++
		order,_ = Mb.MakeOrderFromStream([]byte(jsonStr1))
		order.OrdSeq = Coin.OrdSeq
		order.RegNum = order.OrdTime
		err = phoga.InsertOrder(&order)
		if err != nil {
			Log.Error("%v", err)
			break;
		}
	}

	phoga,err = Coin.NewHoga("249.9999", Mb.MEDO)
	if err != nil{ Log.Error("%v", err); return }
	phoga,err = Coin.NewHoga("248.8888", Mb.MEDO)
	if err != nil{ Log.Error("%v", err); return }
	phoga,err = Coin.NewHoga("247.7777", Mb.MEDO)
	if err != nil{ Log.Error("%v", err); return }
	phoga,err = Coin.NewHoga("246.6666", Mb.MEDO)
	if err != nil{ Log.Error("%v", err); return }

	Coin.DumpHoga()

	/// DELETE TEST
	err = Coin.DeleteHogaByPrice(Mb.MEDO, "248.8888")
	if err != nil {Log.Error("%v", err)}
	err = Coin.DeleteHogaByPrice(Mb.MESU, "123.0402")
	if err != nil {Log.Error("%v", err)}
	err = Coin.DeleteHogaByPrice(Mb.MESU, "124.0402")
	if err != nil {Log.Error("%v", err)}

	Coin.SetWindow(Mb.MEDO)
	Coin.SetWindow(Mb.MESU)
	Log.Infor("=== 최종 Dump =====")
	Coin.DumpHoga()
/********************
	// HOGA 추가
	jsonStr  :=  `{"OCTYPE":"1", "BUYSELL":"2", "PRICE":"124.0402", "QTY":"0.12345", "MRATE":"0.001", "TRATE":"0.001", "MTRADE":"0.0001"}`
	// find hoga
	hoga := Coin.FindHoga(Coin.Mesu, "124.0402")
	if hoga == nil{
		err = Coin.NewHoga("124.0402", Mb.MESU)
		Log.Error("MESU 호가 없음 : 124.0402")
	}else{
		Log.Infor("MESU 호가 찾음")
	}

	// Test 해야 함: Network으로 Order 송수신 -- O
	// ./network에 sample 있음
	
	// Create Other Medo
	// Sort Medo
	// Create New Mesu
	// Create Other Medo
	// Sort Medo

	// Window Operation

	// Day Operation

**************/
}
