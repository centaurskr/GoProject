///
/// @brief   Network Api
/// @file    network.go
/// @date    2024. 09. 20. (금) 14:39:31 KST
/// @author  Cento
///
package api

import (
	"net"
	"fmt"
	"errors"
	"syscall"
	"strconv"
	"os"
)

func GetMyIpv4Address() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs{
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("Can't get IPV4 address")
}
////////////////////////////////////////////////////////////////////////////////
/// size + Data 쓰기 function
/// size : string 4 Bytes
/// @fn     : func WriteDataStr4(fd int, data string)(int, error)
/// @param  : socketfd , data
/// @return : size, error
////////////////////////////////////////////////////////////////////////////////
func WriteDataStr4(fd int, data string)(int, error){
    packet := fmt.Sprintf("%04d%s", len(data), data)
    size, err:=syscall.Write(fd, []byte(packet))
    if err != nil{
        return -1, err
    }
    size = size - 4
    return size, err
}
////////////////////////////////////////////////////////////////////////////////
/// size + Data 읽기 function 
/// size : string 4 Bytes
/// @fn     : func ReadDataStr4(fd int) (int, []byte, error) 
/// @param  : socket fd
/// @return :size, packet , error
////////////////////////////////////////////////////////////////////////////////
func ReadDataStr4(fd int) (int, []byte, error) {
    szStr := make([]byte, 4)
    _,err :=    syscall.Read(fd, szStr)
    if err != nil {
        return -1, nil, err
    }
    size,_ := strconv.Atoi(string(szStr[:]))

    data := make([]byte, size)
    _, err =    syscall.Read(fd, data)
    if err != nil {
        return -2, nil, err
    }

    return len(data), data, nil
}
////////////////////////////////////////////////////////////////////////////////
/// fd 기반의  Accept
/// @fn      func WaitConnectRaw(listenfd int)(int, string, error)
/// @param   listenfd : Listen Fd
/// @return  newfd : Accept fd , peername : ipv4:port, error
////////////////////////////////////////////////////////////////////////////////
func WaitConnectRaw(listenfd int)(int, string, error){
	newfd, sa, err := syscall.Accept(listenfd)
	if err != nil{
		return -1, "", fmt.Errorf("WaitConnectRaw : %v", err);
	}
	ip,_ := sa.(*syscall.SockaddrInet4);
	peer := net.IPv4(ip.Addr[0], ip.Addr[1], ip.Addr[2], ip.Addr[3]).String()
	return newfd , peer, nil
}
///////////////////////////////////////////////////////////////////////////////
/// Open Tcp Server mode - RQW Mode (socket fd)
/// @fn     : func OpenTcpServerRaw(port string)(int , error)
/// @param  : port : ex) ":32767", 주의 : ":" 추가하기
/// @return : socketfd, error
////////////////////////////////////////////////////////////////////////////////
func OpenTcpServerRaw(port string)(int, error){
    addr, err := net.ResolveTCPAddr("tcp", port)
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

    // Bind the socket to the address
    var sa syscall.SockaddrInet4
    sa.Port = int(addr.Port)
    copy(sa.Addr[:], addr.IP)
    err = syscall.Bind(fd, &sa)
    if err != nil {
        syscall.Close(fd)
        return -1, fmt.Errorf("Bind : %v", err);
    }
    err = syscall.Listen(fd, 200)
    if err != nil {
        syscall.Close(fd)
        return -1, fmt.Errorf("Listen(200) : %v", err);
    }

    return fd, nil
}
func WaitConnect(listener *net.TCPListener)(*net.TCPConn, string, error){
	conn, err := listener.Accept()
	if err != nil{
		return nil, "", fmt.Errorf("WaitConnect : %v", err);
	}
	newconn := conn.(*net.TCPConn)
	return newconn, newconn.RemoteAddr().String(), nil
}
///////////////////////////////////////////////////////////////////////////////
/// Open Tcp Server mode 
/// @fn     : func OpenTcpServer(port string)(conn interface{}, error)
/// @param  : port : ex) ":32767", 주의 : ":" 추가하기
/// @return : *net.TCPListener, error
////////////////////////////////////////////////////////////////////////////////
func OpenTcpServer(port string)(*net.TCPListener, error){
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
        return nil, fmt.Errorf("FileListener : %v", err);
    }
	listener, _ := ln.(*net.TCPListener)
    return listener, nil
}
////////////////////////////////////////////////////////////////////////////////
/// TCP Client open 
/// @fn     : func OpenTcpClient(endpoint string)(int , error){
/// @param  : endpoint : 접속 point ex) "192.168.0.2:32767"
/// @return : socket fd , error
////////////////////////////////////////////////////////////////////////////////
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
