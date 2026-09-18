package main
import (
	Log "centolib/logout"
	Mb "centolib/matchingboard"
	"centolib/api"
	"strings"
	"unsafe"
	"time"
	"sort"
	"fmt"
)
var (
    MapPoint []byte
    Header   *Mb.MapHeader
    OrdCond  uintptr
    MapName  string
)
func LoadCoin()error{
    Log.Infor("주문가능 코인을읽는다")

    var coinList []string

    mapName, err := api.GetPlatformEnv("ORDERCOND_MAP")
    if err != nil{ return err; }

    Log.Infor("ORDERCOND_MAP : %s", mapName)

    idx := 1
    for {
        strOrd := fmt.Sprintf("ORD_%02d", idx)
        val, err := api.GetPlatformEnv(strOrd)
        if err != nil{ break; }
        list := strings.Split(strings.Trim(val, " :"), ":")
        coinList = append(coinList, list...)
        idx ++
    }
    if len(coinList) == 0{
        return fmt.Errorf("등록된 코인이 없습니다")
    }
    size := api.SizeOf(Mb.MapHeader{}) + api.SizeOf(Mb.OrderCondition{}) * len(coinList)
    MapPoint, err = api.CreateMapfile(mapName, size)
    if err != nil{ return err;}
    Header  = (*Mb.MapHeader)(unsafe.Pointer(&MapPoint[0]))
    OrdCond = uintptr(unsafe.Pointer(&MapPoint[unsafe.Sizeof(Mb.MapHeader{})]))
    // Header
    now := fmt.Sprintf("%s", time.Now().Format("2006 01 02 15:04:05.00"))
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
    Log.Infor("LoadCoin() OK! Coin 개수[%d]", len(coinList))
    // sorting
    sortSlice := api.SortSlice{
        StBase  : OrdCond,
        StCount : Header.Count,
        StSize  : uintptr(unsafe.Sizeof(Mb.OrderCondition{})),
        LessFunc : func (a, b unsafe.Pointer) bool{
            o1 := (*Mb.OrderCondition)(a)
            o2 := (*Mb.OrderCondition)(b)
            rtn := strings.Compare(string(o1.Symbol[:]), string(o2.Symbol[:]))
            return  rtn < 0 // 오름차순
        },
    }
    sort.Sort(sortSlice)
    Log.Infor("LoadCoin() OK! Coin 개수[%d]", len(coinList))
    return nil
}
func main(){
	Log.InitLogger("Tang", Log.CONSOLE_MODE)
	LoadCoin()
}

