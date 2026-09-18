package main
import (
	"net"
	"syscall"
	"fmt"
	"os"
	//"time"
)
type TcpHandler struct{
	Conn     interface{}
	Args     []string
	Handler  func(env *GoTcpEvent, conn interface{}, args []string)
}
type GoTcpEvent struct{
	Run      bool
	EpollFd  int
	Connections map[int]TcpHandler
}
func InitEvent()(*GoTcpEvent, error){
	epollFd , err := syscall.EpollCreate1(0)
	if err != nil { return nil, err;}
	event := &GoTcpEvent{
		Run         : true,
		EpollFd     : epollFd,
		Connections : make(map[int]TcpHandler),
	}
	return event, nil
}
func (p *GoTcpEvent)AddEvent(conn interface{}, 
				handler func (p *GoTcpEvent, conn interface{}, arg []string),
				args ...string)error{
	var fd int
	switch v := conn.(type){
		case *net.TCPListener :
			//file , err := v.(*net.TCPListener).File()
			file , err := v.File()
			if err != nil { return fmt.Errorf("AddEvent.TCPListener : %v",err);}
			fd = int(file.Fd())
			//file.Close()
		case *net.TCPConn:
			file,err := v.File()
			if err != nil { return fmt.Errorf("AddEvent.TCPConn : %v", err);}
			fd = int(file.Fd())
			//file.Close()
		default :
			return fmt.Errorf("지원하지 않는 connection Type입니다. %T", v)
	}
		
	event := &syscall.EpollEvent{
		Events : syscall.EPOLLIN,
		Fd     : int32(fd),
	}
	err := syscall.EpollCtl(p.EpollFd, 
				syscall.EPOLL_CTL_ADD, fd, event)
	if err !=nil{return err;} 
	p.Connections[fd]=TcpHandler{
		Conn    : conn,
		Args    : args,
		Handler : handler,
	}
	fmt.Printf("Add : %v\n", p.Connections[fd])
	return nil
}
func (p *GoTcpEvent)DelEvent(conn interface{})error{
	var fd int
	switch v := conn.(type){
		case *net.TCPListener :
			file , err := v.File()
			if err !=nil{return err;} 
			fd = int(file.Fd())
			file.Close()
		case *net.TCPConn:
			file,err := v.File()
			if err != nil { return err;}
			fd = int(file.Fd())
			file.Close()
	}
	if closer, ok := conn.(interface{ Close() error }); ok {
	// If the connection has a Close method, call it
		if err := closer.Close(); err != nil {
				return fmt.Errorf("error closing connection: %v", err)
			}
	}
	err := syscall.EpollCtl(p.EpollFd, 
				syscall.EPOLL_CTL_DEL, int(fd), nil)
	if err != nil{return err;}
	delete(p.Connections, fd)
	return nil
}
func (p *GoTcpEvent)EventLoop(){
	for p.Run == true {
		polls := make([]syscall.EpollEvent, 100)
		n, _ := syscall.EpollWait(p.EpollFd, polls, -1)
		for i := 0; i < n; i++{
			fd := int(polls[i].Fd)
			item, exist := p.Connections[fd]
			if exist {
				item.Handler(p, item.Conn, item.Args)
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
/// Open Tcp Server mode 
/// @fn     : func OpenTcpServer(port srring)(conn interface{}, error)
/// @param  :
/// @return :
////////////////////////////////////////////////////////////////////////////////
func OpenTcpServer(port string)(interface{}, error){
	addr, err := net.ResolveTCPAddr("tcp", port)
	if err != nil{ return nil, fmt.Errorf("ResolveTCPAddr : %v", err);}

	// Create a raw socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil { return nil, fmt.Errorf("Socket : %v",err);}

	// Set SO_REUSEADDR, SO_KEEPALIVE option
	err = syscall.SetsockoptInt(fd,syscall.SOL_SOCKET,syscall.SO_REUSEADDR, 1)
	if err != nil { 
		syscall.Close(fd)
		return nil, fmt.Errorf("REUSEADDR : %v" ,err);
	}

	err = syscall.SetsockoptInt(fd,syscall.SOL_SOCKET,syscall.SO_KEEPALIVE, 1)
	if err != nil { 
		syscall.Close(fd)
		return nil, fmt.Errorf("KEEPALIVE : %v" ,err);
	}		

	// Bind the socket to the address
	var sa syscall.SockaddrInet4
	sa.Port = int(addr.Port)
	copy(sa.Addr[:], addr.IP)
	err = syscall.Bind(fd, &sa)
	if err != nil { 
		syscall.Close(fd)
		return nil, fmt.Errorf("Bind : %v", err);
	}		

	err = syscall.Listen(fd, 200)
	if err != nil {
		syscall.Close(fd)
		return nil, fmt.Errorf("Listen(200) : %v", err);
	}		

	ln, err := net.FileListener(os.NewFile(uintptr(fd), "listener"))
	if err != nil { 
		syscall.Close(fd)
		return nil, fmt.Errorf("FileListenr : %v", err);
	}

	return ln, nil
}
func OpenTcpClient(endpoint string)(int , error){
	addr, err := net.ResolveTCPAddr("tcp", endpoint)
	if err != nil{ return -1, fmt.Errorf("ResolveTCPAddr : %v", err);}
    // Create a raw socket
    fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
    if err != nil { return -1, fmt.Errorf("Socket : %v",err);}

    // Set SO_REUSEADDR, SO_KEEPALIVE option
    err = syscall.SetsockoptInt(fd,syscall.SOL_SOCKET,syscall.SO_REUSEADDR, 1)
    if err != nil { 
		syscall.Close(fd)
		return -1, fmt.Errorf("REUSEADDR : %v" ,err);
	}

    err = syscall.SetsockoptInt(fd,syscall.SOL_SOCKET,syscall.SO_KEEPALIVE, 1)
    if err != nil { 
		syscall.Close(fd)
		return -1, fmt.Errorf("KEEPALIVE : %v" ,err);
	}		

    var sa syscall.SockaddrInet4
    sa.Port = int(addr.Port)
    copy(sa.Addr[:], addr.IP)

	err = syscall.Connect(fd, &sa)
	if err != nil && err != syscall.EINPROGRESS {
		syscall.Close(fd)
		return -1, fmt.Errorf("Connect failed: %v", err)
	}
	return fd, nil
}
func ClientHandler(ent *GoTcpEvent , conn interface{}, args []string){
	fmt.Printf("Tcp Client handler ----- %v\n", args)
}
func ConnectHandler(ent *GoTcpEvent , conn interface{}, args []string){
	l := conn.(*net.TCPListener)
	newcon, err:= l.AcceptTCP()
	if err !=nil{ return; }

	addr := newcon.RemoteAddr()
	fmt.Printf("TcpConnect handler - %vClient Address[%v]\n", args, addr)
	
	err = ent.AddEvent(newcon, ClientHandler, fmt.Sprintf("%v", addr))
	if err != nil{
		fmt.Printf("ERR: %v\n", err)
		newcon.Close()
	}
	return
}
func main(){
	event, err:= InitEvent()
	if err != nil{
		fmt.Printf("ERROR : %v\n", err)
		os.Exit(0)
	}

	l, err:= OpenTcpServer(":32768")
	if err != nil{
		fmt.Printf("OpenTcpServer : %v\n", err)
		os.Exit(0)
	}
	err = event.AddEvent(l, ConnectHandler, "tcp", ":32768")
	if err != nil{
		fmt.Printf("AddEvent : %v\n", err)	
		os.Exit(0)
	}

	fmt.Printf("Server Mode Start! port(:32768)\n")
	event.EventLoop()
}
