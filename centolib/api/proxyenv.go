///
/// Proxy 관련 환경변수를 yaml file에서 읽는 Function
/// package name : centolib/api
/// file    proxyenv.go
/// date    2024. 12. 04. (수) 09:38:24 KST
/// author  Cento
///
package api
import (
	"time"
	"os"
	"gopkg.in/yaml.v3"
	"fmt"
)
///
/// Proxy Information
/// Capture endpoint에 대한 생각
///  PUS/SUB Proxy의 Capture는 Backend point에서 수신하는것으로 .....
///
type ProxyInfor struct{
	ID               int       // Proxy ID
	RunHost          string	   // Running host
	Comment          string    // Comment
	ProxyType        string    // REP/REQ or  PUB/SUB ...
	FrontEndPoint    string    // Front Endpoint
    BackEndPoint     string    // Back Endpoint
    ControlEndPoint  string    // Control Endpoint
    CaptureEndPoint  string    // Capture Endpoint - REP/REQ인 경우만
    StartTime        string    // Start Time
    ReadCount        int       // Read Packet count of frontendpoint
    WriteCount       int       // Write Packet count of backendpoint
	FrontCount       int       // Frontend Connection count
	BackCount        int       // Backend  Connection count
	CaptureCount     int       // Capture Connection count - REP REQ만
    Pause            bool      // PAUSE 상태
}
type ProxyYaml struct {
	ID              int    `yaml:"id"`
    Comment         string `yaml:"comment"`
    ProxyType       string `yaml:"proxytype"`
    FrontEndPoint   string `yaml:"front"`
    BackEndPoint    string `yaml:"back"`
    ControlEndPoint string `yaml:"control"`
    CaptureEndPoint string `yaml:"capture"`
}

func ReadYaml(yamlfile, name string)( *ProxyYaml, error){
	data, err := os.ReadFile(yamlfile)
	if err != nil{ return nil, err; }
	var objs map[string]ProxyYaml
	err = yaml.Unmarshal(data, &objs)
	if err != nil{ return nil, err;}
	if obj, exist := objs[name]; exist{
		return &obj, nil
	}else{
		return nil, fmt.Errorf("%s 는 없습니다", name)
	}
}

func NewProxyInfor(file, name string)(*ProxyInfor, error){
	yinfor, err := ReadYaml(file, name)
	if err != nil{ return nil, err; }

	ip, _:= GetMyIpv4Address()
	now := fmt.Sprintf("%s", time.Now().Format("2006 01 02 15:04:05.000"))
	infor := ProxyInfor{
		ID             : yinfor.ID,              // Proxy ID
    	RunHost        : ip ,                    // Running host
    	Comment        : yinfor.Comment,         // Comment
    	ProxyType      : yinfor.ProxyType,       // REP/REQ or  PUB/SUB ...
    	FrontEndPoint  : yinfor.FrontEndPoint,   // Front Endpoint
    	BackEndPoint   : yinfor.BackEndPoint,    // Back Endpoint
    	ControlEndPoint: yinfor.ControlEndPoint, // Control Endpoint
    	CaptureEndPoint: yinfor.CaptureEndPoint, // Capture Endpoint
    	StartTime      : now ,  // Start Time
    	ReadCount      :0 ,     // Read Packet count of frontendpoint
    	WriteCount     :0 ,     // Write Packet count of backendpoint
    	FrontCount     :0 ,     // Frontend Connection count
    	BackCount      :0  ,    // Backend  Connection count
    	Pause          :false , // PAUSE 상태
	}

	return &infor, nil
}

// Proxy의 정보를 yaml에서 얻기
// file : Proxy yaml file 이름 
// name : Proxy name
// key  : Information KEY string
func GetProxyInfor(file, name, key string)(string, error){
	pinfor, err := ReadYaml(file, name)
	if err != nil { return "", err;}
	switch key {
		case "id" : return fmt.Sprintf("%d", pinfor.ID), nil;
		case "comment": return pinfor.Comment, nil;
		case "proxytype": return pinfor.ProxyType, nil;
		case "front" : return pinfor.FrontEndPoint, nil;
		case "back" : return pinfor.BackEndPoint, nil;
		case "control" : return pinfor.ControlEndPoint, nil;
		case "capture" : return pinfor.CaptureEndPoint, nil;
	}
	return "", fmt.Errorf("%s는 없습니다", key);
}

func GetProxyList(yamlFile string)([]string, error){
	data, err := os.ReadFile(yamlFile)
	if err != nil{ return nil, err; }
	var node yaml.Node	
	err = yaml.Unmarshal(data, &node)
	if err != nil{ return nil, err; }
	var list []string
	root := node.Content[0]
	for i :=0; i < len(root.Content) ;i += 2{
		list = append(list,root.Content[i].Value)
	}
	return list, nil
}
