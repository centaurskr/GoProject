///
/// @brief  GO TCP Server
/// @file   : ordbrg.go
/// @date   : 2024. 09. 11. (수) 11:34:01 KST
/// @author : Cento
///

package main

import (
	Log "centolib/logout"
	"net"
	"os"
	"io"
	"runtime"
	"errors"
	zmq "github.com/pebbe/zmq4"
)

var endPoint string

var context *zmq.Context

////////////////////////////////////////////////////////////////////////////////
///
/// @fn     : func ClientHandler(socket net.Conn){
/// @param  :
/// @return :
////////////////////////////////////////////////////////////////////////////////
func ClientHandler(c *net.TCPConn){

	defer c.Close()

	buff1 := make([]byte, 2048)
	buff2 := make([]byte, 2048)

	for {
		// --------------TCP  수신
		ln, err := c.Read(buff1)
		if err != nil { 
			if errors.Is(err, io.EOF) {
				Log.Infor("Connection closed by %s", c.RemoteAddr())
			}else{
				Log.Error("%v", err);
			}
			return 
		}

		//Log.Infor("수신 DATA :[%d] %s", ln, string(buffer[:ln]))

		// 1. Zmq REQ Connect
		zsock, err := context.NewSocket(zmq.REQ)
		if err != nil {
			Log.Error("zmq.REQ %v", err)
			return	
		}
		defer zsock.Close()
		err = zsock.Connect(os.Args[2])
		if err != nil{
			Log.Error("Connect(%s)\n\t%v", os.Args[2], err)
			return
		}
		// 2. Zmq Data 전송
		ln, err = zsock.SendBytes(buff1[:ln], 0)
		if err != nil{
			Log.Error("Send(%s)\n\t%v", string(buff1[:ln]), err)
			return
		}
		// 3. Zmq Data 수신
		buff2, _= zsock.RecvBytes(0)

		// -----------TCP 응답 전송
		ln, err = c.Write(buff2)
		if err!= nil{
			Log.Error("Write Error : %v", err)
			return
		}
		//Log.Infor("Write size[%d]", ln)
	}
}
////////////////////////////////////////////////////////////////////////////////
///
/// @fn     : func main(){
/// @param  :
/// @return :
////////////////////////////////////////////////////////////////////////////////
func main(){
	Log.InitLogger("", Log.CONSOLE_MODE);defer Log.Close()
	if len(os.Args) != 3 {
		Log.Error("GO TCP server")
		Log.Error("USING : %s <port> <broker zmq point>", os.Args[0])
		os.Exit(0)
	}

	runtime.GOMAXPROCS(4) // runtime pkg. max 4개의 CPU 사용
	Log.Infor("OS(%v) ARCH(%v) CPUS(%d)", runtime.GOOS, runtime.GOARCH,runtime.NumCPU) 

	// Zmq Context 
	context, _ = zmq.NewContext()

	// TCP Listen
	port := ":" + os.Args[1]
	l, err := net.Listen("tcp", port)
	if err != nil {
		Log.Error("%v", err)
		Log.Error(" or Check <port>, must be number without colon!")
		os.Exit(0)
	}
	defer l.Close()

	// Zmq Proxy REQ Endpoint
	endPoint = os.Args[2]

	Log.Infor("Start OrdBrg port[%s] endpoint[%s]", os.Args[1], os.Args[2])

	// Main Loop
	for {
		var con net.Conn

		// TCP Accept()
		con, err := l.Accept()
		if err != nil {
			Log.Error("Fail to Accept : %v", err)
			continue;
		}

		// Connect socket 얻기
		socket, _ := con.(*net.TCPConn)
		Log.Infor("Connect from : %v", con.RemoteAddr())

		// Run Go Routine(socket close 포함)
		go ClientHandler(socket)
	}

}
