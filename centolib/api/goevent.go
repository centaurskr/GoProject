///
/// @brief   Simple Event Api for Go Lang
/// @file    goevent.go
/// @date    2025. 02. 13. (목) 10:37:15 KST
/// @author  Cento
///
package api

import (
    zmq "github.com/pebbe/zmq4"
    "fmt"
    "syscall"
    "time"
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


///////////////// Event 초기화 ///////////
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
///// Add Zmq Client /////////////////////
func (p *GoEvent)AddZmqClient(socket *zmq.Socket,
        h func(*GoEvent, *zmq.Socket))error{
    fd,err := socket.GetFd()
    if err != nil{return err;}
    p.ZmqClient[int32(fd)] = ZmqEvent{
        Socket : socket,
        Handler : h,
    }
    p.Poller.Add(socket, zmq.POLLIN)
    return nil
}
///// Add Tcp Client /////////////////////
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
//////// Delete Tcp Client ////////////
func (p *GoEvent)DelTcpClient(fd int)error{
    err := syscall.EpollCtl(p.Efd, syscall.EPOLL_CTL_DEL, fd, nil)
    if err != nil{
        return fmt.Errorf("Error DelTcpClient() : %v", err)
    }
    delete(p.TcpClient, int32(fd))
	syscall.Close(fd)
    return nil
}
/////// Delete Zmq Client  /////////////
func (p *GoEvent)DelZmqClient(socket *zmq.Socket)error{
    err := p.Poller.RemoveBySocket(socket)
    if err != nil{
        return fmt.Errorf("Error DelZmqClient() : %v", err)
    }
    fd,_ := socket.GetFd()
    delete(p.ZmqClient, int32(fd))
	socket.Close()
    return nil
}
///// Event Loop Shutdown ///////////
func (p *GoEvent)GoEventDown(){
    p.Run = false
}
///// Event Loop Start ////////////
func (p *GoEvent)GoEventLoop()error{
    for p.Run{
        // Zmq poll
        if len(p.ZmqClient) > 0{
            polls, _ := p.Poller.Poll(10 * time.Millisecond)
            for _, polled :=range polls{
                fd,_ := polled.Socket.GetFd()
                item, exist := p.ZmqClient[int32(fd)]
                if !exist { continue; }
                item.Handler(p, item.Socket)
            }
        }
        if len(p.TcpClient) == 0 {continue}
        // TCP Poll
        events := make([]syscall.EpollEvent, 10)
        n, err:= syscall.EpollWait(p.Efd, events, 10)
        if err != nil{
            if err != syscall.EINTR{
                return fmt.Errorf("EpollWait error:%v (Type:%T)\n", err, err)
                //errno,_:=err.(syscall.Errno)
                //fmt.Printf("\tDetail : %s(%d)\n", errno.Error(), errno)
                //fmt.Println("\t\t interrupted by a signal, retrying...")
            }
            continue
        }
        for i:= 0; i < n; i++{
            fd := events[i].Fd
            item, exist := p.TcpClient[int32(fd)]
            if !exist { continue; }
            item.Handler(p, int(fd))
        }
    } // Main Loop
	return nil
}

