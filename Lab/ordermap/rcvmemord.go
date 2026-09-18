//
// Test : 주문 수신 Test
//
package main
import (
	Log "centolib/logout"
	Mb "centolib/matchingboard"
	tbzmq "centolib/tbzmq"
	"centolib/api"
	"os"
	"unsafe"
)
func main(){
	Log.InitLogger("", Log.CONSOLE_MODE);defer Log.Close()
	if len(os.Args) != 2 {
		Log.Error("주문수신 Test Program")
		Log.Error("%s <coin>", os.Args[0])
		os.Exit(0)
	}
	// 주문 수신 SUB point 얻기
	rcvEndpoint, err := api.GetEnvValue("Platform.env", "ORDRCVPOINT")
	if err != nil{
		Log.Error("%v", err)
		os.Exit(0)
	}

	// MEMCOIN 생성
	ordhome := os.Getenv("ORDPATH")
	if ordhome == ""{
		Log.Error("환경 변수에 ORDPATH가 없습니다")
		os.Exit(1)
	}
	memcoin := &Mb.MEMCOIN{}
	err = memcoin.LoadMEMCOIN(ordhome, os.Args[1])
	if err != nil{
		Log.Error("LoadMEMCOIN(%s)\n\t%v", os.Args[1], err)
		os.Exit(1)
	}
	memcoin.Dump()

	// make SUB socket 
	var sub tbzmq.SockWrapper
	err = sub.OpenSubSocket0mq(rcvEndpoint); defer sub.Close()
	if err != nil{ Log.Error("OpenSubSocket0mq()\n\t%v", err);os.Exit(1)}
	
	// Add Subscribe topic
	err = sub.SetSubscribe0mq(os.Args[1])
	if err != nil{ Log.Error("SetSubscribe0mq()\n\t%v", err);os.Exit(1)}

	// 주문 수신 Loop
	var hoga *MEMHOGA
	for{

		topic, stream, err := sub.ReceiveTopic0mq()
		if err != nil{ Log.Error("ReceiveTopic0mq()\n\t%v", err);os.Exit(1)}

		order := ( *Mb.Order)(unsafe.Pointer(&stream[0]))
		order.OrdSeq = memcoin.GetOrderSeq() // 주문번호 채번

		// 0. FDS용 DATA 생성 및 전송

		Log.Infor("Topic [%s] OrdSeq : %d", topic, order.OrdSeq)
		order.Dump(1)
		
		// 1. 매칭

		// 2. 매칭 후 잔량이 없으면(.... ing)
		if 	order.IsZeroRemQty() == true{ 

		}
		// 3. 매칭 후 잔량이 있으면 호가 등록
		hoga = memcoin.FindHoga(Mb.BuySell(order.BuySell), order.Price)
		if hoga == nil{ // 호가 없음
			// 비어있는 호가 찾기
			hoga = memcoin.FindEmptyHoga(Mb.BuySell(order.BuySell))
			if hoga == nil{ // 비어있는 호가 없음
				// New Hoga (NewHoga 후에 Window in 상태)
				hoga, err := memcoin.NewHoga(order.Price, 
										Mb.BuySell(order.BuySell))
				if err != nil{
					Log.Error("호가생성 오류\n\t%v\n%s", 
						err, order.OrdPacket)
					os.Exit(0)
				}
				// 주문 insert (Insert 후 Window out 상태)
				err = hoga.InsertOrder(order)
				if err != nil{
					Log.Error("호가생성 후 주문 Insert\n\t%v\n%s", 
						err, order.OrdPacket)
					os.Exit(0)
				}
				// Hoga append
				if Mb.BuySell(order.BuySell) == Mb.MEDO {
					memcoin.Medo = append(memcoin.Medo, *hoga)
				}else{
					memcoin.Mesu = append(memcoin.Mesu, *hoga)
				}
				// Hoga Sort
				memcoin.SortHoga(Mb.BuySell(order.BuySell))
			}else{ // 비어있는 호가 있음
				// 1.가격 설정 (비어있는 호가)
				hoga.InWindow()

				hoga.SetPrice(order.Price)

				// 2. 주문 추가(hoga.InsertOrder)
				//    2.1 map
				//    2.2 주문 Copy
				//    2.3 MEMHOGA AddSumQty(), OrderCount ++
				//    2.4 HD      AddSumQty(), OrderCount ++
				//    2.5 unmap
				err :=hoga.InsertOrder(order)
				if err != nil {
					Log.Error("InertOrder(호가없음)\n%v\n%s", 
							err, order.OrdPacket)
					os.Exit(1)
				}
			}
		}else{ // 호가있음
			err :=hoga.InsertOrder(order)
			if err != nil {
				Log.Error("InertOrder(호가있음)\n%v\n%s", 
						err, order.OrdPacket)
				os.Exit(1)
			}
		}
	} // For Loop
}
