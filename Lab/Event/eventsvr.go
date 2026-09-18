package main
import (
	"centolib/api"
	zmq "github.com/pebbe/zmq4"
	"fmt"
	"os"
)

func ClientHandler(event *api.GoEvent, socket int){
	sz, packet, err := api.ReadDataStr4(socket)
	if sz <= 0{
		fmt.Printf("Error ReadDataStr()-%d-%v\n", sz, err)
		event.DelTcpClient(socket)
		return
	}
	fmt.Printf("수신[%d] : %s\n", sz, string(packet))
	sz, err = api.WriteDataStr4(socket, string(packet))
	if err != nil{
		fmt.Println("Error Write ", err)
		return
	}
	fmt.Println("송신 : ", sz)
	return
}
func ConnectHandler(event *api.GoEvent, lfd int){
	newsock, addr, err := api.WaitConnectRaw(lfd)
	if err != nil{
		fmt.Printf("AcceptTCP : %v\n", err)
		return
	}
	fmt.Printf("Connect from : %s\n", addr)
	// Add to Client List
	event.AddTcpClient(newsock, ClientHandler)
	return
}
func ZmqSubHandler(event *api.GoEvent, socket *zmq.Socket){
	msg, err := socket.RecvMessage(0)
	if err !=nil {
		fmt.Println("ZmqSubHandler ", err)
		return
	}
	fmt.Println(msg)
	return
}
func main(){
	if len(os.Args) < 3{
		fmt.Printf("TCP and zmq sub Server Program- GoEvent Version\n")
		fmt.Printf("eventsvr <tcpport> <subEndPoint>\n")
		fmt.Printf(" <tcpport>    : tcp Connect port\n")
		fmt.Printf(" <subEndPoint>: zmq.SUB connect point\n")
		fmt.Printf("ex) eventsvr :32768 tcp://127.0.0.1:32769\n")
		os.Exit(0)
	}
	event := &api.GoEvent{}
	event.InitGoEvent()

	// TCP Listener
	listenfd, err := api.OpenTcpServerRaw(os.Args[1])
	if err != nil {
		fmt.Printf("ERR : %v\n", err)
		os.Exit(0)
	}
	event.AddTcpClient(listenfd, ConnectHandler)

	// zmq SUB
	socket, _:= zmq.NewSocket(zmq.SUB)
	err = socket.Bind(os.Args[2])
	if err != nil{
		fmt.Printf("zmq Bind(%s) : %v\n", os.Args[2], err)
		os.Exit(0)
	}
	socket.SetSubscribe("")
	event.AddZmqClient(socket, ZmqSubHandler)

	fmt.Println("Start!!!\nTCP and Zmq sub Test Server - GoEvent Version")
	fmt.Printf("TCP port -> %s, zmq PUB point ->%s\n", os.Args[1], os.Args[2])

	event.GoEventLoop()
}
