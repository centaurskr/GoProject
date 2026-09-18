package main
import (
	"fmt"
	"net"
	zmq "github.com/pebbe/zmq4"
)

func CheckType(val interface{}){
	switch v:=val.(type){
	case net.Listener :
		fmt.Printf("Listener Type : %T(%v)\n",  val, v)
	case zmq.Socket :
		fmt.Printf("zmq.Socket Type : %T(%v)\n",  val, v)
	default :
		fmt.Printf("Type : %T\n", val)
	}
}
func main(){
	listener, _ := net.Listen("tcp", ":32767")
	CheckType(listener)	

	pub, _:= zmq.NewSocket(zmq.PUB)
	CheckType(pub)	

}
