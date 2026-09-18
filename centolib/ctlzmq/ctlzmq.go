//
// CONTROL PROXY를 통한 Control Message Mechanism
//
//  "CONTROL" + FROM + TO + message (PUB) 🡆  ┌─────────────────┐
//                                           │ 🔴CONTOL PROXY  │
//                     (SUB)              🡄  └─────────────────┘
// 
package ctlzmq
import (
	"centolib/api"
	zmq "github.com/pebbe/zmq4"
	"time"
)
var (
	CtlPubSocket *zmq.Socket
	FromName      string
)
///
/// Make control 
///
func MakeControlSocket(ctx *zmq.Context, pubPoint, subPoint, name string)(*zmq.Socket, error){
	so, err := ctx.NewSocket(zmq.PUB)
	if err != nil {return nil, err;}
	so.Connect(pubPoint)

	CtlPubSocket = so
	FromName = name

	so, err = ctx.NewSocket(zmq.SUB)
	if err != nil {return nil, err;}
	so.Connect(subPoint)

	so.SetSubscribe("CONTROL")
	time.Sleep(100 * time.Millisecond)

	return so, nil
} 


/// 
/// Send Control Message
// 
func SendControlMessage(toName,cmd string){
	CtlPubSocket.SendMessage("CONTROL", FromName, toName, cmd)
}

///
/// Close Control Message
///
func CloseControlMessage(){
	CtlPubSocket.Close()
}

///
/// Init control Message
///
func InitControlMessage(ctx *zmq.Context, name string)(*zmq.Socket, error){
    // Control Message
    val, err := api.GetPlatformEnv("PROXYENV_YAML")
    if err != nil{
        return nil, err
    }
    tstr,err := api.GetProxyInfor(val, "EXTRAINFOR", "front")
    if err != nil{
        return nil,err
    }
    pubPoint := tstr

    tstr,err = api.GetProxyInfor(val, "EXTRAINFOR", "back")
    if err != nil{
        return nil, err
    }
    subPoint := tstr

    ctlSub, err := MakeControlSocket(ctx, pubPoint, subPoint, name)
    if err != nil{
        return nil, err
    }
    return ctlSub, nil
}

