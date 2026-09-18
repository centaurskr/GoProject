///
/// @brief   Test Memory mapped file 
/// @file    tmmap.go
/// @date    2024. 10. 08. (화) 15:22:43 KST
/// @author  Cento
///
package main
import (
	"fmt"
	"unsafe"
	"sort"
	"centolib/api"
	Dc "github.com/shopspring/decimal"
)
// Mapfile은 한 종목의 한 호가 단위로 ......
// Memory mapped file Hedaer for OrderBoard
type CoinInfo struct{
	Name     [32]byte    // Coin 명
	Mkidx    int32       // Market index
	Cidx     int32       // Coin index
	Mkcidx   int32       // 결제 통화 index
}

type Order struct{
	OrdType  int32       // 주문종류(1:LIMIT, 2:MARKET)
	BuySell  int32       // 매도수구분 (1:매도, 2:매수)
	OrdTime  uint64      // 주문시각, UinxNano(), 1728372731578396031
	OrdSeq   uint64      // 주문번호
	RegNum   uint64      // 접수번호
	CheSeq   uint32      // 최종체결번호

	Price    Dc.Decimal  // 주문가격
	Qty      Dc.Decimal  // 주문수량
	RemQty   Dc.Decimal  // 주문잔량
	MakeRate Dc.Decimal  // Maker 수수료율
	TakeRate Dc.Decimal  // Taker 수수료율
	MinTrade Dc.Decimal  // 최소 호가단위(결제통화기준)
	Using    Dc.Decimal  // 주문증거금
	Fee      Dc.Decimal  // 체결 예상 증거금

	OrdPacket [1024]byte // 주문 수신 JSON
}
type OrderHeader struct{
	Price    Dc.Decimal  // 가격
	Coin     CoinInfo    // Coin정보
	BuySell  int32       // 1:Buy, 2:Sell
	BufCount int32       // Map file의 Order Buffer 갯수
	OrdCount int32       // Map file의 Order 갯수
	OrdQty   Dc.Decimal  // 주문 총수량
}
//////////////////////////////////////////////////////////////////////////////
// 가격당 하나 씩(가격 하나당 Map File 하나)
type OrderInfo struct{
	IsMemory bool        // Memory:true, Mapfile:false
	Hd      OrderHeader  // Mapfile Header Copy (File에 있는 경우를 위해 Copy)

	pMap    []byte       // Memory Map pointer
	pHd     *OrderHeader // Memory Map pointer
	Orders  *Order       // Mapfile 주문 List(한호가의 주문List)
}
// Matching Board (Coin당 한개)
type MatchingBoard struct{
	Coin        CoinInfo       // Coin
	OrdSeq      uint64         // 주문번호(일별)
	BfPrice     Dc.Decimal     // 직전 체결가
	MedoQty     Dc.Decimal     // 매도 호가수량
	MesuQty     Dc.Decimal     // 매수 호가수량

	Medo       []OrderInfo
	Mesu       []OrderInfo
}

func (mb *MatchingBoard)Initialize(path string) error {

	return nil
}

////////////////////////////////////////////////////////////////////////////////
/// MEDO 가격별sroting(오름차순 정렬, 낮은가격이 먼저)
/// @fn     : func (mb *MatchingBoard)MedoSort(){
/// @param  :
/// @return :
////////////////////////////////////////////////////////////////////////////////
func (mb *MatchingBoard)MedoSort(){
	sort.Slice(mb.Medo, func(i, j int)bool{
		rtn := mb.Medo[i].Hd.Price.Compare(mb.Medo[j].Hd.Price)
		if rtn == -1 {  //i < j
			return true
		}else{  // i >= j
			return false
		}
	})
	return
}
////////////////////////////////////////////////////////////////////////////////
/// MESU가격별 sorting(내림차순정렬, 높은가격이 먼저)
/// @fn     : func (mb *MatchingBoard)MesuSort(){
/// @param  :
/// @return :
////////////////////////////////////////////////////////////////////////////////
func (mb *MatchingBoard)MesuSort(){
	sort.Slice(mb.Mesu, func(i, j int)bool{
		rtn := mb.Mesu[i].Hd.Price.Compare(mb.Mesu[j].Hd.Price)
		if rtn == 1 {  //i > j
			return true
		}else{  // i <= j
			return false
		}
	})
	return
}

////////////////////////////////////////////////////////////////////////////////
/// Medo Array(오름차순 정렬)안에서 가격으로 주문정보 찾기
/// @fn     : func (mb *MatchingBoard)MedoSearch(price Dc.Decimal) *OrderInfo
/// @param   price  찾을 가격
/// @return  성공 : OrderInfo pointer, 실패 : nil
////////////////////////////////////////////////////////////////////////////////
func (mb *MatchingBoard)MedoSearch(price Dc.Decimal) *OrderInfo{

	return nil
}

func main() {
	mapHd := &OrderHeader{
		BuySell  : 1,
		BufCount : 10,
		OrdCount : 20,
	}
	
	mapHd.Price, _ = Dc.NewFromString("999999999999.999999999999")
	name := "BTC.USDT"
	//copy(mapHd.Coin.Name[:], fmt.Sprintf("%s", "BTC.USDT"))
	copy(mapHd.Coin.Name[:], name)
	mapHd.Coin.Mkidx  = 1234
	mapHd.Coin.Cidx   = 4567
	mapHd.Coin.Mkcidx = 8904
	mapHd.OrdQty, _=Dc.NewFromString("123456789012.123456789")
	fmt.Printf("Create [%s] %+v\n", string(mapHd.Coin.Name[:]), *mapHd)

	byteHd := api.GetBytesOfStruct(mapHd)
	fmt.Println("api.GetBytes : ", byteHd, len(byteHd), unsafe.Sizeof(OrderHeader{}))

	/// Create map file
	mappoint, err := api.CreateMapfile("TEST.MAP", 1024)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Copy Header	
	copy(mappoint[:], byteHd)
	// Unmap
	api.UnMapfile(mappoint)

	// Attach for read OrderHeader
	fmt.Printf("\n\n\n========Attach file : TEST.MAP============= \n")
	pMap, _, err := api.AttachMapfile("TEST.MAP")
	if err != nil {
		fmt.Println(err)
		return
	}
	rdHd := (*OrderHeader)(unsafe.Pointer(&pMap[0]))
	fmt.Printf("COIN [%s]: %+v\n", string(rdHd.Coin.Name[:]), *rdHd)
	api.UnMapfile(pMap)
}
