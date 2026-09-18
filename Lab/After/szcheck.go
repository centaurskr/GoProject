package main
import (
	"fmt"
	Mb "centolib/matchingboard"
	"centolib/api"
)
type Packet struct{
	tr string
	msg []byte
}
func main(){
var che      Mb.InnerChePacket
var ordJson  Mb.OrderJson
var packet   Packet

	fmt.Printf("size InnerChePacket : %d\n", api.SizeOf(che))
	fmt.Printf("size OrderJson      : %d\n", api.SizeOf(ordJson))

	packet.tr = "DBSEC"
	packet.msg = api.GetBytes(&che)
	fmt.Printf("sizeof packet (che) : %d\n", api.SizeOf(packet))
	fmt.Printf("%+v\n", packet)
	packet.tr = "DBCHE"
	packet.msg = api.GetBytes(&ordJson)
	fmt.Printf("sizeof packet (ordJson) : %d\n", api.SizeOf(packet))
	fmt.Printf("%+v\n", packet)
}
