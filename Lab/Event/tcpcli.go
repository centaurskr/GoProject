package main
import (
	"centolib/api"
	"syscall"
	"fmt"
	"time"
	"os"
)
func main(){
	if len(os.Args) < 2{
		fmt.Println("TCP Client Program\n")
		fmt.Println("Using) tcpcli <endpoint\n")
		fmt.Println("\tex) tcpcli 127.0.0.1:32768\n")
		os.Exit(0)
	}
	fd, err:= api.OpenTcpClient(os.Args[1])
	if err != nil{
		fmt.Println("OpenTcpClient() : ", err)
		os.Exit(0)
	}
	defer syscall.Close(fd)
	fmt.Printf("Open Ok---- fd [%d]\n", fd)
	for i:= 0; i < 10; i++{
		packet := fmt.Sprintf("CLIENT DATA %v - %d", time.Now(), i)	
		wsz, err :=api.WriteDataStr4(fd, packet)
		if err != nil {
			fmt.Println("Write Error : ", err)
			os.Exit(0)
		}
		fmt.Printf("Write[%d] Size [%d]\n", i, wsz)
/*
		rsz,rpacket, err :=api.ReadDataStr4(fd)
		if err !=nil || rsz <= 0{
			fmt.Println("Read Error ", rsz, err)
			break
		}
		fmt.Printf("Read Data [%d] : %s\n", rsz, string(rpacket))
*/
		time.Sleep(1 * time.Second)
	}
}
