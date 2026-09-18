// 증거금 정리 프로세스 메인 파일.
// REALJSON Proxy에서 체결/취소 데이터를 수신하여 증거금을 정산한다.
package main

import (
	zmq "github.com/pebbe/zmq4"
    Log "centolib/logout"
    //Mb "centolib/matchingboard"
	"centolib/api"
	//"strconv"
	"os"
	"time"
	"fmt"
	"database/sql"
)
var (
	SubSocket *zmq.Socket
	DbContext *sql.DB
)
type Message struct{ // Channel Message Format
    tr string
    msg [][]byte
}
var Channel chan Message

//////////// Data Handler ////////////////////////////////////
func dataHandler(ent *api.GoEvent, socket *zmq.Socket){
    msg, err := socket.RecvMessageBytes(0)
    // msg[0] -> USER?? (Topic)
    // msg[1] -> DBCHE, DBERR or DBSEQ
    // msg[2] -> []byte data
    if err != nil{ return ; }
	packet := Message{
		tr : string(msg[1][:]),
	}
	for i := 2; i < len(msg); i++ {
		packet.msg = append(packet.msg, msg[i])
	}
	Channel <- packet
	return
}
/////////////////////////////////// Main //////////////////////
func main(){
	
	if len(os.Args) != 2{
		fmt.Printf("Using) goWallet <mode>\n")
		os.Exit(0)
	}
    switch os.Args[1] {
        case "file" :
        Log.InitLogger("", Log.FILE_MODE);
        case "console" :
        Log.InitLogger("", Log.CONSOLE_MODE);
        case "all" :
        Log.InitLogger("", Log.ALL_MODE);
        default :
        Log.InitLogger("", Log.CONSOLE_MODE);
    }   
    defer Log.Close()

	api.UsingMaxProcess()

    // 중복실행 check
    isrun, err := api.IsRunning("goWallet")
    if isrun == true {
        Log.Error("goWallet 은 이미 수행중입니다")
        return
    }

    // DB 처리 후 Data 수신 Proxy 환경설정
    yamlFile ,err := api.GetPlatformEnv("PROXYENV_YAML")
    if err != nil{
        Log.Error("%v", err)
        os.Exit(0)
    }
	val, err := api.GetProxyInfor(yamlFile, "REALJSON", "back")
	if err != nil{
		Log.Error("%s에 REALJSON:back이 없습니다", yamlFile)
		os.Exit(0)
	}

	ctx , _ := zmq.NewContext()

	sub, err := ctx.NewSocket(zmq.SUB)
	if err != nil{
		Log.Error("Sub socket : %v", err)
		os.Exit(0)
	}
	SubSocket = sub
	SubSocket.Connect(val)
	defer SubSocket.Close()
	time.Sleep(100*time.Microsecond) // for good zmq connection
	SubSocket.SetSubscribe("") // 모두

    // Db Connect
    dbyaml,err := api.GetPlatformEnv("DBCONFIG_YAML")
    if err != nil{
        Log.Error("%s", err)
        os.Exit(0)
    }
    DbContext, err = api.ConnectDb(dbyaml)
    if err  != nil{
        Log.Error("ConnectDb(%s):%v", dbyaml, err)
        os.Exit(0)
    }
	defer DbContext.Close()

	event := &api.GoEvent{}
	err = event.InitGoEvent()
	if err != nil{
		Log.Error("InitGoEvent : %v ", err)	
		os.Exit(0)
	}

	Channel = make(chan Message, 5000)
	go goDataProcess()

	Log.Infor("Start ")

	event.AddZmqClient(SubSocket, dataHandler)

	event.GoEventLoop()
}
