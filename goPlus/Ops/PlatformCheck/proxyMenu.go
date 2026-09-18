package main
import (
	"fmt"
	sc "centolib/screen"
	"centolib/api"
	zmq "github.com/pebbe/zmq4"
	"encoding/json"
	"strings"
)
const (
	PROXY_WIDTH = 20
)

func displayProxyList(list []string){
	sc.SAVE_CURSOR()
	clearArea(1, AppTop, PROXY_WIDTH, AppBottom - AppTop)
	i := 0
	for _, name := range(list){
		fmt.Printf("[%d;1H%s", AppTop + i, name)
		i++
	}
	for y:= AppTop ; y <= AppBottom; y++{
		fmt.Printf("[%d;%dH│", y, PROXY_WIDTH + 1)
	}
	fmt.Printf("[%d;%dH┬", TitlePos + 1, PROXY_WIDTH + 1) 
	sc.RESTORE_CURSOR()
}
func getProxyInfor(proxyList *map[string]api.ProxyYaml)([]string, error){
	proxyYaml, err := api.GetPlatformEnv("PROXYENV_YAML")
	if err != nil{ return nil, err; }
	list, err := api.GetProxyList(proxyYaml)
	if err != nil{ return nil, err;}

	for i := 0; i < len(list); i++{
		proxy , err := api.ReadYaml(proxyYaml, list[i])
		if err != nil{return nil, err;}
		(*proxyList)[list[i]] = *proxy
	}
	return list, nil
}
func displayProxyInfor(name string, infor api.ProxyYaml)error{
	clearArea(PROXY_WIDTH + 3, AppTop, 
		MaxX- PROXY_WIDTH + 3, AppBottom - AppTop)
	i := 0	
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i,
				"%s", "【🔴 YAML 내용 】");i++
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"Proxy Name       : %s", name); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"ID               : %d", infor.ID); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"Comment          : %s", infor.Comment); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"Proxy Type       : %s", infor.ProxyType); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"FrontEnd Point   : %s", infor.FrontEndPoint); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"BackEnd Point    : %s", infor.BackEndPoint); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"ControlEnd Point : %s", infor.ControlEndPoint); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"CaptureEnd Point : %s", infor.CaptureEndPoint); i++;
	DrawHLine(true, PROXY_WIDTH +1, AppTop + i);i++;

	if infor.ControlEndPoint == ""{
		return fmt.Errorf("%s는 control end point가 없습니다.", name)
	}
	socket, err := zmq.NewSocket(zmq.REQ)
	if err != nil { return err;}
	socket.Connect(infor.ControlEndPoint)
	defer socket.Close()

	socket.SendMessage("INFOR")
	msg, _ := socket.RecvMessage(0)

	var pInfor api.ProxyInfor
	
	cleanMsg := strings.TrimSpace(msg[0])
	err = json.Unmarshal([]byte(cleanMsg), &pInfor)
	if err != nil{ return err; }
	
	//sc.COLOR_MAGENTA()
	sc.COLOR_SCYAN()
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i,
					"%s", "【🔴 Process 정보】");i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"ID               : %d", pInfor.ID); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"Comment          : %s", pInfor.Comment); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"RunHost          : %s", pInfor.RunHost); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"Proxy Type       : %s", pInfor.ProxyType); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"FrontEnd Point   : %s", pInfor.FrontEndPoint); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"BackEnd Point    : %s", pInfor.BackEndPoint); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"ControlEnd Point : %s", pInfor.ControlEndPoint); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"CaptureEnd Point : %s", pInfor.CaptureEndPoint); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"StartTime        : %s", pInfor.StartTime); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"ReadCount        : %d", pInfor.ReadCount); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"WriteCount       : %d", pInfor.WriteCount); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"FrontCount       : %d", pInfor.FrontCount); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"BackCount        : %d", pInfor.BackCount); i++;
	outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"CaptureCount     : %d", pInfor.CaptureCount); i++;
	if pInfor.Pause == true{
		sc.COLOR_RED()
		outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"Pasue            : 정지 상태(PAUSE)"); i++;
	}else{
		outStrXy(true, PROXY_WIDTH + 3, AppTop + i, 
				"Pasue            : 수행중(RUN)"); i++;
	}

	sc.COLOR_RESET()
	return nil
}
// ProxyMenu는 Proxy 정보 메뉴를 표시하고 키 입력을 처리한다.
func ProxyMenu() bool{
	proxyList := make(map[string]api.ProxyYaml)

	list, err := getProxyInfor(&proxyList)
	if err != nil{
		fmt.Printf("%v\n", err)
		return false
	}
	clearArea(1, AppTop, MaxX - 1, AppBottom - AppTop)

	displayProxyList(list)

	var (
		run bool
		pos int
	)
	run =true
	pos = 0
	MenuSelect(pos, list[pos])
	for run {
		char , _:=sc.GetKey()
		switch {
			case char == "q" || char == "Q" :
				return true
			case char == sc.DNARR || char == sc.RTARR :
				MenuUnSelect(pos, list[pos])
				pos ++
				if pos >= len(list) { pos = 0;}
				MenuSelect(pos, list[pos])
			case char == sc.UPARR || char == sc.LTARR :
				MenuUnSelect(pos, list[pos])
				pos --
				if pos < 0 { pos = len(list) - 1;}
				MenuSelect(pos, list[pos])
			case char == sc.ENTER :
				err = displayProxyInfor(list[pos], proxyList[list[pos]])

		}
		if err != nil{
			messageLine("ERROR : %v ", err)
		}else{
			messageLine("POS : %d-%s ", pos, list[pos])
		}
	}
	return true
}
