//
//【주문 수신 Process】
// 𝟭. 주문 pakcet의 vaildation check
// 𝟮. 접수번호 생성
// 𝟯. 증거금 Check (증거금 증가)
// 𝟰. ORDER_CALL Insert
// 𝟱. 응답 전송
//
// ┌──────────┐  ①    ┌────────────────┐
// │ OrdBrg   │  ⇄    │ REQ/REP Proxy  │
// └──────────┘  ✓    │ (goReqRepProxy)│       ┌────────────────┐ 
//          접수번호  └────────────────┘    ┌─ │  wallet using  │ 증거금Locking 
//                           ⑤ ⇅ ②          │  └────────────────┘  
//               ⑥    ┌────────────────┐ ③  │         o
//   ORDER_CALL   🡄   │ 🔴RecvOrder    │ ⇄  ┤         o
//  (주문 Insert)     └────────────────┘ ④  │         o
//                           ⑦ ↓            │  ┌────────────────┐
//                    ┌────────────────┐    └─ │  wallet using  │
//                    │  PUB/SUB Proxy │       └────────────────┘
//                    │ (goPubSubProxy)│
//                    └────────────────┘
//
// 【접수번호 체계】
//  Type : uint64 (unsigned long)
//  yymmdd+00+00000000000(11자리+100억)
//
 
package main

import (
	"fmt"
	"strconv"
	Log "centolib/netlog"
	Mb "centolib/matchingboard"
	zmq "github.com/pebbe/zmq4"
	ctl "centolib/ctlzmq"
	"centolib/api"
	"encoding/json"
	"database/sql"
	"strings"
	"unsafe"
	"os"
	"time"
)
var (
	Today, Index int             // yyyymmdd
	RegMap       []byte
	RegNumInfor  *Mb.RegNumber
	UsdExrate     Mb.EXRATE
	PubSocket    *zmq.Socket	
	DbContext    *sql.DB
	OrdCondMap   []byte	
	OrdCondHd    *Mb.MapHeader
	OrdCond      uintptr
	OrdCondSearch *api.SortSlice
) 

// Using은 프로세스 사용법을 표준 출력에 출력한다.
func Using(){
	fmt.Printf("주문 수신 Process\n")
	fmt.Printf("USING) %s <index> \n", os.Args[0])
	fmt.Printf("\tindex : 0 ~ 99\n")
}
//////// Order condition Attach //////////
func getOrderCondition()error{
	if len(OrdCondMap) == 0{
		api.UnMapfile(OrdCondMap)
		Log.Infor("기존 OrdCondition UnMap!")
	}
	mapName, err := api.GetPlatformEnv("ORDERCOND_MAP")
	if err != nil{return err;}
	OrdCondMap, _, err := api.AttachMapfile(mapName)
	if err != nil{return err;}
	OrdCondHd = (*Mb.MapHeader)(unsafe.Pointer(&OrdCondMap[0]))
	OrdCond = uintptr(unsafe.Pointer(&OrdCondMap[unsafe.Sizeof(Mb.MapHeader{})]))
	// Search Slice 
	// goMatchingCtl에 의하여 오름차순으로 정렬되어 있는 상태 
	// (by Symbol)
	OrdCondSearch = &api.SortSlice{
		StBase  : OrdCond,
		StCount : OrdCondHd.Count,
		StSize  : uintptr(unsafe.Sizeof(Mb.OrderCondition{})),
		LessFunc : func (a, b unsafe.Pointer) bool{
		    o1 := (*Mb.OrderCondition)(a)
            o2 := (*Mb.OrderCondition)(b)
            rtn := strings.Compare(string(o1.Symbol[:]), string(o2.Symbol[:]))
            return  rtn < 0  // 오름차순
		},
	}
	Log.Infor("주문상태(OrderCondition) 접속 OK!")
	return nil
}
////// 주문 상태 Check 1:정상, 0:주문 중지, 3:매도만 가능 //////
func checkOrderCondition(symbol string)int{
	var code [32]byte	
	copy(code[:], symbol)
	key := unsafe.Pointer(&Mb.OrderCondition{
		Symbol : code,
		})
	sp := OrdCondSearch.Search(key)
	if sp == nil{ return -1;}
	cond := (*Mb.OrderCondition)(sp)
	return cond.Status
}
//////// 접수번호 정보 Attach /////////////
func getRegNumberInfor()error{
	if len(RegMap) > 0{
		api.UnMapfile(RegMap)
	}
	mapfile, err := api.GetPlatformEnv("REGNUM_MAP")
	if err != nil{ return err; }
	RegMap,_,err := api.AttachMapfile(mapfile)
	if err != nil{
		return fmt.Errorf("AttachMapfile(%s) : %v", mapfile, err)
	}

	hd := (*Mb.MapHeader)(unsafe.Pointer(&RegMap[0]))

	Log.Infor(`
	접수번호 MAP
	생성일:%s\n\tBuffer[%d] Count[%d]`, 
	api.CleanString(hd.Create[:]), hd.Buffer, hd.Count)

	offset := api.SizeOf(Mb.MapHeader{}) + api.SizeOf(Mb.RegNumber{}) * Index
	RegNumInfor = (*Mb.RegNumber)(unsafe.Pointer(&RegMap[offset]))
	return nil
}

// 일자변경
func changeDate(t time.Time){
	//now := api.GetNumYyyymmdd()
	dateStr := t.Format("20060102")
	now, _ := strconv.Atoi(dateStr)
	if now != Today{
		numstr := fmt.Sprintf("%s%02d00000000000", api.GetStrYymmdd(), Index)
		seq, _ := strconv.ParseUint(numstr, 10, 64)
		Log.Infor(`
			일자변경 %d 🡆 %d
			접수번호 초기화 Index[%d] 접번호[%d] 🡆 [%d] 
			`,
			Today, now, Index, RegNumInfor.RegNum, seq)
		RegNumInfor.RegNum = seq
		Today = now
	}
}
// 일변경 Check go routine
func goCheckDate(){
	Log.Infor("Start CheckDate go routine. ToDay(%d]", Today)
	ticker := time.NewTicker(5 * time.Second)
	for{
		select{
		case t := <-ticker.C :
			changeDate(t)
		}
	}
}
// ExrateHandler는 EXTRAINFOR Proxy에서 환율 정보를 수신하여 UsdExrate를 갱신한다.
func ExrateHandler(ent *api.GoEvent, socket *zmq.Socket){
	exrate := Mb.EXRATE {}
	exmsg, _:= socket.RecvMessageBytes(0)
	err := json.Unmarshal(exmsg[1], &exrate)	
	if err != nil{
		Log.Error("환율 수신 Unmarshal()오류.%v:%s", string(exmsg[1][:]))
		return
	}
	//if exrate.Currency == "USDT"{
	if exrate.Currency == "USD"{
		UsdExrate = exrate
	}
	Log.Infor("환율수신 : %+v", UsdExrate)
	return
}
// ControlHandler는 EXTRAINFOR Proxy에서 제어 메시지를 수신하여 처리한다.
func ControlHandler(ent *api.GoEvent, socket *zmq.Socket){
	msg, err := socket.RecvMessage(0)
	if err != nil{
		Log.Error("RecvMessage for control Message : %v", err)
		return
	}
	if len(msg) < 4 || msg[0] != "CONTROL"{
		Log.Warn("Control Message 오류 : %v", msg)
		return
	}
	// msg[0] : topic "CONTROL"
	// msg[1] : From 
	// msg[2] : to
	// msg[3] : Command
	if msg[1] == "goMatchingCtl" && msg[2] == "goRecvOrder" && msg[3] == "REATTACH"{
		err := getOrderCondition()
		if err != nil{
			Log.Error("주문상태(OrderCondition))을 REATTACH할 수 없습니다")
			Log.Error("조치가 필요합니다")
		}
		Log.Error("주문상태(OrderCondition))을 REATTACH 성공")
	}
	//Log.Infor("CONTROL : %v", msg)
	return
}
// ReqHandler는 ORDERPASS Proxy에서 주문 REQ 메시지를 수신하여 처리하고 응답을 반환한다.
// 처리 순서: JSON 파싱 → 접수번호 생성 → 주문 정합성 검증 → 증거금 확인 → DB 저장 → 응답 전송 → PUB
func ReqHandler(ent *api.GoEvent, socket *zmq.Socket){
var rtnJson     Mb.OrderJson    // for REP

	// 주문 Data처리, ⛔ 처리 순서 주의
	msg, _ := socket.RecvBytes(0)
	Log.Infor("수신 : %s\n", string(msg[:]))	

	// Json에 문제가 없는가?
	err := json.Unmarshal(msg, &rtnJson) // []byte -> OrderJson
	if err != nil{
		rtnJson.State = "COM-IF-0"
		rtnJson.Message = "주문 Format(json)오류입니다"
		reqmsg, _ := json.Marshal(rtnJson) // OrderJson -> []byte
		reqmsg = append(reqmsg, msg...)
		socket.SendMessage(reqmsg)
		Log.Warn("%+v", rtnJson)
		return
	}
	// 1. 접수번호 && 접수시각
	RegNumInfor.RegNum ++

	// 2. 주문 정합성 Check + 주문 Packet Covert(환율, 증거금)
	inner, err := ValidateOrder(&rtnJson)

	// 접수번호
	rtnJson.RegTime = fmt.Sprintf("%d", time.Now().UnixMicro())
	rtnJson.RegNum = fmt.Sprintf("%d", RegNumInfor.RegNum)

	if err != nil{ // ValidateOrder()
		Log.Error("VaidateOrder():%v\n%s", err, string(msg[:]))	
		err := DbInsertOrder(DbContext, &rtnJson)	
		if err != nil{
			Log.Error("DbInsertOrder()\n%v\n%+v", err, rtnJson)	
		}
		// 응답전송
		reqmsg, _ := json.Marshal(rtnJson) // OrderPacket -> JSON
		socket.SendMessage(reqmsg)
		return
	}

	// 3. 증거금 Check
	ok, err := DbCheckMargin(DbContext, &rtnJson) // 2에서 증거금 계산 함
	if err != nil || ok == false{
		if err != nil{
			Log.Error("DbCheckMargin [%+v] : %v", rtnJson, err)
			rtnJson.State = "SYS-DB-1"
			rtnJson.Message = "처리 오류입니다"
		}else if ok == false{
			Log.Warn("증거금 부족 : %+v", rtnJson)
			rtnJson.State = "ORD-IF-1"
			rtnJson.Message = "증거금이 부족합니다"
		}
		err = DbInsertOrder(DbContext, &rtnJson)
		if err != nil{
			Log.Error("DbInsertOrder()\n%v\n%+v", err, rtnJson)	
		}
		// 응답전송
		reqmsg, _ := json.Marshal(rtnJson) // OrderPacket -> JSON
		socket.SendMessage(reqmsg)
		return
	}

	// 4. InsertOrder(order)
	err = DbInsertOrder(DbContext, &rtnJson)
	if err != nil{
		Log.Error("DbInsertOrder()\n%v\n%+v", err, rtnJson)	
		rtnJson.State = "SYS-DB-2"
		rtnJson.Message = "처리 오류입니다(DB-INS)"
		//응답전송
		reqmsg, _ := json.Marshal(rtnJson) // OrderPacket -> JSON
		socket.SendMessage(reqmsg)

		// 증거금 rollback해야함-- TODO

		//ent.GoEventDown()
		return
	}

	// 5. 응답 전송
	reqmsg, _ := json.Marshal(rtnJson) // OrderPacket -> JSON
	socket.SendMessage(reqmsg)

	// 주문번호와 접수번호를 같게 한다(Matching 이전)
	inner.MemData.RegNum = RegNumInfor.RegNum
	inner.MemData.OrdNum = RegNumInfor.RegNum
	copy(inner.MemData.OrdPacket[:], reqmsg) //응답 Packet을 Json으로

	// 5. 내부 Packet PUB
	pubmsg := api.GetBytes(&inner)
	PubSocket.SendMessage(rtnJson.Symbol, pubmsg)

	Log.Infor("응답 : %+v", rtnJson)
	return
}
///////////////////// Main Function ///////////////////
func main(){
	if len(os.Args) != 2{
		Using()
		os.Exit(0)
	}
	Index,_ = strconv.Atoi(os.Args[1])
	if Index < 0 || Index > 99{
		Using()	
		os.Exit(0)
	}
	err := Log.InitNetLogout(os.Args[0], os.Args[1]);	
	if err != nil {
		fmt.Printf("InitNetLogout() %v\n", err)
		os.Exit(0)
	}
    defer Log.Close()

	api.UsingMaxProcess()

	Log.Infor("주문 수신 Process -- INDEX[%s]", os.Args[1])

	// Create zmq Context
	context, err := zmq.NewContext()
	if err != nil{
		Log.Error("Zmq.NewCointext() : %v", err)
		os.Exit(0)
	}

	// 1/9. 중복실행 검사
	lockname := fmt.Sprintf("goRecvOrder-%02d",  Index)
	isrun, err := api.IsRunning(lockname)
	if isrun == true || err != nil{
		Log.Error("\n\t%s Index[%02d}는 실행중입니다.\n\t%v", 
			os.Args[0], Index, err)
		os.Exit(0)
	}
	Log.Infor("1/10 중복실행 검사 OK...")

	// 2/9.접수번호 Map 
	err = getRegNumberInfor()
	if err != nil{
		Log.Error("%v", err)
		os.Exit(0)
	}
	Log.Infor("2/10. Index : %d, 접수번호 : %d", Index, RegNumInfor.RegNum)

	// 3/10. 주문 상태 Attach
	err = getOrderCondition()
	if err != nil{
		Log.Error("%v", err)
		os.Exit(0)
	}
	Log.Infor("3/10 Get 주문정보 : Date[%s] Count[%d]", 
			api.CleanString(OrdCondHd.Create[:]), OrdCondHd.Count)

	// 4/10. GoEvent
	event := api.GoEvent{}
	event.InitGoEvent()
	Log.Infor("4/10 Event 생성")

	// 5/10. 통신 정보 설정
	yamlfile,err := api.GetPlatformEnv("PROXYENV_YAML")
	if err!=nil{
		Log.Error("%v", err)
		os.Exit(0)
	}
	reppoint, err := api.GetProxyInfor(yamlfile, "ORDERPASS", "back")
	if err != nil{
		Log.Error("%s의 ORDERPASS Proxy에 back이 없습니다", yamlfile)
		os.Exit(0)
	}
	repsock, err := context.NewSocket(zmq.REP)
	if err != nil{
		Log.Error("주문수신(REP) Socket 생성 : %v", err)
		os.Exit(0)
	}
	err = repsock.Connect(reppoint)
	if err != nil{
		Log.Error("주문수신(REP) Connect(%s) : %v", reppoint, err)
		os.Exit(0)
	}
	event.AddZmqClient(repsock, ReqHandler)
	Log.Infor("5.1/10 ORDERPASS Proxy back(REP) end point is : %s", reppoint)

	pubpoint, err := api.GetProxyInfor(yamlfile, "BEFOREMATCHING", "front")
	if err != nil{
		Log.Error("%s file의 BEFOREMATCHING Proxy에 front가 없습니다", yamlfile)
		os.Exit(0)
	}
	PubSocket, err = context.NewSocket(zmq.PUB)
	if err != nil{
		Log.Error("주문전달(PUB) Socket 생성  : %v", err)
		os.Exit(0)
	}
	err = PubSocket.Connect(pubpoint)
	if err != nil{
		Log.Error("주문전달(PUB) Connect(%s) : %v", pubpoint, err)
		os.Exit(0)
	}
	Log.Infor("5.2/10 BEFOREMATCHING Proxy fron(PUB) end point is : %s", pubpoint)
	time.Sleep(1000 * time.Millisecond) // for good zmq connect

	// 6/10. 환율 정보 Sub
	subpoint, err := api.GetProxyInfor(yamlfile, "EXTRAINFOR", "back")
	if err != nil{
		Log.Error("%s file의 EXTRAINFOR Proxy에 back이 없습니다", yamlfile)
		os.Exit(0)
	}
	exsubsocket, err := context.NewSocket(zmq.SUB)
	if err != nil{
		Log.Error("환율수신(SUB) socket 생성 : %v", err)
		os.Exit(0)
	}
	err = exsubsocket.Connect(subpoint)
	if err != nil{
		Log.Error("환율수신(SUB) Connect(%s) : %v", subpoint, err)
		os.Exit(0)
	}
	exsubsocket.SetSubscribe("EXRATE")
	event.AddZmqClient(exsubsocket, ExrateHandler)
	Log.Infor("6/10 환율 SUB point is : %s", subpoint)

	// 7/10. 오늘 일자 yyyymmdd
	Today = api.GetNumYyyymmdd()
	Log.Infor("7/10 오늘 : %d", Today)
	go goCheckDate()

	// 8/10. Db Connect
	dbyaml,err := api.GetPlatformEnv("DBCONFIG_YAML")
	if err != nil{
		Log.Error("%v", err)
		os.Exit(0)
	}
	DbContext, err = api.ConnectDb(dbyaml)
	if err  != nil{
		Log.Error("ConnectDb(%s):%v", dbyaml, err)
		os.Exit(0)
	}
	Log.Infor("8/10 Db Connect Ok")

	// 9/10.환율 읽기 (USD)
	basetime, exrate, err := DbSelectLastExrate(DbContext, "USD")
	if err != nil{
		Log.Error("환율 : %v", err)
		os.Exit(0)
	}
	UsdExrate.BaseTime = basetime
	UsdExrate.Currency = "USD"
	UsdExrate.Rate = fmt.Sprintf("%f", exrate)
	Log.Infor("9/10 환율 : %+v", UsdExrate)

	// 10/10. Control Message
	ctlSocket, err := ctl.InitControlMessage(context, "goRecvOrder")
	if err != nil{
		Log.Error("InitControlMessage : %v", err)
	}
	ctlSocket.SetSubscribe("goRecvOrder")
	ctlSocket.SetSubscribe("ALL")
	event.AddZmqClient(ctlSocket, ControlHandler)
	Log.Infor("10/10 Control Message 초기화, Name \"goRecvOrder\"")

	Log.Infor("START goRecvOrder -- Index [%d]", Index)

	event.GoEventLoop()
}
