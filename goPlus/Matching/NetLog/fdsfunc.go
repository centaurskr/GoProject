// FDS 처리를 위한 function
//
package main
import (
	zmq "github.com/pebbe/zmq4"
	Log "centolib/netlog"
    //Mb "centolib/matchingboard"
	//"encoding/json"
	"encoding/binary"
	"bytes"
	//"time"
	"os"
)

// FDS처리를 위한 Go function
func gofFdsFunction(ctx *zmq.Context){
	ipcsub, err := ctx.NewSocket(zmq.SUB)
	if err != nil{
		Log.Error("NewSocket(zmq.SUB : %v", err)	
		os.Exit(0)
	}
	Log.Infor("Start FDS. Ipc[%s]", InnerBack)
	err = ipcsub.Connect(InnerBack)
	if err != nil{Log.Error("v", err);os.Exit(0);}

	ipcsub.SetSubscribe("IPCFDS")

	for {
		// TOPIC + ORDER + HOGA
		msg, err := ipcsub.RecvMessageBytes(0)
		//if err != nil{Log.Error("%v", err);continue;}
		if err != nil{Log.Error("%v", err);os.Exit(0);}
		// msg[0] "IPCFDS", 
		// msg[1] memcoin.MstInfor.BfPrice.String()
		// msg[2] memcoin.MstInfor.BfTime
		// msg[3] streamjson
		// msg[4] streamhoga

		var bfTime	 int64
 		binary.Read(bytes.NewReader(msg[2]), binary.LittleEndian, &bfTime)
/********************************
		var (
			jsonpk Mb.OrderJson
		)
		err = json.Unmarshal(msg[3], &jsonpk)

		Log.Debug("IPCFDS-Topic %s", string(msg[0]))
		Log.Debug("IPCFDS BfPrice :%s", msg[1])
		Log.Debug("IPCFDS BfTime  : %d", bfTime)
		Log.Debug("IPCFDS Json : %+v", jsonpk)
		Log.Debug("IPCFDS hoga : %s", string(msg[4]))
********************/
	}
}
