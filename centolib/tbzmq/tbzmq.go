//
// Zmq wrapper function
package tbzmq

import (
	"github.com/pebbe/zmq4"
	"time"
	"strconv"
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
/// Zmq monitor socket을 만든다
/// @fn      func MakeMonitorSocket(base string,
///            (*zmq4.Socket, error){
/// @param   base  : PAIR socket을 만들 이름 BASE(없으면 자동생성)
/// @param   sock  : Monitor할 Zmq socket
/// @return   monitor socket, error
////////////////////////////////////////////////////////////////////////////////
func MakeMonitorSocket(base string, sock *zmq4.Socket) (*zmq4.Socket, error){
	var ipcname string
	var monsock *zmq4.Socket
	// PAIR socket 이름 Base가 없다
	if base == "" {
		ipcname = "inproc://" + strconv.FormatInt(time.Now().UnixNano(), 10) + ".mon"
	}else{
		ipcname = "inproc://"+base+strconv.FormatInt(time.Now().UnixNano(), 10) + ".mon"
	}
	
	// Monitor 대상 socket을 "inporc"에 연결(EVENT_ALL)
	err := sock.Monitor(ipcname, zmq4.EVENT_ACCEPTED | zmq4.EVENT_DISCONNECTED)
	if err != nil {
		return nil, fmt.Errorf("sock.Monitor) %v",err)
	}
	// Context 얻기
	ctx,_ := sock.Context()
	// Monitor socket을 만든다.(PAIR type 만 가능)
	monsock, err = ctx.NewSocket(zmq4.PAIR)
	if err != nil {
		return nil, fmt.Errorf("zmq4.PAIR : %v ", err)
	}
	// Monitor socket을 "inproc"에 연결
	err = monsock.Connect(ipcname)
	if err != nil {
		return nil, fmt.Errorf("monsock.Connect(%s) %v", ipcname, err)
	}

	return monsock, nil
}
////////////////////////////////////////////////////////////////////////////////
/// PUB SUB Proxy go routine
/// @fn      func RunPubSubProxy(ctx *zmq4.Context, f, b string)error
/// @param   f : Front End point (PUB point) 
/// @param   b : Back End point  (SUB point) 
/// @return 
////////////////////////////////////////////////////////////////////////////////
func RunPubSubProxy(ctx *zmq4.Context, f, b string)error{
	front, err := ctx.NewSocket(zmq4.XSUB)
	if err != nil{return fmt.Errorf("NewSocket(XSUB) : %v", err);}
	back, err := ctx.NewSocket(zmq4.XPUB)
	if err != nil{front.Close();return fmt.Errorf("NewSocket(XPUB) : %v",err);}
	err = front.Bind(f)
	if err != nil{
		front.Close()
		back.Close()
		return fmt.Errorf("Bind(%s) : %v", f,err)
	}
	err = back.Bind(b)
	if err != nil{
		front.Close()
		back.Close()
		return fmt.Errorf("Bind(%s) : %v", b,err)
	}
	time.Sleep(100 * time.Microsecond)

	go func(){
		zmq4.Proxy(front, back, nil)
	}()
	return nil
}
