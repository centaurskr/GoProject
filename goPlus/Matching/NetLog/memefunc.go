// Matching Meme function
//                                   ┌───────────────┐
//   【PUB/SUB Proxy】               │  goMatching   │    ⇆   control Message
//            🡇                      └───────────────┘
//     ┌────────────┐  "CHANGEDATE" + yyyymmdd  🡇  
//     │DO┌─────────┴──┐      🡄       ┌─────────────┐
//     └──┤ ET┌────────┴──┐   🡆       │ Inner Proxy │
//        └───┤  BTC.USDT │           └─────────────┘
//            └───────────┘  "IPCFDS"   【gofFdsFunction  】🡆
//      (🔴gofCoinMatching)  "IPCSISE"  【gofSiseFunction 】🡆
//     생성 "IPCFDS"         "IPCHOGA" 
//          "IPCCHE"         "IPCCHE"   【gofAfterFunction】🡆
//          "IPCHOGA"        "IPCMKDONE"  생성 "IPCSISE"
//
package main
import (
	zmq "github.com/pebbe/zmq4"
	Log "centolib/netlog"
    Mb "centolib/matchingboard"
	Bd "centolib/bindecimal"
	Dc "github.com/shopspring/decimal"
	"centolib/api"
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"unsafe"
	"fmt"
	"os"
)
// Hoga Json string을 만든다
// return 
// -. FULL String : mode = 0
// -. 간략 String : mode = 1
func makeHogaString(memcoin *Mb.MEMCOIN, mode int)string{
	// Max : 15 호가
	// {"sy":"BTC.USDT", "t":"", "so":"xx", "sq":"xx","bo":"yy","bq":"",
	//  "S":[{"p":"","c":"","q":""},...],
	//  "B":[{"p":"","c":"","q":""},...]}
	var (
		hogastr string
		oc int
		oq Dc.Decimal
	)
	if mode == 0 { // Full Mode
		// 종목 Format(2006 01 02 15 04 05")
		hogastr = fmt.Sprintf("{\"sy\":\"%s\",\"t\":\"%s\",", 
			memcoin.GetSymbol(),time.Now().Format("20060102150405.000"))
		// 총 매도호가 수량 , 주문건수 (SELL)
		oc = 0
		oq = Dc.NewFromInt(0)
		for _, h:= range memcoin.Medo{
			oc = oc + h.OrderCount	// 주문 건수
			oq = oq.Add(h.SumQty)   // 매도호가 수량
		}
		tstr := fmt.Sprintf("\"so\":\"%d\",\"sq\":\"%s\",",oc, oq.String())	
		hogastr = hogastr + tstr
		// 총 매수호가 수량 , 주문건수 (BUY)
		oc = 0
		oq = Dc.NewFromInt(0)
		for _, h:= range memcoin.Mesu{
			oc = oc + h.OrderCount	// 주문 건수
			oq = oq.Add(h.SumQty)   // 매수호가 수량
		}
		tstr = fmt.Sprintf("\"bo\":\"%d\",\"bq\":\"%s\",",oc, oq.String())	
		hogastr = hogastr + tstr
	}
	// 매도
	hogastr = hogastr + "\"S\":["
	for i, h := range memcoin.Medo{
		if i >= 15{ break;}
		if Bd.DecimalEqual(h.Price, Mb.MAX_PRICE) == 1{break;}
		tstr := fmt.Sprintf("{\"p\":\"%s\",\"c\":\"%d\",\"q\":\"%s\"},",
			h.Price.String(), h.OrderCount, h.SumQty.String())
		hogastr = hogastr + tstr
		
	}
	// 마지막 "," 지우기
	idx := strings.LastIndex(hogastr, ",")
	if idx != -1{
		if idx == len(hogastr) -1 || strings.Contains("}]", 
										string(hogastr[idx+1])){
			hogastr = hogastr[:idx] + hogastr[idx + 1:]
		}
	}

	// 매수
	hogastr = hogastr + "], \"B\":["
	for i, h:= range memcoin.Mesu{
		if i >= 15{ break;}
		if Bd.DecimalEqual(h.Price, Mb.MIN_PRICE) == 1{break;}
		tstr := fmt.Sprintf("{\"p\":\"%s\",\"c\":\"%d\",\"q\":\"%s\"},",
			h.Price.String(), h.OrderCount, h.SumQty.String())
		hogastr = hogastr + tstr
	}

	hogastr = hogastr + "]"
	// 마지막 "," 지우기
	idx = strings.LastIndex(hogastr, ",")
	if idx != -1{
		if idx == len(hogastr) -1 || strings.Contains("}]", 
										string(hogastr[idx+1])){
			hogastr = hogastr[:idx] + hogastr[idx + 1:]
		}
	}
	if mode == 0{
		hogastr = hogastr + "}"
	}
	return hogastr
}
///// FDS 전송 /////
func sendFdsData(ipcpub *zmq.Socket, jsonpk Mb.OrderJson, memcoin *Mb.MEMCOIN)error{
	streamjson, err := json.Marshal(jsonpk) // struct -> json stream
	if err != nil{
		Log.Error("COIN[%s]- %v", memcoin.GetSymbol(), err)
		return err
	}
	
	streamhoga := makeHogaString(memcoin, 1) // 간소화 호가 String
	ipcpub.SendMessage("IPCFDS", 
			memcoin.MstInfor.BfPrice.String(),
			api.GetBytes(&memcoin.MstInfor.BfTime),
			streamjson, streamhoga)
	return nil
}

// InnerControlProcess는 내부 Control 소켓에서 메시지(일자변경 등)를 수신하여 처리한다.
// 반환값: 갱신된 MEMCOIN 포인터, 계속 실행 여부(bool)
func InnerControlProcess(socket *zmq.Socket, memcoin *Mb.MEMCOIN)(*Mb.MEMCOIN, bool){
	newmemcoin := &Mb.MEMCOIN{}
	flag := true

	msg, err := socket.RecvMessage(0)
	if err != nil{
		Log.Warn("InnerControl 수신 오류: %v", err)
		return memcoin, true
	}

	//
	// -. msg[0] = 명령어
	//    "CHANGEDATE" : 일자변경
	//    "STOP"       : 매칭종료
	// -. msg[1] = 코인심볼
	//
	switch msg[0] {
	case "CHANGEDATE" :
		Log.Infor(
			"COIN[%s] 일자변경[%s]. 주문번호 초기화. Map정리", 
			memcoin.GetSymbol(), msg[1])

		// 주문번호 초기화 
		memcoin.InitOrderSeq()
		
		symbol := memcoin.GetSymbol()
		path := os.Getenv("ORDPATH")

		// Unmap
		memcoin.UnLoadMEMCOIN()
		Log.Debug("[%s] UnLoadMEMCOIN OK and Map정리 시작", symbol)

		// 코인 Mapfile 정리
		err := CleanOrderMapFile(path, symbol)
		if err != nil { // 
			Log.Error("[%s]Map정리 :%v", symbol, err);
			newmemcoin = nil
			flag = false
		}else{
			Log.Infor("[%s] Map정리 후 Reload 시작", symbol)
			newmemcoin.LoadMEMCOIN(path, symbol)
			Log.Infor("[%s] Map정리 후 Reload OK", symbol)
			flag = true
		}
		Log.Infor("COIN[%s] 일자변경. Map정리 성공", symbol)
		return newmemcoin, flag

	case "STOP" : // 종료 Message
		if len(msg) < 2{ 
			Log.Warn("비정상 STOP Message : %v", msg)
			return memcoin, true;
		}
		if msg[1] == memcoin.GetSymbol() {
			Log.Infor("%s STOP 수신", memcoin.GetSymbol())
			flag = false
		}
	}	
	return memcoin, flag
}
/////////////////// 호가  등록 ///////////////////////////////////////////////
func addNewHoga(memcoin *Mb.MEMCOIN, 
		mode     Mb.BuySell, 
		order   *Mb.Order){
Log.Debug("\t호가 등록")
	// 1. 호가를 찾는다
	hoga := memcoin.FindHoga(mode, order.Price)
	if hoga == nil{
		// 1.1 없으면 비어있는  호가 찾기
Log.Debug("\t\t호가 없음")
		hoga = memcoin.FindEmptyHoga(mode)
		if hoga == nil{
			//1.1.1 비어있는 호가도 없다. NewHoga는 OutWindow상태
Log.Debug("\t\t빈 호가(X) 없음")
			ha, err := memcoin.NewHoga(order.Price, mode)
			if err != nil{
				Log.Error(
					"COIN[%s] 호가생성 오류: %v\n%s",
					memcoin.GetSymbol(), 
					err,
					string(order.OrdPacket[:]))
				os.Exit(0)
			}
			ha.InWindow()
			err = ha.InsertOrder(order) // IsWin 자동처리
			if err != nil {
				Log.Error(
				"COIN[%s] 호가생성 후  주문등록 : %v\n%s",
				memcoin.GetSymbol(), 
				err,
				string(order.OrdPacket[:]))
				os.Exit(0)
			}
			// 1.1.3 호가등록
			if mode == Mb.MEDO {
				memcoin.Medo = append(memcoin.Medo, *ha)
			}else{
				memcoin.Mesu = append(memcoin.Mesu, *ha)
			}
		}else{ // ------- FindEmptyHoga()
Log.Debug("\t\t빈 호가(O) 있음")
			// 1.1.2 주문등록
			hoga.InWindow()
			err := hoga.InsertOrder(order) // IsWin 자동처리
			if err != nil {
				Log.Error(
				"COIN[%s] 호가생성 후  주문등록 : %v\n%s",
				memcoin.GetSymbol(), 
				err,
				string(order.OrdPacket[:]))
				os.Exit(0)
			}
			hoga.SetPrice(order.Price)	
		}
		// 1.1.4 호가 Sorting
		memcoin.SortHoga(mode)
	}else{ // --------------- FindHoga()
		// 1.2 있으면 주문 등록
		err := hoga.InsertOrder(order) //<- IsWin 처리 자동
		if err != nil{
			Log.Error(
			"COIN[%s] hoga.InsertOrder()\n%v\n%s", 
			memcoin.GetSymbol(), 
			err,
			string(order.OrdPacket[:]))
			os.Exit(0)
		}
	}
	// Window 정리
	memcoin.SetWindow(mode)
}
/////////////////////////  체결수량 계산  Function ///////////////////////////
func calcCheQty(maker, taker *Mb.Order, remcp int)Bd.BinDecimal{
	var cheqty Bd.BinDecimal
	switch remcp {
	case -1 : // maker < taker
Log.Debug("REM_COMP : maker < taker ")
		cheqty = maker.RemQty
		maker.RemQty.Set("0.0") // maker.RemQty =0
		taker.RemQty,_ = taker.RemQty.Sub(cheqty,
				Bd.NOROUND, 8) // taker.RemQty -= cheqty
	case  0 : // maker == taker
Log.Debug("REM_COMP : maker == taker")
		cheqty = maker.RemQty
		maker.RemQty.Set("0.0") // maker.RemQty = 0
		taker.RemQty.Set("0.0") // taker.RemQty = 0
	case  1 : // maker > taker
Log.Debug("REM_COMP : maker > taker")
		cheqty = taker.RemQty
		maker.RemQty,_ = maker.RemQty.Sub(cheqty,
				Bd.NOROUND, 8)  // make.RemQty -= cheqty
		taker.RemQty.Set("0.0") // taker.RemQty = 0
	} // Switch
	return cheqty
}

////////////////////// 지정가 Matchin function ////////////////////////////
func limitMatching(ipcpub *zmq.Socket,
				memcoin *Mb.MEMCOIN,
				makerMode Mb.BuySell,
				taker   *Mb.Order)bool{
	const (
		SORT_HOGA   = 0x10   // MEMHOGA Sort
		SORT_ORDER  = 0x01   // MEMHOGA 내 Order Sort
	)
	var (
		sortFlag    int
		takerIsZero bool
		cheqty      Bd.BinDecimal
		maker      *Mb.Order
		prcComp     int
	)

	makerHoga := memcoin.Mesu
	if makerMode == Mb.MEDO{
		makerHoga = memcoin.Medo
	}
	sortFlag = 0
	takerIsZero = false  // Taker 잔량 있음
	for idx := range makerHoga{
		mHoga := &makerHoga[idx] // 🔴㉽ Pointer Handling *MEMHOGA (Maker) 
		err := mHoga.InWindow() // Map
		if err != nil{
			Log.Error("COIN[%s] :%v", memcoin.GetSymbol(), err)
			os.Exit(0)
		}
		prcComp = Bd.DecimalCompare(mHoga.Price, taker.Price)
		// prcComp  1 : maker > taker
		//          0 : maker = taker
		//         -1 : maker < taker
		if makerMode == Mb.MEDO && prcComp == 1 { 
			break 	
		}else if makerMode == Mb.MESU && prcComp == -1{
			break
		}
		Log.Debug("INDEX [%d] Hoga.Prc[%s]", idx, mHoga.Price.String())

		cnt := mHoga.OrderCount // 주의 : 아래 Loop에서 OrderCount 변함

		for i:=0; i < cnt; i++{  // 호가 내 주문 Loop

			maker,_ = mHoga.GetOrdPointerByIndex(i) // 🔴㉽ Pointer Handling

			// 주문이 마지막인가
			if maker.OrdNum == Mb.ORDSEQ_MAX { break;}

			// 체결수량, maker/taker의 주문 잔량 조정
			// remcp은 아래에서 잔량 CHECK용도로 사용함
			// remcp -1 : maker < taker
			//        0 : maker = taker
			//        1 : maker > taker
			remcp := Bd.DecimalCompare(maker.RemQty, taker.RemQty)
			cheqty = calcCheQty(maker, taker, remcp)	

Log.Debug("cheqty [%s]", cheqty.String())
			// maker 최종 체결번호 증가(0 부터 시작)
			maker.CheSeq ++
			// taker 최종 체결번호 증가(0 부터 시작)
			taker.CheSeq ++

			// IPCCHE는 
			// 주문증거금 해소를 위해
			// 잔량(maker,taker).RemQty 계산 완료 후 전송 해야함
			// 잔량이 "0"인 경우 
			// 🔴 Send 체결 Packet (maker and taker)
			// 체결 PairKey 생성
			pairkey := SnowNode.Generate()
			chetime := uint64(time.Now().UnixMicro()) // 체결시각
			ipcpub.SendMessage("IPCCHE",
								api.GetBytes(&pairkey), 
								api.GetBytes(&chetime),
								cheqty.String(),
								api.GetBytes(maker),
								api.GetBytes(taker),
			)
			// MEMCOIN 최종 체결가격 갱신
			memcoin.MstInfor.BfPrice.Set(maker.Price)
			// MEMCOIN 최종 체결시각 갱신 (Micro)
			memcoin.MstInfor.BfTime = chetime
			// maker MEMHOGA 주문수량 & 주문 Count 차감
			// (㉼ maker-MESU, taker-MEDO)
			mHoga.SubSumQty(cheqty)
			
			if remcp != 1{ // maker 잔량소진 (잔량이 maker <= taker 인경우)
				maker.OrdNum = Mb.ORDSEQ_MAX // 주문 Clear
				maker.Margin.Set("0.0") // 주문 증거금
				maker.Fee.Set   ("0.0") // 체결 증거금
				mHoga.DecOrderCount()  // MEMHOGA OrderCount --
				// MEMHOGA Sort는 매수 MEMHOGA LOOP 밖에서
				sortFlag |= SORT_ORDER
				if mHoga.OrderCount == 0{
					if mHoga.Hd.BuySell == 'B' {
						mHoga.SetPrice(Mb.MIN_PRICE)  // 매수: MIN_PRICE
					}else{
						mHoga.SetPrice(Mb.MAX_PRICE)  // 매도: MAX_PRICE
					}
					sortFlag |= SORT_HOGA
				}

			} else{ // maker잔량 있음 (maker > taker)
				//매수(maker)  잔량이 남아 있음 증거금 차감
				// Margin = 주문 가격 X 체결수량(주의:체결수량)
				// Fee    = Margin X MAX(MakerRate, TakerRate)
				// 내림(버림)
				mg,_ := maker.Price.Mul(cheqty, Bd.ROUND_DN, 8)
				fe,_ := mg.Mul(Bd.DecimalMax(maker.MakeRate, 
								maker.TakeRate),
								Bd.ROUND_DN, 8)

				maker.Margin,_ = maker.Margin.Sub(mg, Bd.ROUND_DN, 8)
				maker.Fee,_    = maker.Fee.Sub   (fe, Bd.ROUND_DN, 8)
			}
			// taker 잔량이 0 이면
			if remcp != -1{// DecimalCompare(make.RemQry, taker.RemQry) 
				takerIsZero = true
				break
			}else if taker.BuySell == 'B'{
				// taker 잔량이 남아 있고 매수인경우 
				// 체결 수량만큼 taker 증거금 차감 
				mg,_ := taker.Price.Mul(cheqty, Bd.ROUND_DN, 8)
				fe,_ := mg.Mul(Bd.DecimalMax(taker.MakeRate, 
							taker.TakeRate),
							Bd.ROUND_DN, 8)

				taker.Margin,_ = taker.Margin.Sub(mg, Bd.ROUND_DN, 8)
				taker.Fee,_    = taker.Fee.Sub   (fe, Bd.ROUND_DN, 8)
			}
		} // Maker 호가 내 주문 Loop

		// 현재의 상태는
		// maker의 한 호가 모두 소진 되었거나 (다음 MEMHOGA)
		// taker 호가가 모두 소진 된경우

		// Hoga내의 주문 Sort(maker의 호가가 모두 소진일 경우 필요 없음)
		if sortFlag&SORT_ORDER != 0 && mHoga.OrderCount != 0{
			Log.Debug("매수 주문 Sorting")
			mHoga.SortOrder() // OrderNum 오름(매도)/내림(매수) 차순
		}
		// taker의 잔량이 모두 소화 된 상태
		if takerIsZero == true { break; }		

		// 다음 MEMHOGA를 위해 SORT_ORDER Clear
		sortFlag &^= SORT_ORDER
				
	} // Maker 전체 MEMHOGA Loop (매칭 Main Loop)

	if sortFlag&SORT_HOGA != 0 { // maker Hoga Sort
		Log.Debug("매수 호가 Sorting")
		memcoin.SortHoga(makerMode) // Price 내림차순
	}
	// Maker Window 정리
	//memcoin.WindowProcessing(makerMode)
	memcoin.SetWindow(makerMode)

	return takerIsZero
} 

/////////////////////// 지정가 /////////////////////////////////////////
//  매도  매수
// ─────┬─────               │   Tacker            │     Taker
//  152 │                    │     매도            │       매수
//  151 │          ──────────┼─────────────────────┼────────────────────
//  150 │           매   칭  │  130 >= taker.Price │  150 <= taker.Price
// ─────┼─────     ──────────┼─────────────────────┼────────────────────
//      │ 130       호가등록 │  130 < taker.Price  │  150 > taker.Price
//      │ 129       
//      │ 128       ✓ 130: 매수 1호가,  150:매도 1호가
//      │ 110       ✓ 매칭이후 잔량이 있을 경우 신규호가등록
//////////////////////////////////////////////////////////////////////////// 
// Parameter 
// ipcpub  : 내부 Proxy socket
// memcoin : 처리 COIN (최상위 Structure Pointer)
// inner   : Tacker의 Inner packet (Pointer)
// jsonpk  : Tacker의 주문 JSON
//////////////////////////////////////////////////////////////////////////// 
func limitOrder(ipcpub  *zmq.Socket, 
				memcoin *Mb.MEMCOIN, 
				inner   *Mb.InnerOrdPacket,
				jsonpk   Mb.OrderJson){ 

	var (
		taker      *Mb.Order
		takerIsZero bool // Taker 잔량 0
	)

	taker = &inner.MemData

	switch(taker.BuySell){
	case 'S':   // 매도 주문
		// 매도주문(taker) 가격이 매수1호가 보다 크면, 매도 신규등록
		//                                     Decimal, BinDecimal
		cmp := Bd.DecimalCompare(memcoin.Mesu[0].Price, taker.Price)

Log.Debug("MEDO - 매수1호가[%s] taker 가격[%s] 주번[%d]Compare[%d]",
			memcoin.Mesu[0].Price.String(), taker.Price.String(), 
			taker.OrdNum, cmp)

		if cmp == -1{// ----호가(O),체결/시세(X)---  a < b, 매도 신규등록
			addNewHoga(memcoin, Mb.MEDO, taker)
		}else{       //-----호가(O), 체결/시세(O)-- a >=b ---- 매도 매칭
			takerIsZero = limitMatching(ipcpub, memcoin, Mb.MESU, taker)
			if takerIsZero == false{
				// 매칭후 Taker 잔량이 있으면 호가 등록
				Log.Infor("Taker 매칭 후 잔량 있음, MEDO등록")
				addNewHoga(memcoin, Mb.MEDO, taker)
			}
		}
	case 'B' : // 매수주문
		cmp := Bd.DecimalCompare(memcoin.Medo[0].Price, taker.Price)
Log.Debug("MESU - 매도1호가[%s] taker 가격[%s] Compare[%d]",
			memcoin.Medo[0].Price.String(), taker.Price.String(), cmp)
		if cmp == 1{// ----호가(O),체결/시세(X)---  a > b, 매수 신규등록
			addNewHoga(memcoin, Mb.MESU, taker)
		}else{      //-----호가(O), 체결/시세(O)-- a <=b ---- 매수 매칭
			takerIsZero = limitMatching(ipcpub, memcoin, Mb.MEDO, taker)
			if takerIsZero == false{
				// 매칭후 Taker 잔량이 있으면 호가 등록
				Log.Infor("Taker 매칭 후 잔량 있음, MESU등록")
				addNewHoga(memcoin, Mb.MESU, taker)
			}
		}
	}
}

////////////////////// 시장가 Matchin function ////////////////////////////
func marketMatching(ipcpub *zmq.Socket,
				memcoin *Mb.MEMCOIN,
				makerMode Mb.BuySell,
				taker   *Mb.Order)bool{
	const (
		SORT_HOGA   = 0x10   // MEMHOGA Sort
		SORT_ORDER  = 0x01   // MEMHOGA 내 Order Sort
	)
	var (
		sortFlag    int
		takerIsZero bool
		cheqty      Bd.BinDecimal
		maker      *Mb.Order
	)

	makerHoga := memcoin.Mesu
	if makerMode == Mb.MEDO{
		makerHoga = memcoin.Medo
	}
	sortFlag = 0
	for idx := range makerHoga{
		mHoga := &makerHoga[idx] // 🔴㉽ Pointer Handling *MEMHOGA (Maker) 
		if  Bd.DecimalCompare(mHoga.Price, Mb.MAX_PRICE) == 0 ||
		 	Bd.DecimalCompare(mHoga.Price, Mb.MIN_PRICE) == 0{
			Log.Debug("시장가 주문 Maker 마지막. 매칭종료")
			break;
		}
		err := mHoga.InWindow() // Map
		if err != nil{
			Log.Error("COIN[%s] :%v", memcoin.GetSymbol(), err)
			os.Exit(0)
		}
		//  taker   │    price    │  RemQty     │ Margin  │ Fee  │ 
		// ─────────┼─────────────┼─────────────┼─────────┼──────┼
		//  매수(B) │ maker.Price │ 수량계산    │ margin  │ fee  │
		// ─────────┼─────────────┼─────────────┼─────────┼──────┼
		//  매도(S) │ maker.Price │주문가능 수량│  0.0    │ 0.0  │
		// ─────────┴─────────────┴─────────────┴─────────┴──────┴
		// 주문가능금액 = margin + fee
		// ㉼시장가경우 taker.RemQty는 goRecvOrder에서 설정하지 않음
		//   주문당시 taker.Price에 매수주문금액 있음이
		//            taker.Qty에   매도 수량이 있음
		//   매수인 경우
		//     fee    = 주문금액 * taker.TakeRate
		//     margin = 주문금액 - fee 로 계산함

		// 시장가 매수인 경우 주문잔량(RemQty) 설정
		if taker.BuySell == 'B'{
			// 수량 계산
			// ex) 100원 5개가 매수체결을 완료하기 위한 금액 계산
			//    "체결 금액 + 체결 수수료"가 필요,
			//     한 개의 수량에 필요한 금액 = (1 + 체결 수수료 율) X 가격
			//    따라서, 주문가능금액 / ( (1 + 체결 수수료 율) X 가격)을 분모로 해야 함
			// qty = (margin + fee) / (1 + taker.TakeRate) X mHoga.Price
			margin, _ := taker.Margin.Add(taker.Fee, Bd.NOROUND, 8)
			unit,_ := Bd.NewFromString("1.0")
			unit,_ = unit.Add(taker.TakeRate, Bd.NOROUND, 8)
			unit ,_ = unit.Mul(mHoga.Price , Bd.NOROUND, 8)
			remQty, _ := margin.Div(unit, Bd.ROUND_DN, 8) // 버림
			Log.Infor("시장가수량 =  주문가능금액 / ( ( 1 + TakerRate) X 가격 ), %s = (%s) / %s", 
				remQty.String(), margin.String(), unit.String())

			// 수량 정규화 
			// RemQty = RemQty - (RemQty % MinTrade)
			val, _:= remQty.Mod(taker.MinTrade)
			taker.RemQty,_ = remQty.Sub(val,Bd.ROUND_DN, 8)
			Log.Infor("잔량 정규화 : %s = %s - %s", 
				taker.RemQty.String(), remQty.String(), val.String())
			if Bd.DecimalIsZero(taker.RemQty) == true{
				Log.Warn("시장가 최소주문수량 조건(매칭종료) [%s:%c] Rem[%s] Min[%s] margin[%s] Fee[%s]",
					memcoin.GetSymbol(), taker.BuySell, 
					taker.RemQty.String(),
					taker.MinTrade.String(),
					taker.Margin.String(),
					taker.Fee.String())

				// taker.RemQty == 0, 매칭종료
				break
			}
		}

		// 시장가 매도 또는 매수, 가격을 Maker가격으로 
		taker.Price.Set(mHoga.Price)

Log.Infor("🔴시장가 주문 [%s:%c] Price[%s] Qty[%s] Rem[%s] margin[%s] Fee[%s]",
	memcoin.GetSymbol(), taker.BuySell, 
	taker.Price.String(),
	taker.Qty.String(),
	taker.RemQty.String(),
	taker.Margin.String(),
	taker.Fee.String())

		cnt := mHoga.OrderCount // 주의 : 아래 Loop에서 OrderCount 변함

		for i:=0; i < cnt; i++{  // 호가 내 주문 Loop(가격 같음)

			maker,_ = mHoga.GetOrdPointerByIndex(i) // ㉽🔴 Pointer Handling

			// 주문이 마지막인가
			if maker.OrdNum == Mb.ORDSEQ_MAX { break;}

			// 체결수량, maker/taker의 주문 잔량 조정
			// remcp은 아래에서 잔량 CHECK용도로 사용함
			// remcp -1 : maker < taker
			//        0 : maker = taker
			//        1 : maker > taker
			remcp := Bd.DecimalCompare(maker.RemQty, taker.RemQty)
			// 체결 수량 계산
			cheqty = calcCheQty(maker, taker, remcp)
//Log.Debug("cheqty [%s]", cheqty.String())
			// maker 최종 체결번호 증가(0 부터 시작)
			maker.CheSeq ++
			// taker 최종 체결번호 증가(0 부터 시작)
			taker.CheSeq ++

			// IPCCHE는 주문증거금 해소를 위해
			// 잔량(maker,taker).RemQty 계산 완료 후 전송 해야함
			// (calcCheQty() 이후 전송)

			// 🔴 Send 체결 Packet (maker and taker)
			// 체결 PairKey 생성
			pairkey := SnowNode.Generate()
			chetime := uint64(time.Now().UnixMicro()) // 체결시각
			ipcpub.SendMessage("IPCCHE",
								api.GetBytes(&pairkey), 
								api.GetBytes(&chetime),
								cheqty.String(),
								api.GetBytes(maker),
								api.GetBytes(taker),
			)

			// MEMCOIN 최종 체결가격 갱신
			memcoin.MstInfor.BfPrice.Set(maker.Price)
			// MEMCOIN 최종 체결시각 갱신 (Micro)
			memcoin.MstInfor.BfTime = chetime
			// maker MEMHOGA 주문수량 & 주문 Count 차감
			// (㉼ maker-MESU, taker-MEDO)
			mHoga.SubSumQty(cheqty)

			switch remcp {
				case -1 : // maker < taker (maker.RemQty = 0)
					// Maker 처리 (잔량 소진)
					maker.OrdNum = Mb.ORDSEQ_MAX
					if maker.BuySell == 'B' {
						maker.Margin.Set("0.0") // 주문 증거금
						maker.Fee.Set   ("0.0") // 체결 증거금
					}
					mHoga.DecOrderCount() // MEMHOGA OrderCount --
					if mHoga.OrderCount == 0{
						if mHoga.Hd.BuySell == 'B' {
							mHoga.SetPrice(Mb.MIN_PRICE)  // 매수: MIN_PRICE
						}else{
							mHoga.SetPrice(Mb.MAX_PRICE)  // 매도: MAX_PRICE
						}
					}
					sortFlag |= SORT_ORDER
					// 다음 주문
				case  0 : // maker = taker (RemQty = 0)
					// Maker 처리 
					maker.OrdNum = Mb.ORDSEQ_MAX
					if maker.BuySell == 'B' {
						maker.Margin.Set("0.0") // 주문 증거금
						maker.Fee.Set   ("0.0") // 체결 증거금
					}
					mHoga.DecOrderCount() // MEMHOGA OrderCount --
					if mHoga.OrderCount == 0{
						if mHoga.Hd.BuySell == 'B' {
							mHoga.SetPrice(Mb.MIN_PRICE)  // 매수: MIN_PRICE
						}else{
							mHoga.SetPrice(Mb.MAX_PRICE)  // 매도: MAX_PRICE
						}
					}
					sortFlag |= SORT_ORDER
					// 매칭 종료
				case  1 : // maker > taker
					// Maker 처리
					if maker.BuySell == 'B'{
					//매수(maker)  잔량이 남아 있음 증거금 차감
					// Margin = 주문 가격 X 체결수량(주의:체결수량)
					// Fee    = Margin X MAX(MakerRate, TakerRate)
					// 내림(버림)
						mg,_ := maker.Price.Mul(cheqty, Bd.ROUND_DN, 8)
						fe,_ := mg.Mul(Bd.DecimalMax(maker.MakeRate, 
								maker.TakeRate),
								Bd.ROUND_DN, 8)

						maker.Margin,_ = maker.Margin.Sub(mg, Bd.ROUND_DN, 8)
						maker.Fee,_    = maker.Fee.Sub   (fe, Bd.ROUND_DN, 8)
					}
					//매칭 종료
			} // Switch remcmp
			// Taker 증거금 처리. 수량 처리는 calcCheQty()에서...
			if taker.BuySell == 'B' {
				mg,_ := taker.Price.Mul(cheqty, Bd.ROUND_DN, 8)
				fe,_ := mg.Mul(taker.TakeRate, Bd.ROUND_DN, 8)

				taker.Margin,_ = taker.Margin.Sub(mg, Bd.ROUND_DN, 8)
				taker.Fee,_    = taker.Fee.Sub   (fe, Bd.ROUND_DN, 8)
			}
			// 매칭종료
			// remcp -1 : maker < taker
			//        0 : maker = taker
			//        1 : maker > taker
			if remcp != -1 {
				Log.Infor("매칭후 종료 조건 maker >= taker. [%s:%c] margin[%s] fee[%s]",
					memcoin.GetSymbol(),taker.BuySell,
					taker.Margin.String(), taker.Fee.String())
				break
			}		

		} // Maker 호가 내 주문 Loop

		// 현재의 상태는
		// maker의 한 호가 모두 소진 되었거나 (다음 MEMHOGA)
		// taker 호가가 모두 소진 된경우

		// Hoga내의 주문 Sort(maker의 호가가 모두 소진일 경우 필요 없음)
		if sortFlag&SORT_ORDER != 0 && mHoga.OrderCount != 0{
			Log.Debug("매수 주문 Sorting")
			mHoga.SortOrder() // OrderNum 오름(매도)/내림(매수) 차순
		}

		// 다음 MEMHOGA를 위해 SORT_ORDER Clear
		sortFlag &^= SORT_ORDER
				
	} // Maker 전체 MEMHOGA Loop (매칭 Main Loop)

	if sortFlag&SORT_HOGA != 0 { // maker Hoga Sort
		Log.Debug("매수 호가 Sorting")
		memcoin.SortHoga(makerMode) // Price 내림차순
	}
	// Maker Window 정리
	//memcoin.WindowProcessing(makerMode)
	memcoin.SetWindow(makerMode)

	// 매도 시장가. MKDONE을 위한 설정
	if taker.BuySell == 'S' { 
		taker.Margin.Set(taker.RemQty)
		taker.Fee.Set("0.0")
	}
	if Bd.DecimalIsZero(taker.Margin) == true &&
	   Bd.DecimalIsZero(taker.Fee) == true{
		takerIsZero = true
	}else{
		takerIsZero = false
	}

	Log.Infor("시장가 종료 [%s:%c] isZero[%t] margin[%s] fee[%s]",
		memcoin.GetSymbol(), taker.BuySell, takerIsZero, 
		taker.Margin.String(), taker.Fee.String())

	return takerIsZero
} 

/////////////////////// 시장가 /////////////////////////////////////////
// ❗㉼ 시장가 주문 정보 taker.price , taker.qty
//          │   taker.price    │   taker.qty           
// ─────────┼──────────────────┼─────────────────
//  매수(B) │   주문가능 금액  │      0.0                
// ─────────┼──────────────────┼─────────────────
//  매도(S) │      0.0         │  주문가능 수량                
// ─────────┴──────────────────┴─────────────────
func marketOrder(ipcpub  *zmq.Socket, 
				memcoin *Mb.MEMCOIN, 
				inner   *Mb.InnerOrdPacket,
				jsonpk   Mb.OrderJson){ 
	var (
		taker      *Mb.Order
		takerIsZero bool // Taker 잔량 0
	)

	taker = &inner.MemData
	if (taker.BuySell == 'S' && memcoin.Mesu[0].OrderCount  == 0) ||
	   (taker.BuySell == 'B' && memcoin.Medo[0].OrderCount == 0){
		//상대 호가 없음
		Log.Error(
		"시장가 주문 [%s:%c] 상대호가  없음", 
		memcoin.GetSymbol(), taker.BuySell)
		jsonpk.State = "ORD-MA-4"
		jsonpk.Message = "상대호가가 없습니다."
		///// 오류 전송 
		jsonstream,_ := json.Marshal(&jsonpk)
		ipcpub.SendMessage("IPCERR","MARKET", jsonstream)
		return
	}

	switch(taker.BuySell){
	case 'S':   // 매도 주문
		// taker정보(goRecvOrder에서Set))
		// taker.Price = 0.0, taker.Qty = 시장가수량
		// taker.Margin = 0.0 taker.Fee = 0.0
		takerIsZero = marketMatching(ipcpub, memcoin, Mb.MESU, taker)
		if takerIsZero == false{
			// 매칭후 Taker 잔량이 있으면 
			// 증거금 해소 목적으로  IPCMKDON 전송
			// 잔량은 jsonpk.Margin에 설정한다
			jsonpk.Margin = taker.RemQty.String()
			Log.Infor("시장가 매도 후 잔량 있음, IPCMKDONE 전송")
			jsonstream,_ := json.Marshal(&jsonpk)
			// To gofAfterFunction
			ipcpub.SendMessage("IPCMKDONE","MARKET", jsonstream) 
		}
	case 'B' : // 매수주문
		// taker정보(goRecvOrder에서Set))
		// taker.Price = 시장가 주문 금액, taker.Qty = 0.0
		// taker.Fee = taker.Price(시장가 주문 금액 ) * taker.TakeRate
		// (사장가 주문은 항상 taker임)
		// taker.Margin = taker.Price - taker.Fee 
		takerIsZero = marketMatching(ipcpub, memcoin, Mb.MEDO, taker)
		if takerIsZero == false{
			// 증거금 해소 목적으로  IPCMKDON 전송
			// jsonpk.Margin과 jsonpk.Fee가 
			// 주문 증거금 해소용 임
			Log.Infor("시장가 매수 후 주문금 남았음. IPCMKDONE 전송")
			jsonpk.Margin = taker.Margin.String()
			jsonpk.Fee    = taker.Fee.String()
			jsonstream,_ := json.Marshal(&jsonpk)
			// To gofAfterFunction
			ipcpub.SendMessage("IPCMKDONE","MARKET", jsonstream)
		}
	}
}
// 취소주문 처리
// ㉽ 원주문과 취소주문의 UserId가 다른 경우 취소 불가
// 원 주문이 매수주문인 경우 
// jsonpk.Margin, jsonpk.Fee에는 주문증거금 해소금액을 넣는다
// jonspk.Qty에는 orgp.RemQty (원주문의 주문잔량-취소수량)을 넣는다
//
func cancelOrder(ipcpub *zmq.Socket, 
				memcoin *Mb.MEMCOIN, 
				inner   *Mb.InnerOrdPacket,
				jsonpk   Mb.OrderJson){

	var (
		orgp, cordp *Mb.Order // 원주문,취소주문 Pointer
		mode Mb.BuySell
	)

	cordp = &inner.MemData // 취소주문

	Log.Infor("취소주문 : 원주문번호(%v) 주문번호(%v)", 
			inner.OrgNum, cordp.OrdNum)

	//순서: 호가찾기 -> 원주문 찾기	-> 취소처리
	// 𝟭 호가 찾기
	if cordp.BuySell == 'S' {
		mode = Mb.MEDO
	} else  {
		mode = Mb.MESU
	}

	orgHoga := memcoin.FindHoga(mode, cordp.Price)
	if orgHoga == nil{
		Log.Error(
		"[%s:%s]원주문 호가 없음 : 원주문번호[%d] 호가[%s]",
				memcoin.GetSymbol(), mode.String(), inner.OrgNum, 
				cordp.Price.String())
		jsonpk.State = "ORD-MA-1"
		jsonpk.Message = "원주문 호가 없음. 원주문 가격을 확인하시오 "
		///// 오류 전송 
		jsonstream,_ := json.Marshal(&jsonpk)
		ipcpub.SendMessage("IPCERR","CANCEL", jsonstream)
		return
	}
	// Map 되어 있는가
	saveWindow := orgHoga.IsWindow
	if saveWindow == false{ orgHoga.InWindow();}

	// 𝟮 원주문 찾기
	orgp = nil
	for i := 0; i < orgHoga.Hd.OrderCount; i++{
		op, err := orgHoga.GetOrdPointerByIndex(i)
		if err != nil{ 
			Log.Error("COIN[%s]%v", memcoin.GetSymbol(), err)
			orgp = nil 
			break 
		}			
		if op.OrdNum == inner.OrgNum{ // 찾음
			orgp = op
			break
		}
	}
	if orgp == nil{ // 없다
		msg := fmt.Sprintf(
		"[%s:%s] 주문가격(%s)에 취소를 위한 원주문번호(%d) 없음", 
		memcoin.GetSymbol(), mode.String(), jsonpk.Price , inner.OrgNum)

		Log.Error(msg)
		jsonpk.State = "ORD-MA-2"
		jsonpk.Message = msg

		if saveWindow == false{orgHoga.OutWindow();}	
		/// 오류전송
		jsonstream,_ := json.Marshal(&jsonpk)
		ipcpub.SendMessage("IPCERR","CANCEL", jsonstream)
		return
	}
	// 주문 UserId와 취소주문 UserId가 같은가?
	uid, _ := strconv.Atoi(jsonpk.UserId)
	if orgp.UserId != uid{
		msg := fmt.Sprintf(
		"[%s] 원주문과 취소주문의 UserId가 다릅니다. %d ≠  %d", 
		memcoin.GetSymbol(), orgp.UserId, uid)

		Log.Error("%s-원주문번호(%d) 취소주문번호(%s)",msg, 
			orgp.OrdNum, jsonpk.OrdNum)
		jsonpk.State = "ORD-MA-3"
		jsonpk.Message = msg

		jsonstream,_ := json.Marshal(&jsonpk)
		ipcpub.SendMessage("IPCERR","CANCEL", jsonstream)
		return
	}

	// 𝟯 취소 처리 -  호가 (O) 시세 (X)

	// 취소주문 Packet에 주문증거금 해소용 Set
	jsonpk.Margin = orgp.Margin.String()	
	jsonpk.Fee    = orgp.Fee.String()
	// 취소수량 설정
	jsonpk.Qty    = orgp.RemQty.String()

	// 주문 건수 감소
	orgHoga.DecOrderCount()

	// 주문 수량 감소
	// 취소수량 =  orgp.RemQty(원주문 잔량)
	orgHoga.SubSumQty(orgp.RemQty)

	// 원주문 hoga Memory에서 삭제
	orgp.OrdNum = Mb.ORDSEQ_MAX

	if orgHoga.OrderCount != 0{ // 호가에 남은 주문이 있으면
		// 주문이 취소되었음으로(ORDSEQ_MAX) 
		// 호가 내 주문 Sorting(주번 오름차순)
		orgHoga.SortOrder()
	}else{ // 호가에 남은 주문이 없으면
		// 취소 후 호가에 남은 주문이 없으면(비어있는 호가)
		// 3.2.1 호가 가격에 MAX(매도), MIN(매수)을 설정
		// 3.2.2 호가 Sorting (가격 순) 
		//      매도(오름차순) 매수(내림차순)
		if mode == Mb.MEDO{
			orgHoga.SetPrice(Mb.MAX_PRICE)
			memcoin.SortHoga(Mb.MEDO) // Hoga Sort
		}else{
			orgHoga.SetPrice(Mb.MIN_PRICE)
			memcoin.SortHoga(Mb.MESU) // Hoga Sort
		}
	}

	// 취소완료 Packet전송
	jsonpk.State = "0"
	jsonpk.Message = "취소주문이 완료되었습니다"
/*
	// 호가변동 처리 gofCoinMatching()에서 처리함 나중에 삭제 할 것 
	orgHoga := makeHogaString(memcoin, 0) 
	ipcpub.SendMessage("IPCHOGA", memcoin.GetSymbol(), orgHoga)
*/
	// 기존상태가 unMap인 경우 Out window 
	if saveWindow == false{orgHoga.OutWindow();}	

	// 취소주문 전송
	jsonstream,_ := json.Marshal(&jsonpk)
	ipcpub.SendMessage("IPCCANCEL", jsonstream)

	Log.Infor("취소주문 처리 완료: 원주문번호(%v) 주문번호(%v)", 
			inner.OrgNum, cordp.OrdNum)
	return 
}
/////////////////////////////////////////////////////////////////////////////
// Coin별 matching goRoutine (MAIN)
// ctx : zmq Context (Matching Process 에서는 동일한 Context를 사용 권장)
// symbol :  처리할 코인 symbol
// result :  초기 실행중 오류 상황 return
/////////////////////////////////////////////////////////////////////////////
func gofCoinMatching(ctx *zmq.Context, symbol string, result chan<- error){
	var sub,ipcpub, ipcsub *zmq.Socket

	// 주문 수신을 위한 준비
	s, err := ctx.NewSocket(zmq.SUB)
	if err != nil { 
		result <-  err
		return
	}		
	sub = s
	defer sub.Close()
	err = sub.Connect(OrdPoint)
	if err != nil { 
		result <- err
		return
	}

	sub.SetSubscribe(symbol) // Topic 등록

	// 내부전송을 위한 준비
	s, err = ctx.NewSocket(zmq.PUB)
	if err != nil { 
		result <- err
		return
	}
	ipcpub = s
	defer ipcpub.Close()

	ipcpub.Connect(InnerFront)
	time.Sleep(100*time.Microsecond) // Zmq. 완전한 connect를 위한 잠깐의 여유
	
	ordhome := os.Getenv("ORDPATH")
	if ordhome == ""{
		result <- fmt.Errorf("환경변수 ORDPATH가 없습니다. Coin[%s]", symbol)
		return
	}

	// 내부 control Message "CHANGEDATE" 등
	ipcsub, _ = ctx.NewSocket(zmq.SUB)
	defer ipcsub.Close()

	ipcsub.Connect(InnerBack)
	ipcsub.SetSubscribe("CHANGEDATE") // 일자변경
	ipcsub.SetSubscribe("STOP") // 종료 

	Log.Infor("[%s]ORDPATH : %s", symbol, ordhome)

	// 주문정보 Loading
	// 주문 정보가 없는 경우 생성
	memcoin := &Mb.MEMCOIN{}
	err = memcoin.LoadMEMCOIN(ordhome, symbol)
	if err != nil{
		result <- fmt.Errorf("LoadMEMCOIN(%s)\n%v", symbol, err)
		return
	}
	defer memcoin.UnLoadMEMCOIN()

	poller := zmq.NewPoller()
	poller.Add(sub, zmq.POLLIN)
	poller.Add(ipcsub, zmq.POLLIN)

	// 주문번호 검사(일자가 다르면 주문번호 초기화)
	str := fmt.Sprintf("%d", memcoin.MstInfor.OrdNum)
	str = str[0:8]		
	tm := time.Now()
	today := tm.Format("20060102")
	if str != today {
		Log.Infor("주문번호 일자 다름. 오늘[%s] ≠ OLD [%s:%d]", 
			today, str, memcoin.MstInfor.OrdNum)
		memcoin.InitOrderSeq()
	}
	
	Log.Infor("Matching 시작. COIN[%s], 주문번호[%d], 수신[%s]", 
		symbol,  memcoin.MstInfor.OrdNum, OrdPoint)

	var runFlag bool
	runFlag = true

	result <- nil   // Start ok

	// 호가 전송(초기 기동시 호가 1회 전송)
	hstr := makeHogaString(memcoin, 0)
	ipcpub.SendMessage("IPCHOGA", memcoin.GetSymbol(), hstr)

	for runFlag { // Main Loop
		polls, _:= poller.Poll(1000 * time.Millisecond)

		for _, polled := range polls{
			switch s:= polled.Socket;s{
			case sub : // 주문 SUB

				var (
					ordp         *Mb.Order
					jsonpk        Mb.OrderJson
					inner        *Mb.InnerOrdPacket
				)

				msg, err := sub.RecvMessageBytes(0)
				if err != nil {
					Log.Error("%s:%v", symbol, err)
					time.Sleep(1000*time.Millisecond)
					os.Exit(0)
				}
				// TOPIC -> ex) BTC.KRW
				topic := string(msg[0])

				// Cancel(int) + OrgNum(uint64) + MemData(Order)
				inner = (*Mb.InnerOrdPacket)(unsafe.Pointer(&msg[1][0]))
				ordp = &inner.MemData

				// 주문 Packet -> Json structure
				err = json.Unmarshal(api.CleanByte(ordp.OrdPacket[:]), 
										&jsonpk)
				if err != nil {
					Log.Error("Unmarshal : %v\n%s", 
							err, string(ordp.OrdPacket[:]))
					time.Sleep(1000*time.Millisecond)
					os.Exit(0)
				}

				// 주분번호 채번
				ordp.OrdNum = memcoin.GetOrderSeq() 
				ordp.OrdTime = uint64(time.Now().UnixMicro())
				jsonpk.OrdNum = fmt.Sprintf("%d", ordp.OrdNum)
				jsonpk.OrdTime = fmt.Sprintf("%d", ordp.OrdTime)
				jsonstream,_ := json.Marshal(&jsonpk)
				copy(ordp.OrdPacket[:], jsonstream)

		Log.Infor("수신 Topic : %s OrdNum[%d:%s] RegNum[%d:%s", 
				topic, ordp.OrdNum, jsonpk.OrdNum, 
				ordp.RegNum, jsonpk.RegNum)

				// 채번 후 Update용 Packet 전송
				ipcpub.SendMessage("IPCSEQ", jsonstream)	
			
				// 🔴 FDS 전송 (호가 변동전에 보내야함)
				//    Topic : IPCFDS
				// Think About : FDS주문 보다 체결이 먼저 갈 경우?
				//               Fds처리를 Go Routine으로....
				//               전체적인 속도저하 우려
				sendFdsData(ipcpub, jsonpk, memcoin)

				// 취소주문
				if inner.Cancel == 1{ // --호가(O), 시세/체결(X)
					cancelOrder(ipcpub, memcoin, inner, jsonpk)
				}else{
					// 𝟬 매칭
					switch ordp.OrdType {
					case 'L' : // 지정가
						limitOrder(ipcpub, memcoin, inner, jsonpk)
					case 'M' : // 시장가
						marketOrder(ipcpub, memcoin, inner, jsonpk)
					default : // 주문유형 오류
					}
				}

				// 🔴 호가 전송 -- 호가(O),
				str := makeHogaString(memcoin, 0)  // Full Json string
				ipcpub.SendMessage("IPCHOGA", memcoin.GetSymbol(), str)

			case ipcsub : // Change date 등
				var newmemcoin *Mb.MEMCOIN
				newmemcoin, runFlag = InnerControlProcess(ipcsub, memcoin)
				if newmemcoin == nil {
					Log.Error("%s InnerControlProcess() 실패", symbol)
					break;
				}else{
					memcoin = newmemcoin
					Log.Infor("%s InnerControlProcess() 설공", symbol)
					Log.Infor("%s-주문번호[%d] 직전체결가[%s] 직전체결시각[%d] 매도IDX[%d] 매수IDX[%d]", 
						symbol,	
						memcoin.MstInfor.OrdNum, 
						memcoin.MstInfor.BfPrice.String(), 
						memcoin.MstInfor.BfTime,
						memcoin.MstInfor.MedoNameIdx, 
						memcoin.MstInfor.MesuNameIdx) 
				}
			} // Switch  polled.socket
		}// Poll loop
	}// Main Loop
	Log.Warn("%s 매매 Go Routine 종료", symbol)

	time.Sleep(time.Second)
	return
}
