//
// Client Program
//
package main
import (
	"fmt"
	"net"
	"os"
)
func TcpClient(server string){

	rbuff := make([]byte, 2048)
	for i:= 0; i < 100; i ++{
    	// Connect to the server
    	conn, err := net.Dial("tcp", server)
    	if err != nil {
        	fmt.Println("Error connecting to server:", err)
			break
    	}

		msg := fmt.Sprintf("[%d] Process [%d] ", i, os.Getpid())
		_, err = conn.Write([]byte(msg))
		if err != nil{
			fmt.Printf("Error Write() : %v\n", err)
			break
		}
		//fmt.Printf("S[%d][%s]->  ", ln,msg)

		_, err = conn.Read(rbuff)
		if err != nil {
			fmt.Printf("Error Read() : %v\n", err)
			break
		}
		//fmt.Printf("R[%d][%s]\n", ln, string(rbuff[:ln]))
		
		conn.Close()
	}

}

func main(){

	if len(os.Args) != 2{
		fmt.Printf("TCP Client Program\n")
		fmt.Printf("Stdin -> TCP Send -> TCP Recv\n")
		fmt.Printf("Using) %s <connet port>\n", os.Args[0])
		os.Exit(0)
    }
    // Server address and port
    server := "127.0.0.1:"+os.Args[1]

	for i:=0; i < 100; i++{
		go TcpClient(server)
	}
	select {}
}
