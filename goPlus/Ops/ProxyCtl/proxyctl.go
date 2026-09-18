// PROXY Control Port로 명령어 전달 (REQ)
// 
//
package main
import (
	"fmt"
	//"bytes"
	"encoding/json"
	zmq "github.com/pebbe/zmq4"
	"centolib/api"
	"strings"
	"os"
)

// Using은 프로세스 사용법을 표준 출력에 출력한다.
func Using(){
	fmt.Println("PROXY에REQ(Conrol-Mesage) Program")
	fmt.Printf("Using) goProxyCtl <Proxyname> <cmd>\n")
	fmt.Printf("\t<Proxyname> : ${PROXY_YAML}에 정의된 PROXY Server 이름\n")
	fmt.Printf("\t<cmd>       : 전송할 명령(INFOR,PAUSE,RUN)\n")
}
func main() {
	if len(os.Args) !=3 {
		Using()
		os.Exit(0)
	}
	yamlfile := os.Getenv("PROXY_YAML")
	if yamlfile == "" {
		fmt.Printf("PROXY_YAML 환경변수가 없습니다\n")
		os.Exit(0)
	}
	endpoint, err := api.GetProxyInfor(yamlfile, os.Args[1], "control")
	if err != nil{
		fmt.Printf("Proxy 정보 오류(%s:%s)\n%v\n", yamlfile, os.Args[1], err)
		os.Exit(0)
	}
	fmt.Printf("Proxy(%s) control point(%s)\n", os.Args[1], endpoint)
	reqsocket, _ := zmq.NewSocket(zmq.REQ)
	reqsocket.Connect(endpoint)

	reqsocket.SendMessage(os.Args[2])

	msg, _ := reqsocket.RecvMessage(0)
	cleanMsg := strings.TrimSpace(msg[0])
	fmt.Println("CMD : ", os.Args[2])

	var infor api.ProxyInfor
	err = json.Unmarshal([]byte(cleanMsg), &infor)
	if err != nil{
		fmt.Printf("Err : Unmarshal() :$v\n", err)
		return
	}
	jsonStr, err := json.MarshalIndent(infor, "", "   ")
	if err != nil{
		fmt.Printf("Err : MaeshalIndent() :$v\n", err)
		return
	}

	fmt.Println(string(jsonStr))
/******************
	var mapStr map[string]interface{}
	err = json.Unmarshal(jsonStr, &mapStr)
	if err != nil {
		fmt.Printf("Err : Unmarshal() :%v\n", err)
		return
	}
	fmt.Println()
	for key, value := range mapStr{
		fmt.Printf("KEY : %s Value :%v\n", key, value)
	}
******************/
}
