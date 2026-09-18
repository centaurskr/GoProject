package main
import (
	"centolib/api"
	zmq "github.com/pebbe/zmq4"
	"fmt"
	"syscall"
	"time"
	"os"
)
type TcpEvent struct{
	Fd   int
	Handler  func(*GoEvent, int)
}
type ZmqEvent struct{
	Socket *zmq.Socket
	Handler func(*GoEvent, *zmq.Socket)
}

type GoEvent struct{
	Run  bool
	Efd  int    // for epoll
	Poller *zmq.Poller // for zmq poller
	TcpClient map[int32]TcpEvent
	ZmqClient map[int32]ZmqEvent
}

func (p *GoEvent)InitGoEvent()error{
	efd , err := syscall.EpollCreate1(0)
	if err != nil {return err;}
	p.Efd = efd
	p.Poller = zmq.NewPoller()
	p.TcpClient = make(map[int32]TcpEvent)
	p.ZmqClient = make(map[int32]ZmqEvent)
	p.Run = true
	return nil
}
func (p *GoEvent)AddZmqClient(socket *zmq.Socket, 
		h func(*GoEvent, *zmq.Socket))error{
	fd,err := socket.GetFd()
	if err != nil{return err;}
	p.ZmqClient[int32(fd)] = ZmqEvent{
		Socket : socket,
		Handler : h,
	}
	return nil
}
func (p *GoEvent)AddTcpClient(fd int, h func(*GoEvent, int))error{
	p.TcpClient[int32(fd)] = TcpEvent{
		Fd  : fd,
		Handler : h,
	}
	event := syscall.EpollEvent{
		Events : syscall.EPOLLIN,
		Fd     : int32(fd),
	}
	err := syscall.EpollCtl(p.Efd, syscall.EPOLL_CTL_ADD, fd, &event)
	if err != nil{ return fmt.Errorf("AddClient : %v", err);}
	return nil
}
func (p *GoEvent)DelTcpClient(fd int)error{
	err := syscall.EpollCtl(p.Efd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil{
		return fmt.Errorf("Error DelTcpClient() : %v", err)
	}
	delete(p.TcpClient, int32(fd))
	return nil
}
func (p *GoEvent)DelZmqClient(socket *zmq.Socket)error{
	err := p.Poller.RemoveBySocket(socket)
	if err != nil{
		return fmt.Errorf("Error DelZmqClient() : %v", err)
	}
	fd,_ := socket.GetFd()
	delete(p.ZmqClient, int32(fd))
	return nil
}
func (p *GoEvent)GoEventDown(){
	p.Run = false
}
func (p *GoEvent)GoEventLoop(){
	for p.Run{
		// Zmq poll
		if len(p.ZmqClient) > 0{
			polls, _ := p.Poller.Poll(10 * time.Millisecond)	
			for _, polled :=range polls{
				fd,_ := polled.Socket.GetFd()
				item, exist := p.ZmqClient[int32(fd)]
				if !exist {
					fmt.Printf("없는 zmq.Socket입니다.FD[%d]\n", fd)
					continue
				}
				item.Handler(p, item.Socket)
			}
		}
		if len(p.TcpClient) == 0 {continue}
		// TCP Poll
		events := make([]syscall.EpollEvent, 10)
		n, err:= syscall.EpollWait(p.Efd, events, 10)
		if err != nil{
			if err != syscall.EINTR{
				fmt.Printf("EpollWait error : %v (Type : %T)\n", err, err)
				errno,_:=err.(syscall.Errno)
				fmt.Printf("\tDetail : %s(%d)\n", errno.Error(), errno)
				fmt.Println("\t\t interrupted by a signal, retrying...")
			}
			continue
		}
		for i:= 0; i < n; i++{
			fd := events[i].Fd
			item, exist := p.TcpClient[int32(fd)]
			if !exist {
				fmt.Printf("없는 list 입니다. fd[%d]\n", fd)
				continue
			}
			item.Handler(p, int(fd))
		}
	} // Main Loop
}

func ClientHandler(event *GoEvent, socket int){
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
func ConnectHandler(event *GoEvent, lfd int){
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
func main(){
	if len(os.Args) < 3{
		fmt.Printf("TCP and zmq sub Server Program- GoEvent Version\n")
		fmt.Printf("eventsvr <tcpport> <subEndpoint>\n")
		fmt.Printf(" <tcpport> -> ex) \":32767\"\n")
		os.Exit(0)
	}
	event := &GoEvent{}
	event.InitGoEvent()

	listenfd, err := api.OpenTcpServerRaw(os.Args[1])
	if err != nil {
		fmt.Printf("ERR : %v\n", err)
		os.Exit(0)
	}

	socket, _:= zmq.NewSocket(zmq.SUB)
	socket.Bind("tcp://127.0.0.1:32769")
	socket.SetSubscribe("")

	event.AddTcpClient(listenfd, ConnectHandler)

	event.GoEventLoop()
}
