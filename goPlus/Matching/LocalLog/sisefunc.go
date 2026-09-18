// 시세 전송을 위한  function
//
package main
import (
	zmq "github.com/pebbe/zmq4"
	Log "centolib/logout"
	"centolib/api"
    Mb "centolib/matchingboard"
	"unsafe"
	"fmt"
	"os"
	"time"
)

func gofSiseFunction(ctx *zmq.Context){
	// 내부 수신
	ipcsub, err := ctx.NewSocket(zmq.SUB)
	if err != nil{
		Log.Error("NewSocket(zmq.SUB : %v", err)	
		os.Exit(0)
	}
	Log.Infor("Start SISE. Ipc[%s]", InnerBack)
	err = ipcsub.Connect(InnerBack)
	if err != nil{Log.Error("v", err);os.Exit(0);}
	ipcsub.SetSubscribe("IPCCURR")
	ipcsub.SetSubscribe("IPCHOGA")

	// 현재가 전송
	val, err := api.GetProxyInfor(ProxyYaml, "SISEJSON", "front")
	if err != nil{
		Log.Error("SISEJSON:front : %v", err)
		os.Exit(0)
	}

	currPub , err := ctx.NewSocket(zmq.PUB)
	if err != nil{
		Log.Error("NewSocket(zmq.PUB : %v", err)	
		os.Exit(0)
	}
	err = currPub.Connect(val)
	if err != nil{Log.Error("v", err);os.Exit(0);}

	// 호가 전송
	val, err = api.GetProxyInfor(ProxyYaml, "HOGAJSON", "front")
	if err != nil{
		Log.Error("HOGAJSON:front : %v", err)
		os.Exit(0)
	}

	hogaPub , err := ctx.NewSocket(zmq.PUB)
	if err != nil{
		Log.Error("NewSocket(zmq.PUB : %v", err)	
		os.Exit(0)
	}
	err = hogaPub.Connect(val)
	if err != nil{Log.Error("v", err);os.Exit(0);}

	time.Sleep(100*time.Microsecond) 

	for {
		msg, err := ipcsub.RecvMessageBytes(0)
		if err != nil{Log.Error("%v", err);os.Exit(0);}
		topic := string(msg[0])
		switch topic {
			case "IPCHOGA" :
			// msg[0] -- IPCHOGA
            // msg[1] -- Symbol
            // msg[2] -- 호가 String

				symbol  := api.CleanString(msg[1][:])
				hogaStr := api.CleanString(msg[2][:])
				Log.Debug("IPCHOGA [%s] [%s]", symbol, hogaStr)			
				hogaPub.SendMessage(symbol, hogaStr)
			case "IPCCURR" :
			// Maker 체결 Packet으로 현재가를 생성한다
			// msg[0] "IPCCURR",  -- string
            // msg[1] InnerChePacket,    -- []byte Order
				chepk := (* Mb.InnerChePacket)(unsafe.Pointer(&msg[1][0]))
				if chepk.MakeTake == 'T' {continue;}
				// Sise Json 만들기
            	sise := fmt.Sprintf(
        "{\"sy\":\"%s\",\"t\":\"%s\",\"p\":\"%s\",\"q\":\"%s\",\"bs\":\"%c\"}",
                api.CleanString(chepk.Symbol[:]),
                time.UnixMicro(int64(chepk.CheTime)).Format("20060102150405.000"),
                chepk.Price.String(),
                chepk.Qty.String(),
                chepk.BuySell)
				Log.Debug("IPCCURR : %s", sise)	
				currPub.SendMessage(api.CleanString(chepk.Symbol[:]), sise)

		} // switch topic
	}
}
