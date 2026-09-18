// 매칭 이후 처리를 위한 function
// -. 25.01.22
//                       ┌────────────┐  "IPCCURR" 
//【gofCoinMatching()】  │ InnerProxy │   🡆  【gofSiseFunction】
//                       └────────────┘
//                          🡇  "IPCCHE"     
//              "IPCCURR" 🡅    "IPCSEQ" 
//                 ┌───────────────────┐    ┌─────────────┐ 
//                 │ 🔴goAfterFunction │ 🡆  │AFTERMATCHING│
//                 └───────────────────┘    └─────────────┘
//                                      
//  1. 체결 증거금 계산
//  2. Maker/Taker의 "DBCHE"(체결 packet) 
//  3. "IPCCURR" 내부전송(현재가: Maker Price)
//
package main
import (
	zmq "github.com/pebbe/zmq4"
	Log "centolib/logout"
	Bd "centolib/bindecimal"
    Mb "centolib/matchingboard"
	"centolib/api"
	//"unsafe"
	"os"
	"encoding/binary"
	"encoding/json"
	"bytes"
	"strconv"
	"unsafe"
	"fmt"
	"time"
)

///////////// InnerChePacket 만들기 ///////////
func makeInnerChe(ord *Mb.Order, 
	key    uint64, 
	chetm  uint64,
	qty    Bd.BinDecimal, 
	price  Bd.BinDecimal,
	cntnum uint64,
	maketake byte)Mb.InnerChePacket{

	pk := Mb.InnerChePacket{
    	CheKey     : key         ,// 체결 Pair Key   
    	UserId     : ord.UserId  ,// User Id
    	MakeTake   : maketake    ,// "M" Maker,  "T" Taker
    	OrdNum     : ord.OrdNum  ,// 주문번호
		OrdType    : ord.OrdType ,// 주문종류(L:지정가 ,M:시장가)
    	BuySell    : ord.BuySell ,// 매도수구분 (S:매도, B:매수)
    	CheSeq     : ord.CheSeq  ,// 체결번호
    	Price      : price        ,// 체결가격
    	Qty        : qty         ,// 체결수량
    	RemQty     : ord.RemQty  ,// 주문잔량
    	CheTime    : chetm       ,// 체결시각
    	CntOrdNum  : cntnum      ,// 상대 주문번호   

		OrdPacket  : ord.OrdPacket, // 주문 JSON
	}
	var jsonpk Mb.OrderJson
	err := json.Unmarshal(
		bytes.ReplaceAll(ord.OrdPacket[:], []byte{0x00}, nil),
		&jsonpk,
	)
	if err != nil {
		Log.Error(
	"json.Unmarshal :key(%d)chetm(%d)Qty(%s)Prc(%s)cntnum(%d) MakeTake(%c):%+v:%v",
 	key, chetm, qty.String(), price.String(), cntnum, maketake, pk, *ord)
		os.Exit(0)
	}
	copy(pk.Symbol[:], jsonpk.Symbol)
	copy(pk.UserField[:], jsonpk.UserField[:min(len(jsonpk.UserField), 60)])

	pk.ItemCode,_ = strconv.Atoi(jsonpk.ItemCode)
	pk.MkCode  ,_ = strconv.Atoi(jsonpk.MkCode)
	pk.CoinCode,_ = strconv.Atoi(jsonpk.CoinCode)
	pk.ExRate.Set(jsonpk.ExRate) // 환율

	if ord.BuySell == 'B'{// 매수이면 계산
		// 시장가 매수인 경우
		// 잔량이 0이고 거래불가능한 수량이 증거금에 남아있어야 하기 때문에 
		// 잔량이 0이고 지정가인 경우 만 해당함
		if Bd.DecimalIsZero(ord.RemQty) == true && ord.OrdType == 'L'{
			// 잔량이 없으면, 변제금을 나머지로
			pk.Margin = ord.Margin
			pk.Fee    = ord.Fee
		}else{
			// 주문증거금 (주문가격 X 체결 수량)-버림
			pk.Margin,_ = ord.Price.Mul(qty, Bd.ROUND_DN, 8)
			// 체결증거금(변제용)- 버림
    		pk.Fee ,_ = pk.Margin.Mul(Bd.DecimalMax(ord.MakeRate, ord.TakeRate),
						Bd.ROUND_DN, 8)              
		}
	}

	// 체결 금액(체결가격X체결수량)-없음
   	pk.CheAmt,_ = price.Mul(qty, Bd.NOROUND, 8)
	// 체결 수수료(매수/매도 동일)-올림
	if maketake == 'M'{ // Maker
   		pk.CheFee,_ = pk.CheAmt.Mul(ord.MakeRate, Bd.ROUND_UP, 8)
		pk.FeeRate = ord.MakeRate
	}else{ // Taker
   		pk.CheFee,_ = pk.CheAmt.Mul(ord.TakeRate, Bd.ROUND_UP, 8)
		pk.FeeRate = ord.TakeRate
	}
	return pk
}
// Matching 이후 Go Routine
// 1. IPCCURR 생성(현재가, 현재가는 maker 가격))
// 2. DBCHE Packet 생성 전달
// 
func gofAfterFunction(ctx *zmq.Context){
var (
	chetime, pairkey uint64
	cheqty   Bd.BinDecimal
	chepk    Mb.InnerChePacket
)

	// 수신
	ipcsub, err := ctx.NewSocket(zmq.SUB)
	if err != nil{
		Log.Error("NewSocket(zmq.SUB : %v", err)	
		os.Exit(0)
	}
	Log.Infor("Start After. Ipc[%s]", InnerBack)
	err = ipcsub.Connect(InnerBack)
	if err != nil{Log.Error("v", err);os.Exit(0);}
	ipcsub.SetSubscribe("IPCCHE")       // AFTERMATCHING, SISEJSON(maker)
	ipcsub.SetSubscribe("IPCSEQ")       // AFTERMATCHING
	ipcsub.SetSubscribe("IPCERR")       // AFTERMATCHING
	//ipcsub.SetSubscribe("IPCHOGA")      // SISEJSON
	ipcsub.SetSubscribe("IPCCANCEL")    // AFTERMATCHING
	ipcsub.SetSubscribe("IPCMKDONE")     // AFTERMATCHING

	// 외부전송
	val, err := api.GetProxyInfor(ProxyYaml, "AFTERMATCHING", "front")
	if err != nil{
		Log.Error("AFTERMATCHING:front : %v", err)
		os.Exit(0)
	}
	afterPub , err := ctx.NewSocket(zmq.PUB)
	if err != nil{
		Log.Error("NewSocket(zmq.PUB : %v", err)	
		os.Exit(0)
	}
	err = afterPub.Connect(val)
	if err != nil{Log.Error("v", err);os.Exit(0);}

	// 시세 전송(체결) 
	ipcPub , err := ctx.NewSocket(zmq.PUB)
	if err != nil{
		Log.Error("NewSocket(zmq.PUB : %v", err)	
		os.Exit(0)
	}
	err = ipcPub.Connect(InnerFront)
	if err != nil{Log.Error("v", err);os.Exit(0);}

	time.Sleep(100*time.Microsecond) 

	poller := zmq.NewPoller()
	poller.Add(ipcsub, zmq.POLLIN)

	for {
		polls, _ := poller.Poll(1000 * time.Millisecond)
		for _, polled := range polls{

			if polled.Socket != ipcsub { continue; }

			msg, err := ipcsub.RecvMessageBytes(0)
			//if err != nil{Log.Error("%v", err);continue}
			if err != nil{Log.Error("%v", err);os.Exit(0);}
			switch string(msg[0][:]) { // 내부 Topic
			case "IPCSEQ" :
				// msg[0] "IPCSEQ"  -- string
				// msg[1] OrderJson -- []byte
				var ordjson Mb.OrderJson
				err := json.Unmarshal(msg[1], &ordjson)
				if err != nil{
					Log.Error("IPCSEQ JSON 변경 오류 : %v", msg[1])
					continue
				}
				uid, err:= strconv.Atoi(ordjson.UserId)
				if err != nil{
					Log.Error("Uid변환 실패.UserId[%s]:%v", ordjson.UserId, err)
					continue
				}
				outTopic := fmt.Sprintf("AFTER%02d",  uid % Nvalue)
				afterPub.SendMessage(outTopic, "DBSEQ", msg[1])
	Log.Debug("IPCSEQ : OrdNum(%s) UID[%d] NValue(%d) TOPIC(%s) ", ordjson.OrdNum, uid, Nvalue, outTopic )

			case "IPCCHE" :
				// msg[0] "IPCCHE",  -- string
				// msg[1] pairkey,  -- []byte
				// msg[2] chetime,  -- []byte, UnixMicro
				// msg[3] cheqty,   -- string
				// msg[4] maker,    -- []byte Order 
				// msg[5] taker     -- []byte Order 
				// 증거금 계산을 위해
				//   maker와 taker는 모두  주문잔량이 계산된 상태
				//
				// 1. Maker DBCHE 정보 전송
				// 2. Taker DBCHE 정보 전송
				// 3. Maker 정보로 IPCCURR를 만들어 전송

				binary.Read(bytes.NewReader(msg[1]), 
							binary.LittleEndian, &pairkey)
				binary.Read(bytes.NewReader(msg[2]), 
							binary.LittleEndian, &chetime)
				cheqty.Set(string(msg[3][:]))

				maker := (* Mb.Order)(unsafe.Pointer(&msg[4][0]))
				taker := (* Mb.Order)(unsafe.Pointer(&msg[5][0]))

				// Maker
				makerTopic := fmt.Sprintf("AFTER%02d",  maker.UserId % Nvalue)	
				chepk = makeInnerChe(maker, 
								pairkey, chetime, cheqty, maker.Price,
								taker.OrdNum,
								'M')

				afterPub.SendMessage(makerTopic, "DBCHE", api.GetBytes(&chepk))
				
				// IPCCURR to InnerFront
				ipcPub.SendMessage("IPCCURR", api.GetBytes(&chepk))	

				// Taker
				takerTopic := fmt.Sprintf("AFTER%02d",  taker.UserId % Nvalue)	
				chepk = makeInnerChe(taker, 
								pairkey, chetime, cheqty, maker.Price,
								maker.OrdNum,
								'T')
				afterPub.SendMessage(takerTopic, "DBCHE", api.GetBytes(&chepk))
				
Log.Debug(`IPCCHE : 
MakerTopic[%s] TakerTopic[%s]
Tm[%d] CheQty[%s] M(P/Rem)[%s:%s]MNum(%d) T(P/Rem)[%s:%s] tNum(%d)`,
						makerTopic, takerTopic,
						chetime,
						cheqty.String(),
						maker.Price.String(),
						maker.RemQty.String(),
						maker.OrdNum,
						taker.Price.String(),
						taker.RemQty.String(),
						taker.OrdNum,)
			case "IPCERR" :
				// msg[0] -> IPCERR
				// msg[1] -> type ("IPCERR")
				// msg[2] -> OrderJson
				var ordjson Mb.OrderJson
				err := json.Unmarshal(msg[2], &ordjson)
				if err != nil{
					Log.Error("IPCSEQ JSON 변경 오류 : %v", msg[1])
					continue
				}
				uid, err:= strconv.Atoi(ordjson.UserId)
				if err != nil{
					Log.Error("Uid변환 실패.UserId[%s]:%v", ordjson.UserId, err)
					continue
				}
				outTopic := fmt.Sprintf("AFTER%02d",  uid % Nvalue)
				afterPub.SendMessage(outTopic, "DBERR", msg[2])
Log.Debug("DBERR Topic[%s]", outTopic)
			case "IPCCANCEL" : // 취소주문
				// msg[0] -> IPCCANCEL
				// msg[1] -> OrderJson
				var ordjson Mb.OrderJson
				err := json.Unmarshal(msg[1], &ordjson)
				if err != nil{
					Log.Error("IPCSEQ JSON 변경 오류 : %v", msg[1])
					continue
				}
				uid, err:= strconv.Atoi(ordjson.UserId)
				if err != nil{
					Log.Error("Uid변환 실패.UserId[%s]:%v", ordjson.UserId, err)
					continue
				}
				outTopic := fmt.Sprintf("AFTER%02d",  uid % Nvalue)
				afterPub.SendMessage(outTopic, "DBCANCEL", msg[1])
Log.Debug("DBCANCEL Topic[%s]", outTopic)
			case "IPCMKDONE" : 
				// msg[0] -- IPCMKDON
				// msg[1] -- MARKET
				// msg[2] -- OrderJson String
				var ordjson Mb.OrderJson
				err := json.Unmarshal(msg[2], &ordjson)
				if err != nil{
					Log.Error("IPCMKDON JSON 변경 오류 : %v", msg[2])
					continue
				}
				uid, err:= strconv.Atoi(ordjson.UserId)
				if err != nil{
					Log.Error("Uid변환 실패.UserId[%s]:%v", ordjson.UserId, err)
					continue
				}
				outTopic := fmt.Sprintf("AFTER%02d",  uid % Nvalue)
				afterPub.SendMessage(outTopic, "MKDONE", msg[2])
			default :
			} // Switch
		}// range polls
	} // Main Loop
}
