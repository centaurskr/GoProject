// 주문 상태 정보 관련 함수 모음.
//
// OrderCond.map 구조 (~/Order/OrderCond.map):
//
//	┌─────────┬────┬───┐
//	│ BTC.USDT│  1 │ 1 │  G: Matching Group
//	├─────────┼────┼───┤  S(상태): 1(정상), 0(주문중지), 3(매도만 가능)
//	│ ETH.USDT│  2 │ 0 │
//	└─────────┴────┴───┘
package main
import (
	"centolib/api"
	Mb "centolib/matchingboard"
	Log "centolib/logout"
	"unsafe"
	"strings"
	"time"
	"sort"
	"os"
	"fmt"
)
var (
	MapPoint []byte
	Header   *Mb.MapHeader
	OrdCond  uintptr
	MapName  string
)
///////////////////////////////////////////////////////////////////////////////
// LoadCoin은 Platform.env의 ORDER_LIST에서 코인 목록을 읽어 OrderCond.map을 생성한다.
func LoadCoin()error{

	if MapPoint != nil{
Log.Warn(`
주문상태 기존 정보 있음.%s 
생성일-%s
Buffer Count - %d
Data   Count - %d
UnMapfile() 함`,
MapName,
string(Header.Create[:]),
Header.Buffer,
Header.Count,
)
			api.UnMapfile(MapPoint)
	}
	var coinList []string

	mapName, err := api.GetPlatformEnv("ORDERCOND_MAP")
	if err != nil{ return err; }
	MapName = mapName
	Log.Infor("ORDERCOND_MAP : %s", MapName)

	Log.Infor("주문가능 코인을읽는다")
	ordName, err := api.GetPlatformEnv("ORDER_LIST")
	if err != nil{return err;}
	Log.Infor("ORDER_LIST : %s", ordName)	
	coinList, err = Mb.GetAllCoins(ordName)
	if err != nil{return err;}
	if len(coinList) == 0{
		return fmt.Errorf("등록된 코인이 없습니다")
	}
	size := api.SizeOf(Mb.MapHeader{}) + api.SizeOf(Mb.OrderCondition{}) * len(coinList)
	MapPoint, err = api.CreateMapfile(MapName, size)
	if err != nil{ return err;}
	Header  = (*Mb.MapHeader)(unsafe.Pointer(&MapPoint[0]))
	OrdCond = uintptr(unsafe.Pointer(&MapPoint[unsafe.Sizeof(Mb.MapHeader{})]))
	// Header 
	now := fmt.Sprintf("%s", time.Now().Format("2006 01 02 15:04:05.000"))
	copy(Header.Create[:], []byte(now))
	Header.Buffer = len(coinList)
	Header.Count  = len(coinList)

	// Coin 정보
	for i, coin := range coinList{
		list := (*Mb.OrderCondition)(unsafe.Pointer(OrdCond + unsafe.Sizeof(Mb.OrderCondition{}) * uintptr(i)))
		copy(list.Symbol[:], []byte(coin))
		list.Group  = 0
		list.Status = 0
	}
	// sorting
	sortSlice := api.SortSlice{
		StBase  : OrdCond,
		StCount : Header.Count,
		StSize  : uintptr(unsafe.Sizeof(Mb.OrderCondition{})),
		LessFunc : func (a, b unsafe.Pointer) bool{
			o1 := (*Mb.OrderCondition)(a)
			o2 := (*Mb.OrderCondition)(b)
			rtn := strings.Compare(string(o1.Symbol[:]), string(o2.Symbol[:]))
			return  rtn < 1  // 오름차순
		},
	}
	sort.Sort(&sortSlice)
Log.Infor(`
주문상태 Load OK:%s 
생성일-%s
Buffer Count - %d
Data   Count - %d
UnMapfile() 함`,
MapName,
string(Header.Create[:]),
Header.Buffer,
Header.Count,
)
	return nil
}
///////////////////////////////////////////////////////////////////////////////
// ChangeOrderStatus는 특정 코인 symbol의 주문 상태와 그룹을 변경한다.
// status: 1(정상), 0(주문중지), 3(매도만 가능)
func ChangeOrderStatus(symbol string, group, status int)error{
	if Header == nil{
		return fmt.Errorf("Attach되지 않았음")
	}
	for idx := range Header.Count {
		item := (*Mb.OrderCondition)(unsafe.Pointer(OrdCond + unsafe.Sizeof(Mb.OrderCondition{}) * uintptr(idx)))
		if api.CleanString(item.Symbol[:]) == symbol{
			Log.Infor("[%s] 주문상태 변경 그룹(%d)상태(%d) 🡆  [%d:%d]",
					symbol, item.Group, item.Status, group, status)
			item.Group  = group 
			item.Status = status
			return nil
		}
	}
	return fmt.Errorf("상태변경: %s는 미 등록 상태입니다", symbol)
}
// ChangeGroupStatus는 특정 그룹의 모든 코인 주문 상태를 일괄 변경한다.
// status: 1(정상), 0(주문중지), 3(매도만 가능)
func ChangeGroupStatus(group, status int)error{
	if Header == nil{
		return fmt.Errorf("Attach되지 않았음")
	}
	for idx := range Header.Count {
		item := (*Mb.OrderCondition)(unsafe.Pointer(OrdCond + unsafe.Sizeof(Mb.OrderCondition{}) * uintptr(idx)))
		if item.Group == group{
			Log.Infor("[%s] 그룹[%d]상태 변경 [%d] 🡆  [%d]",
					api.CleanString(item.Symbol[:]), group, item.Status, status)
			item.Status = status
		}
	}
	return nil
}
// AttachOrderCondition은 주문 상태 Map 파일(ORDERCOND_MAP)을 메모리에 매핑한다.
func AttachOrderCondition()error{
	if Header != nil{
Log.Infor(`
	주문상태 기존 정보 있음.%s 
	생성일-%s
	Buffer Count - %d
	Data   Count - %d
	UnMapfile() 함`,
	MapName,
	string(Header.Create[:]),
	Header.Buffer,
	Header.Count,
	)
		api.UnMapfile(MapPoint)
	}
	val, err := api.GetPlatformEnv("ORDERCOND_MAP")
	if err != nil{ return err; }
	MapPoint, _, err = api.AttachMapfile(val)
	if err != nil { return err; }
	Header  = (*Mb.MapHeader)(unsafe.Pointer(&MapPoint[0]))
	OrdCond = (uintptr)(unsafe.Pointer(&MapPoint[0])) + uintptr(unsafe.Sizeof(Mb.MapHeader{}))
	MapName = val
Log.Infor(`
	주문상태 정보 Attach Ok
	MapName      : %s
	생성일       : %s
	Buffer Count : %d
	Data   Count : %d `,
	MapName,
	Header.Create,
	Header.Buffer,
	Header.Count,
	)
	return nil
}
// MakeNewCoinOrder는 신규 상장 코인의 주문 정보 파일과 호가 데이터를 생성한다.
func MakeNewCoinOrder(symbol string)error{
	val, err := api.GetPlatformEnv("DBCONFIG_YAML")
	if err != nil{ return err;}
	dbctx, err := api.ConnectDb(val)
	if err != nil {return err;}
	defer dbctx.Close()

	cinfor, err := Mb.GetCoinInfor(dbctx, symbol)
	if err !=nil{return err;}
	
	val, err = api.GetPlatformEnv("ORDPATH")
	if err != nil{ return err;}

	path := val

	tpath := fmt.Sprintf("%s/%s", strings.TrimSuffix(path, "/") , symbol)
	err = os.MkdirAll(tpath, 0777)
	if err !=nil{return err;}
	
	err = Mb.CreateOrderMasterInfor(path, cinfor)
	if err != nil{return err;}

	pmap, err := Mb.GetOrderMasterInfor(path, symbol)
	if err != nil{ return err;}
	defer api.UnMapfile(pmap)

	path = path + "/" + symbol
	mst := (*Mb.OrderMasterInfor)(unsafe.Pointer(&pmap[0]))

    // Master Name index update
    mst.MedoNameIdx = Mb.ORDER_COUNT
    mst.MesuNameIdx = Mb.ORDER_COUNT
	for i := 0; i < Mb.ORDER_COUNT; i ++{
		err:=Mb.CreateHoga(path, symbol, Mb.MEDO, i)
		if err != nil{return err;}
		err =Mb.CreateHoga(path, symbol, Mb.MESU, i)
		if err != nil{return err;}
	}
	return nil
}
