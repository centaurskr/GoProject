package main
import (
	"fmt"
	"centolib/api"
	zmq "github.com/pebbe/zmq4"
	"os"
	"time"
)
var (
	SeqCount, CheCount int
)
// DataHandler는 AFTERMATCHING Proxy에서 수신한 시퀀스 데이터를 처리한다.
func DataHandler(ent *api.GoEvent, socket *zmq.Socket){
	msg, err := socket.RecvMessageBytes(0)
	if err != nil { return; }
	// msg[0] -- AFTER??
	// msg[1] --- DBCHE, DBERR, DBSEQ..
	// msg[2] -- Data
	switch string(msg[1]){
		case "DBSEQ" :
			SeqCount ++;
		case "DBCHE" :
			CheCount ++
	}
	fmt.Printf("Count DBSEQ(%d) + DBCHE(%d) = (%d)\n",
		SeqCount, CheCount, SeqCount + CheCount)
}
func main(){
	fmt.Println("AFTERMATCHING에서 전송하는 Packet 수신하는 Program")
	fmt.Println("DBSEQ, DBCHE의 갯수를 세기 위한 목적으로만든다")
	fmt.Printf("\n\n")
	yamlFile, err := api.GetPlatformEnv("PROXYENV_YAML")
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}
	sub, err := zmq.NewSocket(zmq.SUB)
	val, err := api.GetProxyInfor(yamlFile, "AFTERMATCHING", "back")
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}
	defer sub.Close()
	sub.Connect(val)
	sub.SetSubscribe("")
	time.Sleep(1000 * time.Millisecond)

	event := &api.GoEvent{}
	err = event.InitGoEvent()
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}

	SeqCount = 0
	CheCount = 0

	event.AddZmqClient(sub, DataHandler)
	event.GoEventLoop()
}
