///
/// @brief  GO TCP Server
/// @file   : server.go
/// @date   : 2024. 09. 11. (수) 11:34:01 KST
/// @author : Cento
///

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"runtime"
)
func readData2(c net.Conn) (string, error) {
	sz := make([]byte, 2)
	_,err :=	c.Read(sz)
	if err != nil {
		return "error", err
	}
	size := int(sz[0]) * 256 + int (sz[1])
	sz = nil // Hey GC! Free memory.

	data := make([]byte, size)
	_, err =	c.Read(data)
	if err != nil {
		return "error", err
	}
	str := string(data)
	data = nil // Hey GC!

	return str, nil
}
////////////////////////////////////////////////////////////////////////////////
///
/// @fn     : func ClientHandler(socket net.Conn){
/// @param  :
/// @return :
////////////////////////////////////////////////////////////////////////////////
func ClientHandler(c *net.TCPConn){
	for {
		str, err := readData2(c)
		if err != nil {
			log.Println("ERROR", err)
			break
		}
		fmt.Println(str)
	}
	fmt.Println("C Type is ", reflect.TypeOf(c))
	c.Close()
}
////////////////////////////////////////////////////////////////////////////////
///
/// @fn     : func main(){
/// @param  :
/// @return :
////////////////////////////////////////////////////////////////////////////////
func main(){
	if len(os.Args) != 2 {
		log.Println("GO TCP server")
		log.Println("USING : ", os.Args[0], "<port>")
		os.Exit(0)
	}
	runtime.GOMAXPROCS(4) // runtime pkg. max 4개의 CPU 사용
	fmt.Println("OS :", runtime.GOOS, ",ARCH :", runtime.GOARCH, 
			",CPUS :", runtime.NumCPU())
	port := ":" + os.Args[1]
	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Println(err)
		log.Println(" or Check <port>, must be number without colon!")
		os.Exit(0)
	}
	defer l.Close()

	for {
		var con net.Conn

		con, err := l.Accept()
		if err != nil {
			log.Println("Fail to Accept : ", err)
			continue;
		}
		socket, _ := con.(*net.TCPConn)
		log.Println("Connect from : ", con.RemoteAddr(),
			"\nsocket Type is : ", reflect.TypeOf(socket))
		go ClientHandler(socket)
	}

}
