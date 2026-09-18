///
/// @brief  Matching Board Data 
/// @file   : matchingboardata.go
/// @date   : 2024. 10. 16. (수) 12:30:56 KST
/// @author : Cento
///                 12345678901234567890
/// uint64 최대값 : 18446744073709551615    ( 9(20) )
// 주문번호       :  yymmdd + 0000000000000 (13자리-1조)  
//                   (yymmdd + Seq)
// 접수번호       :  yymmdd+00+00000000000  (11자리-100억)
//                   (yymmdd + line no + Seq)
//
package matchingboard
import (
	"centolib/api"
	Bd "centolib/bindecimal"
	Dc "github.com/shopspring/decimal"
	"unsafe"
	"os"
	"strconv"
	"strings"
	"slices"
	"bufio"
	"time"
	"fmt"
	"sort"
	Log "centolib/logout"
	"path/filepath"
	"errors"
)
var MAX_PRICE , MIN_PRICE Dc.Decimal

////////////////////////////////////////////////////////////////////////////////
/// matchingboard package's Initialize function
/// @fn     : func init(){
/// @param  :
/// @return :
////////////////////////////////////////////////////////////////////////////////
func init(){
	MAX_PRICE,_ = Dc.NewFromString("9999999999999") // 9조....
	MIN_PRICE,_ = Dc.NewFromString("-1.0")
}


const (
	MAX_ORDTIME  = 9999999999999999999 // 19자리
	MIN_ORDTIME  = 0                   // 
	ORDER_COUNT  = 3                   // map 당 Order수(default)
	WINDOW_COUNT = 2                   // HOGA WINDOW COUNT 
)

type BuySell int
const (  
    MEDO BuySell = 1 + iota
    MESU
)
var buysells=[...]string{ "MEDO", "MESU", }
func (bs BuySell) String() string{return buysells[bs-1]}

// 신규 상장시 사용
func AddNewCoin(fileName, newCoin string) error{
	file, err := os.Open(fileName)
	if err != nil{return err; }
	var lines, coins []string
	var index, coinCount int
	scanner := bufio.NewScanner(file)
	coinCount = 9999999 // 초기값
	lineNum := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "ORD_"){
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2{
				list := strings.Split(strings.Trim(parts[1], " :"), ":")
				cnt := len(list)
				if coinCount > cnt {
					coinCount = cnt
					index = lineNum
				}
				coins = append(coins, list...)
			}
			lines = append(lines, line)
		}else{
			lines = append(lines, line)
		}
		lineNum ++
	}
	file.Close()

	// 이미 있는가?
	if exist := slices.Contains(coins, newCoin); exist{
		return fmt.Errorf("AddNewCoin() %s는 이미 등록된 코인입니다", newCoin)
	}

	// Line 마지막에 #이 있는가 검사
	parts := strings.SplitN(lines[index], "#", 2)
	if len(parts) < 2{
		lines[index] = strings.Trim(lines[index], " :") + ":"+newCoin
	}else{
		lines[index] = strings.Trim(parts[0] , " :") + ":" + newCoin + " #" + parts[1]
		//fmt.Printf("Lines [%s]\n", lines[index])
	}

	os.Rename(fileName, fileName+".BACK")
	err = os.WriteFile(fileName, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil{
		os.Rename(fileName+".BACK", fileName)
		return err
	}

	return nil
}

/// Index별 코인 읽기
func GetCoinListByIndex(fileName string, idx int)([]string, error){

	var coins []string

	file, err := os.Open(fileName)
	if err != nil{return nil, err; }
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan(){
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "ORD_"){
			str := strings.SplitN(line, "#", 2) // Line 중 # 제거
			parts := strings.SplitN(str[0], "=", 2)
			if len(parts) != 2{continue;}
			name := strings.SplitN(strings.Trim(parts[0], " "), "_", 2) // Index 구하기 
			num,_ := strconv.Atoi(name[1])
			if num != idx {continue;}
			lists :=strings.Split( strings.Trim(parts[1], " :"), ":")
			for _, item := range(lists){
				coins = append(coins, strings.Trim(item, " "))
			}
		}
	}

	if len(coins) == 0 {
		return nil, fmt.Errorf("ORD[%d]는 등록된 코인이 없습니다.",idx)
	}
	return coins, nil
}
// 모든 코인읽기
func GetAllCoins(fileName string)([]string, error){
	var coins []string
	file, err := os.Open(fileName)
	if err != nil{
		return nil, fmt.Errorf("GetAllCoins() : %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan(){
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "ORD_"){continue;}
		str := strings.SplitN(line, "#", 2) // Line중 #
		parts := strings.SplitN(str[0], "=", 2) // ORD_??= 분리
		if len(parts) != 2{continue;}
		lists := strings.Split( strings.Trim(parts[1], " :"), ":") // Coin 분리
		for _, item := range(lists){
			coins = append(coins, strings.Trim(item, " "))
		}
	}
	if len(coins) == 0 {
		return nil, fmt.Errorf("등록된 코인이 없습니다.")
	}
	return coins, nil
}

//
// Program이 사용할 Memory 
//
// MEMCOIN --> MEMHOGA 
//STRUCTURE/////// MEMHOGA ////////////////////////////
type MEMHOGA struct {
	Price      Dc.Decimal    // 가격
	SumQty     Dc.Decimal    // 주문수량 합
	OrderCount int           // 총 주문건수 
	IsWindow   bool          // Window안에 있는가?(Map되어 있나?)

	MapName    string        // Default Path + MapName
	MapPoint   []byte        // Memory Mapped file attach pointer
	Hd         *OrderHeader  // Header Pointer
	OrdList    uintptr       // Order List start point
}
////////////////////////////////////////////////////////////////////////////////
///
/// @fn     : func (p *MEMHOGA)GetOrdPointerByIndex(index int)(*Order, error)
/// @param  :
/// @return :
////////////////////////////////////////////////////////////////////////////////
func (p *MEMHOGA)GetOrdPointerByIndex(index int)(*Order, error){
	if p.Hd == nil {
		return nil, fmt.Errorf("%s is unmapped", p.MapName)
	}
	if index > p.Hd.BufCount  {
		return nil, fmt.Errorf("Error Index! BufCount[%d] index[%d]", 
			p.Hd.BufCount, index)
	}
	pord := (*Order)(unsafe.Pointer(p.OrdList + unsafe.Sizeof(Order{}) * uintptr(index)))
	return pord, nil
}
////////////////////////////////////////////////////////////////////////////////
/// HOGA의 Order Buffer를 물리적으로 늘린다(ORDER_COUNT 만큼 증가)
/// 수행전 Map 되어 있어야함
/// @fn      func (p *MEMHOGA)IncOrderBuff()error
/// @param  
/// @return error
////////////////////////////////////////////////////////////////////////////////
func (p *MEMHOGA)IncOrderBuff()error{

	api.UnMapfile(p.MapPoint)

	file, err := os.OpenFile(p.MapName, os.O_RDWR, 0666)
	if err != nil { return err }

	finfor, err := file.Stat()
	if err != nil{ return err }
	fsize := finfor.Size() + int64(unsafe.Sizeof(Order{})) * ORDER_COUNT

	// Size 증가
	err = file.Truncate(int64(fsize))
	if err != nil{ return err }
	defer file.Close()

	p.MapPoint, _, err = api.AttachMapfile(p.MapName)
	if err != nil{ return err }
	// Header Buffer count 증가
	p.Hd = (*OrderHeader)(unsafe.Pointer(&p.MapPoint[0]))
	p.Hd.BufCount = p.Hd.BufCount + ORDER_COUNT
	p.OrdList = uintptr(unsafe.Pointer(&p.MapPoint[0])) + uintptr(unsafe.Sizeof(OrderHeader{}))

Log.Debug("\t기존 Size : %d	증가후 SZ: %d, BufCount : %d OrderCount:%d", 
			finfor.Size(), fsize, p.Hd.BufCount, p.Hd.OrderCount)

	return nil
}
////////////////////////////////////////////////////////////////////////////////
/// Window 안으로
/// @fn     : func (p *MEMHOGA)InWindow() error {
/// @param  :
/// @return : error
////////////////////////////////////////////////////////////////////////////////
func (p *MEMHOGA)InWindow() error {
	if p.IsWindow == true {return nil}
	pMap, _, err := api.AttachMapfile(p.MapName)
	if err != nil { 
		msg := fmt.Sprintf("AttachMapfile(%s) : %v", p.MapName, err)
		return errors.New(msg) 
	}		
	p.MapPoint = pMap
	p.Hd        = (*OrderHeader)(unsafe.Pointer(&pMap[0]))
	p.OrdList   = uintptr(unsafe.Pointer(&p.MapPoint[0])) + uintptr(unsafe.Sizeof(OrderHeader{}))
	p.IsWindow = true
Log.Infor("MEMHOGA.InWindow OrderCount[%d] Price[%s] Name[%s]", 
		p.OrderCount, p.Price.String(), p.MapName)
	return nil
}
////////////////////////////////////////////////////////////////////////////////
/// Window 바깥으로
/// @fn     : func (p *MEMHOGA)OutWindow(){
/// @param  :
/// @return :
////////////////////////////////////////////////////////////////////////////////
func (p *MEMHOGA)OutWindow(){
	if p.IsWindow == false {return}
	api.UnMapfile(p.MapPoint)
	p.MapPoint = nil
	p.Hd       = nil
	p.OrdList  = uintptr(0)
	p.IsWindow = false
Log.Infor("MEMHOGA.OutWindow OrderCount[%d] Price[%s] Name[%s]", 
		p.OrderCount, p.Price.String(), p.MapName)
}

////////////////////////////////////////////////////////////////////////////////
/// 해당 Index에 Order를 copy한다(Local function)
/// 🔴 Hd.BufCount가 충분하지 호출 전 검사할 것.
///    Hd.OrderCount 는 증가하지 않음.
/// @fn     : func (p *MEMHOGA)copyOrderByIndex(idx int, ord *Order)int
/// @param  : idx Copy할 Index(0부터), ord : Order structure pointer
/// @return : 
////////////////////////////////////////////////////////////////////////////////
func (p *MEMHOGA)copyOrderByIndex(idx int, ord *Order){
	offset := unsafe.Sizeof(OrderHeader{})+unsafe.Sizeof(Order{})*uintptr(idx)
	ordBytes := (*[unsafe.Sizeof(Order{})]byte) (unsafe.Pointer(ord))[:]
	copy(p.MapPoint[offset:], ordBytes)	
	return 
}
// MEMHOGA SumQty 증가
// MEMHOGA.SumQty     - Decimal
// MEMHOGA.Hd.SumQty  - BinDecimal
func (p *MEMHOGA)AddSumQty(a interface{})error{
	val, ecode := GetDecimal(a)
	if ecode != 0{
		msg := fmt.Sprintf("GetDecimal(%v) 오류", a)
		return errors.New(msg)
	}
	p.SumQty = p.SumQty.Add(val)
	p.Hd.SumQty.Set(p.SumQty)
	return nil
}
// MEMHOGA SumQty 감소 
// MEMHOGA.SumQty     - Decimal
// MEMHOGA.Hd.SumQty  - BinDecimal
func (p *MEMHOGA)SubSumQty(a interface{})error{
	val, ecode := GetDecimal(a)
	if ecode != 0{
		msg := fmt.Sprintf("GetDecimal(%v) 오류", a)
		return errors.New(msg)
	}
	p.SumQty = p.SumQty.Sub(val)
	p.Hd.SumQty.Set(p.SumQty)
	return nil
}
//
// MEMHOGA에 가격을 설정한다
// MEMHOGA.Price    - Decimal
// MEMHOGA.HD.Price - BinDecimal
func (p *MEMHOGA)SetPrice(prc interface{}){
	// 순서 주의
	p.Hd.Price.Set(prc)            // BinDecimal
	p.Price = p.Hd.Price.Decimal() // Decimal
}
// 주문수 증가
func (p *MEMHOGA)IncOrderCount(){
	p.OrderCount  = p.OrderCount + 1       // Memory
	p.Hd.OrderCount = p.Hd.OrderCount + 1  // Map file
}
// 주문 수 감소
func (p *MEMHOGA)DecOrderCount(){
	p.OrderCount  = p.OrderCount - 1      // Memory
	p.Hd.OrderCount = p.Hd.OrderCount - 1 // Mapfile 

}
////////////////////////////////////////////////////////////////////////////////
/// 같은 호가에 주문 Insert
/// 🔴 Window에 없어도 처리됨.(map -> size증가 -> Insert)
///    p.Hd.OrderCount 증가 시킴.
/// @fn      func (p *MEMHOGA)InsertOrder(ord *Order)error
/// @param  
/// @return
////////////////////////////////////////////////////////////////////////////////
func (p *MEMHOGA)InsertOrder(ord *Order)error{
	var isWin bool

	/// window에 없으면(Map 되어있지 않았으면)
	if p.IsWindow == false {
		err := p.InWindow()
		if err != nil {
			msg := fmt.Sprintf("InsertOrder() ORDSEQ:%d\n\t%v\nJSON:%s",
				ord.OrdNum, err, ord.OrdPacket)
			return errors.New(msg)
		}
//Log.Debug("%s, Attach Ok.", p.MapName)
		isWin = false
	}else{
		isWin = true
	}

	////////// 여기서 부터는 map되어있어 야함 ///////
	// Buffer 확인
	if p.Hd.BufCount < (p.Hd.OrderCount + 1){
		err :=p.IncOrderBuff()
		if err != nil{
			return fmt.Errorf("InertOrder() %v", err)
		}
	}

	// 주문 Copy
	p.copyOrderByIndex(p.Hd.OrderCount, ord)

	// MEMHOGA의 총호가수량, 주문 갯수 증가
	// 촐호가수량은 주문 잔량 (RemQty)를 사용
	// Memory와 Mapfile 동시에
	//   중복 DATA에 대한 생각 SumQty, SumQty
	//   Memory load시 속도에 유리할까 해서 중복으로 둔다
	p.AddSumQty(ord.RemQty) 
	p.IncOrderCount()

/*
Log.Debug("H.SumQty(%v) H.OrderCount(%d) Hd.SumQty(%v) Hd.OrderCount(%d)",
	p.SumQty, p.OrderCount, p.Hd.SumQty.Decimal(), p.Hd.OrderCount)
*/
	/// 원래 Window 밖에있는 것이면...
	if isWin == false { p.OutWindow() }

	return nil
}
// 호가 Mapfile을 Memory에 Load한다
// 호출전에 MapName을 Set해야 함
// isWindow = false
func (p *MEMHOGA)LoadMEMHOGA()error{
	pMap, _, err := api.AttachMapfile(p.MapName)
	if err != nil{return err;}
	pHd := (*OrderHeader)(unsafe.Pointer(&pMap[0]))
	// 가격
	p.Price  = pHd.Price.Decimal().Copy() // BinDecimal -> Decimal
	// 주문 수량 합
	p.SumQty = pHd.SumQty.Decimal().Copy()// BinDecimal -> Decimal
	// 총 주문건수
	p.OrderCount = pHd.OrderCount
	// Window 안에 있나?
	p.IsWindow   = false

	// IsWindow = false 조건
	// -. MapPoint = nil
	// -. Hd = nil
	// -. OrdList = nil
	api.UnMapfile(pMap)
	return nil
}
/// MEMHOGA에 있는 Order를 OrdNum으로 Sorting 한다(오름차순)
/// 참고 : centolib/api/sortst.go
func (p *MEMHOGA)SortOrder(){
	sortslice := api.SortSlice{
		StBase: p.OrdList,
		StCount: int(p.Hd.BufCount),
		StSize:  uintptr(unsafe.Sizeof(Order{})),
		LessFunc: func(a, b unsafe.Pointer)bool{
			o1 := (*Order)(a)
			o2 := (*Order)(b)
			return o1.OrdNum < o2.OrdNum
		},
	}
	sort.Sort(&sortslice)
}
//STRUCTURE//////// CoinInfor //////////////////////    
type CoinInfor struct{
    Symbol      [32]byte        // Coin 명
    ItemCode     int            // Market index
    CoinCode     int            // Coin index
    MkCode       int            // 결제 통화 index
} 

func (p *CoinInfor)GetSymbol()string{
	name := string(p.Symbol[:])
	return strings.TrimRight(name, "\x00")
}

//////////////////////// OrderInfor ////////////////////
// 코인별 주문정보 Master
// Mapfile 정보 Type
type OrderMasterInfor struct{
	Coin         CoinInfor     // Coin정보
    OrdNum       uint64        // 주문일련번호 (19:yyyymmdd+9(11))
    BfPrice      Bd.BinDecimal // 직전 체결가
    BfTime       uint64        // 직전 체결시각
    MedoNameIdx  int           // 매도 Mapfile 이름 index
    MesuNameIdx  int           // 매수 Mapfile 이름 index
}
//STRUCTURE/////// MEMCOIN : 최상위 구조////////////////////
type MEMCOIN struct{
	MstP         []byte
	MstInfor     *OrderMasterInfor
	Path         string     // Default Path( $(ORDDIR)/$(Coinname)/ )

/// Medo, Mesu Array는 주문이 있는 hoga만 유지되도록 한다.
/// 주문이 없는 Hoga는 꼭 Delete해야 함.
	Medo         []MEMHOGA    // 매도 호가 정보
	Mesu         []MEMHOGA    // 매수 호가 정보
}

// 해당 코인의 신규 주문Master를 생성한다
// Mapfile을 만들고 Coin정보 Update
// 주문 번호는 "생성당일 + 0"으로 설정한다
// path : 주문정보 Default path
// coin : Coin 정보(이름, MKIDX, CIDX, MKCIDX)
func CreateOrderMasterInfor(path string, coin CoinInfor)error{
	name := coin.GetSymbol()
	tpath := fmt.Sprintf("%s/%s", strings.TrimSuffix(path, "/"), name)

	_, err := os.Stat(tpath)
	if err != nil {
		msg := fmt.Sprintf("%s는 없습니다.\n\t%v", tpath, err)
		return  errors.New(msg)
	}
	fullpath := tpath + "/" + name + ".MASTER"	
	_, err = os.Stat(fullpath)
	if err == nil{
		msg := fmt.Sprintf("%s는 이미 존재합니다.", fullpath)
		return errors.New(msg)
	}
	mPoint, err := api.CreateMapfile(fullpath, api.SizeOf(OrderMasterInfor{}))
	if err != nil {return err}
	pMaster := (*OrderMasterInfor)(unsafe.Pointer(&mPoint[0]))
	pMaster.Coin = coin

	// CoinInfor Move
	copy(pMaster.Coin.Symbol[:], coin.Symbol[:]) 
	pMaster.Coin.ItemCode   = coin.ItemCode 
	pMaster.Coin.CoinCode    = coin.CoinCode 
	pMaster.Coin.MkCode  = coin.MkCode 

	// 주문번호 설정
	seqstr :=  time.Now().Format("20060102") + "00000000000"
    pMaster.OrdNum, _ = strconv.ParseUint(seqstr, 10, 64)
	api.UnMapfile(mPoint)
	return nil
}

// Coin 주문Master를 Attach하고 Mappoint를 return한다
// Attach file의 coin 이름이 다른 경우 error
// path : 주문정보 full path
// name : Coin 이름
func GetOrderMasterInfor(path, name string)([]byte, error){
	fullpath := path +"/"+ name + "/" + name + ".MASTER"	

	mPoint , _, err := api.AttachMapfile(fullpath)
	if err != nil{return nil, err;}
	mst := (*OrderMasterInfor)(unsafe.Pointer(&mPoint[0]))
	//if name != string(mst.Coin.Symbol[:]) {
	if name != mst.Coin.GetSymbol(){
		msg := fmt.Sprintf("%s와 Master name(%s)가 다름: Mapname(%s)",
			name, 
			string(mst.Coin.Symbol[:]),
			fullpath)
		api.UnMapfile(mPoint)
		return nil, errors.New(msg)
	}
	return mPoint, nil
}
//////////////////////////////////////

func (p *MEMCOIN)GetSymbol()string{
	return string(p.MstInfor.Coin.GetSymbol())
}

////////////////////////////////////////////////////////////////////////////////
/// MEDO 또는 MESU Window를 정리한다
/// Window?
///  Os 내 fd 수가 제한되어 있음에 따라
///  Open되어 있는 Memory mapped file의 개수를 줄이기 위해
///  최상위 호가 기준 WINDOW_COUNT 수 만 attach해서 운영하려는 목적으로...
/// @fn      func (p *MEMCOIN)SetWindow(mode BuySell)
/// @param   mode : MEDO, MESU
/// @return
////////////////////////////////////////////////////////////////////////////////
func (p *MEMCOIN)SetWindow(mode BuySell){
	if mode == MEDO{
		for i := range p.Medo{
			if i < WINDOW_COUNT { 
				p.Medo[i].InWindow()
			}else{
				p.Medo[i].OutWindow()
			}
		}
	}else{
		for i := range p.Mesu{
			if i < WINDOW_COUNT { 
				p.Mesu[i].InWindow()
			}else{
				p.Mesu[i].OutWindow()
			}
		} // for loop
	} // else
}
////////////////////////////////////////////////////////////////////////////////
/// 해당가격의 호가를 지운다(Map file and Memory(array))
/// @fn      (p *MEMCOIN)DeleteHogaByPrice(mode BuySell, prc interface{})error {
/// @param   mode  MEDO or MESU
/// @param   price  가격 
/// @return  error
////////////////////////////////////////////////////////////////////////////////
func (p *MEMCOIN)DeleteHogaByPrice(mode BuySell, prc interface{})error {
	index := p.FindHogaIndex(mode, prc)
	if index == -1 {
		dec,_ := GetDecimal(prc)
		return fmt.Errorf("%s-%v는 %s에 없습니다", 
			p.GetSymbol(), dec, BuySell(mode))
	}		

	if mode == MEDO {
		if p.Medo[index].OrderCount != 0{
			return fmt.Errorf("Fail to Delete. 주문이%d개 있습니다. 가격 %v-%s ", 
				p.Medo[index].OrderCount,
				p.Medo[index].Price, p.Medo[index].MapName)
		}
		// Delete Mapfile
		if p.Medo[index].IsWindow == true{
			p.Medo[index].OutWindow()
		}
		err := os.Remove(p.Medo[index].MapName)
		if err != nil{return err }

		Log.Debug("Delete Mapfile.Price(%v) (%s)", 
			p.Medo[index].Price,
			p.Medo[index].MapName) 

		// Array Delete
		p.Medo = append(p.Medo[:index], p.Medo[index+1:]...)
	}else{
		if p.Mesu[index].OrderCount != 0{
			return fmt.Errorf("Fail to Delete. 주문이%d개 있습니다. 가격 %v-%s ", 
				p.Mesu[index].OrderCount,
				p.Mesu[index].Price, p.Mesu[index].MapName)
		}
		// Delete Mapfile
		if p.Mesu[index].IsWindow == true{
			p.Mesu[index].OutWindow()
		}
		err := os.Remove(p.Mesu[index].MapName)
		if err != nil{ return err }	

		Log.Debug("Delete Mapfile.Price(%v) (%s)", 
			p.Mesu[index].Price,
			p.Mesu[index].MapName) 

		// Array Delete
		p.Mesu = append(p.Mesu[:index], p.Mesu[index+1:]...)
	}
	return nil	
}
////////////////////////////////////////////////////////////////////////////////
/// 주어진 가격으로 HOGA Index를 찾는다
/// @fn     : func (p *MEMCOIN)FindHogaIndex(bs BuySell, prc interface {})int
/// @param  :
/// @return : -1:없음
////////////////////////////////////////////////////////////////////////////////
func (p *MEMCOIN)FindHogaIndex(mode BuySell, prc interface{})int{
	target, ecode := GetDecimal(prc)
	if ecode == -1{return ecode}

	var hoga []MEMHOGA
	if mode == MEDO{
		hoga = p.Medo
	}else{
		hoga = p.Mesu
	}
	for i := range hoga{
		if hoga[i].Price.Equal(target){return i}
	}
	return -1
}
////////////////////////////////////////////////////////////////////////////////
/// 주어진 가격으로 HOGA를 찾는다
/// @fn     : func (p *MEMCOIN)FindHoga(bs BuySell, prc interface {}) *MEMHOGA
/// @param  :
/// @return :
////////////////////////////////////////////////////////////////////////////////
func (p *MEMCOIN)FindHoga(mode BuySell, prc interface {}) *MEMHOGA{

	target, ecode := GetDecimal(prc)
	if ecode == -1 {return nil}
	if mode == MEDO {
		for i,hoga := range p.Medo{
			//if p.Medo[i].Price.Equal(target){return &p.Medo[i]}
			if hoga.Price.Equal(target){return &p.Medo[i]}
		}
	}else{
		for i,hoga := range p.Mesu{
			if hoga.Price.Equal(target){return &p.Mesu[i]}
		}
	}
	return nil
}

// 호가 Array중에 비어있는 호가를 찾는다
// 비어있는 호가 : MEMHOGA.SumQty == 0 and MEMHOGA.OrderCount == 0
//    호가수량(SumQty)이 0이고 주문수(OrderCount)가 0인 호가, 
//    가격은 -1(매수) 9999999999999(매도)
//
func (p *MEMCOIN)FindEmptyHoga(mode BuySell) *MEMHOGA{
	if mode == MEDO {
		for i, hoga := range p.Medo{
			if hoga.SumQty.IsZero() && hoga.OrderCount == 0 { 
				return &p.Medo[i] 
			}
		}
	}else{
		for i, hoga := range p.Mesu{
			if hoga.SumQty.IsZero() && hoga.OrderCount == 0 { 
				return &p.Mesu[i] 
			}	
		}
	}
	return nil	
}
// 주문번호 초기화
func (p *MEMCOIN)InitOrderSeq(){
	tm := time.Now()
	seqstr :=  tm.Format("20060102") + "00000000000"
	p.MstInfor.OrdNum, _ = strconv.ParseUint(seqstr, 10, 64)

	Log.Infor( "[%s]주문번호 초기화 : %d", p.GetSymbol(), p.MstInfor.OrdNum)
}
// 주문번호 채번
func (p *MEMCOIN)GetOrderSeq()uint64{
	p.MstInfor.OrdNum ++;
	return p.MstInfor.OrdNum
}
//
// 신규 호가 Mapfile을 생성한다
// -.호가 가격 
//   매도 : MAX_PRICE
//   매수 : MIN_PRICE
// path : Coin 주문 default path (ex) $(OEDHOME)/BTC.USDT
// name : coin name
// idx  : Mapfile name index
//
func CreateHoga(path ,name string, mode BuySell, idx int )error{

	var mapname string	
	if mode == MEDO{
		mapname = fmt.Sprintf("%s/%s.%04d", path, BuySell(mode), idx)
	}else{
		mapname = fmt.Sprintf("%s/%s.%04d", path, BuySell(mode), idx)
	}
	// 3. Create Hoga Memory mapped file
	totSize := api.SizeOf(OrderHeader{}) + api.SizeOf(Order{}) * ORDER_COUNT
	pMap, err := api.CreateMapfile(mapname, int(totSize))
	if err != nil{ return err }

	// Header
	hd := (*OrderHeader)(unsafe.Pointer(&pMap[0]))
	// Symbol 
	copy(hd.Symbol[:], name)
	// Buy Sell,  Header 가격
	if mode == MEDO { 
		hd.BuySell = 'S';
		hd.SetPrice(MAX_PRICE)
	}else{
		hd.BuySell = 'B';
		hd.SetPrice(MIN_PRICE)
	}		
	// Buffer count
	hd.BufCount = ORDER_COUNT
	// OrderCount
	hd.OrderCount = 0
	// 총주문 수랭
	hd.SetSumQty("0.0")

	ord := uintptr(unsafe.Pointer(&pMap[unsafe.Sizeof(OrderHeader{})] ))
	// List에 있는 주문번호 초기화
	for i := 0; i < hd.BufCount; i++{
		pord := (*Order)( unsafe.Pointer( ord + unsafe.Sizeof(Order{}) * uintptr(i) ))
		pord.OrdNum = ORDSEQ_MAX
	}
	api.UnMapfile(pMap)

	return nil
}
////////////////////////////////////////////////////////////////////////////////
/// Process 수행중에 호가가 없는 경우 호출하는 function
/// @fn      func (p *MEMCOIN)NewHoga(prcstr Bd.BinDecimal)
/// @param   prcstr : 가격 BinDecimal
/// @return
////////////////////////////////////////////////////////////////////////////////
func (p *MEMCOIN)NewHoga(prc Bd.BinDecimal, mode BuySell )(*MEMHOGA, error){
	// 1. Price , IsWindow, MapName
 	// 2. MedoName Index 증가 (다음을 위해)
	var mapname string	
	if mode == MEDO{
		mapname = fmt.Sprintf("%s/%s.%04d", 
			p.Path, BuySell(mode), p.MstInfor.MedoNameIdx)
		p.MstInfor.MedoNameIdx= p.MstInfor.MedoNameIdx + 1
	}else{
		mapname = fmt.Sprintf("%s/%s.%04d", 
			p.Path, BuySell(mode), p.MstInfor.MesuNameIdx)
		p.MstInfor.MesuNameIdx = p.MstInfor.MesuNameIdx + 1
	}
	//price := prc.Decimal()
	hoga := &MEMHOGA{
		IsWindow : true, 
		MapName : mapname,
	}

	// 3. Create Hoga Memory mapped file
	totSize := api.SizeOf(OrderHeader{}) + api.SizeOf(Order{}) * ORDER_COUNT
	pMap, err := api.CreateMapfile(hoga.MapName, int(totSize))
	if err != nil{ 
		msg := fmt.Sprintf("NewHoga(%s-%s) Price(%s)\n\t%v", 
			p.GetSymbol(), BuySell(mode), prc.String())
		return nil, errors.New(msg)
	}		

	// 4. MapPoint
	hoga.MapPoint = pMap 

	// 5. Header 
	header := &OrderHeader{
		BufCount : ORDER_COUNT,
		OrderCount : 0,
	}
	if mode == MEDO { 
		header.BuySell = 'S'
	} else { 
		header.BuySell = 'B'
	}		

	copy(header.Symbol[:], p.MstInfor.Coin.Symbol[:])
	header.SetSumQty("0.0")

	// Mapfile에 Header Copy
	binhd := api.GetBytes(header)
	copy(pMap[:], binhd)

	// 6. Header point
	hoga.Hd       = (*OrderHeader)(unsafe.Pointer(&pMap[0]))
	// 7. Order Start Point
	hoga.OrdList  = uintptr(unsafe.Pointer(&pMap[unsafe.Sizeof(OrderHeader{})]))

	// 8. List에 있는 주문 가격 초기화 & Medo List에 추가
	if mode == MEDO { 
		for i := 0; i < hoga.Hd.BufCount; i++{
			ord, _ := hoga.GetOrdPointerByIndex(i)
			//ord.SetPrice(fmt.Sprintf("%d", MAX_PRICE))
			ord.SetPrice(MAX_PRICE)
		}
	}else{
		for i := 0; i < hoga.Hd.BufCount; i++{
			ord, _ := hoga.GetOrdPointerByIndex(i)
			//ord.SetPrice(fmt.Sprintf("%d", MIN_PRICE))
			ord.SetPrice(MIN_PRICE)
		}
	}

	hoga.SetPrice(prc) // Memory 와 Hd(Mapfile)
	hoga.OutWindow() 
/*
Log.Infor("NewHoga  입력가격(%s), 정렬 후 Array Pointer 가격 (%v)",
	 prcstr, phoga.Price)
*/
	return hoga, nil
}
////////////////////////////////////////////////////////////////////////////////
/// MEMCOIN에 있는 HOGA를 sroting 한다(Local Function)
/// 매도 : 오름차순, 매수 : 내림차순
/// @fn      func (p *MEMCOIN)SortHoga(mode BuySell )
/// @param   mode : MEDO or MESU
/// @return
////////////////////////////////////////////////////////////////////////////////
func (p *MEMCOIN)SortHoga(mode BuySell ){
	if mode == MEDO {
		/// 매도 : 오름차순
		sort.Slice(p.Medo, func(i, j int) bool {
			rtn := p.Medo[i].Price.Compare(p.Medo[j].Price)
			if rtn == -1{ return true} // i < j
			return false
		})
	}else{
		/// 매수 : 내림차순
		sort.Slice(p.Mesu, func(i, j int) bool{
			rtn := p.Mesu[i].Price.Compare(p.Mesu[j].Price)
			if rtn == 1{ return true} // i > j
			return false
		})

	}
}
// 매도/매수 호가를 Loading 한다
// 호가 Map file 모두(MEDO.* or  MESU.*)를 open해서
// 신규 MEMHOGA를 만든 후 p.Mesu[] or p.Medo[]에 append()
// Loading 후 오름차순 sorting (MEDO)
// Loading 후 내림차순 sorting (MESU)
func (p *MEMCOIN)loadAllHoga(mode BuySell) error{
	dir, err := os.Open(p.Path) // $(ORDPATH)/COIN
	if err != nil{return err}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil{return err}

	var pattern string
	var hogaList *[]MEMHOGA

	if mode == MEDO{
		pattern = "MEDO.*"
		hogaList = &p.Medo
	}else{
		pattern = "MESU.*"
		hogaList = &p.Mesu
	}
	for _,file := range files{
		if file.IsDir(){continue}
		match, err := filepath.Match(pattern, file.Name())
		if err != nil{return err}
		if match {
			// make new MEMHOGA
			hoga := MEMHOGA{}
			hoga.MapName = p.Path + "/" + file.Name()
			err := hoga.LoadMEMHOGA()
			if err != nil{return err}
			*hogaList = append(*hogaList, hoga)
		}
	}
	// sorting Hoga
	p.SortHoga(mode)
	return nil
}
//
// Window 정리(사용 안함-25.05.14)
// 
func (p *MEMCOIN)WindowProcessing(mode BuySell) error{
	var hogaList *[]MEMHOGA
	if mode == MEDO {
		hogaList = &p.Medo
	} else {
		hogaList = &p.Mesu
	}
	for i := range *hogaList{
		if i < WINDOW_COUNT {
			err := (*hogaList)[i].InWindow()
			if err != nil {return err}
		}else{
			(*hogaList)[i].OutWindow()
		}
	}
	return nil
}
// Coin의 주문/체결을 위한 
// 주문 Memory 구조를 만든다
// 최상위 function
func (p *MEMCOIN)LoadMEMCOIN(path, name string)error{
	// 0. Path가 있나?
	p.Path = path + "/" + name 
	if _, err := os.Stat(p.Path); os.IsNotExist(err) {
		msg := fmt.Sprintf("%s는 없습니다.", p.Path)
		return  errors.New(msg)
	}

	// 1. 주문 Master Information
	bytes, err := GetOrderMasterInfor(path, name)
	if err != nil{return  err;}
	p.MstP = bytes
	p.MstInfor = (*OrderMasterInfor)(unsafe.Pointer(&bytes[0]))

	// 2. MEDO 주문 Load
	err = p.loadAllHoga(MEDO)	
	if err != nil{return  err;}
	// 3. MESU 주문 Load
	err = p.loadAllHoga(MESU)	
	if err != nil{return  err;}
	
	// 4. Window 정리
	p.SetWindow(MEDO)
	p.SetWindow(MESU)

	return nil	
}
func (p *MEMCOIN)UnLoadMEMCOIN(){
	if len(p.MstP) != 0{
		api.UnMapfile(p.MstP)
	}
	for _, h := range(p.Medo) {
		if h.IsWindow == true {
			h.OutWindow()
		}
	}
	for _, h := range(p.Mesu){
		if h.IsWindow == true{
			h.OutWindow()
		}
	}
}
/*************
func NewMEMCOIN(path, name string)(* MEMCOIN, error){
	// 1. Coin
	memcoin := &MEMCOIN{
		Coin: GetCoinInfor(name),
	}
	// 2. OrdNum
	memcoin.InitOrderSeq()

	// 3. BfPrice, BfTime
	memcoin.BfPrice,_ = Dc.NewFromString("0.0")
	memcoin.BfTime  = 0

	// 4.MedoNameIdx, MesuNameIdx
	memcoin.MedoNameIdx = 0
	memcoin.MesuNameIdx = 0

	// 5. Path
	coinpath := path + "/" + name + "/"
	if _, err := os.Stat(coinpath); os.IsNotExist(err) {
		err := os.MkdirAll(coinpath, 0766)	
		if err != nil {
			return nil, fmt.Errorf("fail to create dir : %w", err)
		}
	}else{
		return nil, fmt.Errorf("Directory already exists : %s", coinpath)
	}
	memcoin.Path = coinpath

	Log.Infor( "MEMCOIN 생성. Path : %s", memcoin.Path)
	return memcoin, nil
}


func SortOrderAscByOrdNum(mappoint unsafe.Pointer){
	pHd := (*OrderHeader)(mappoint)
	pOrd := uintptr(mappoint) + uintptr(unsafe.Sizeof(OrderHeader{}))
	sortslice := api.SortSlice{
		StBase: pOrd,
		StCount: int(pHd.BufCount),
		StSize:  uintptr(unsafe.Sizeof(Order{})),
		LessFunc: func(a, b unsafe.Pointer)bool{
			o1 := (*Order)(a)
			o2 := (*Order)(b)
			return o1.OrdNum < o2.OrdNum
		},
	}
	sort.Sort(&sortslice)
}
*****************/
///////////////////////// Dump Function ////////////////////////
func (p *MEMCOIN)Dump(){
	if p == nil {return}
	fmt.Println("=========DUMP MEMCOIN=========")
	fmt.Printf("\tPath : %s\n\n", p.Path)
	p.MstInfor.Dump(1)
	p.DumpHoga()
}
//
func (p *MEMCOIN)DumpHoga() {
	fmt.Printf("\t%s-MEMHOGA List Dump)====\n", 
			string(p.MstInfor.Coin.Symbol[:]))
	// 매도 
	fmt.Println("\t🔵===매도=======")
	dolen := len(p.Medo)
	for i:= dolen -1; i >= 0 ; i-- {
		hoga := p.Medo[i]
	fmt.Printf("\t%3d(%13v) SumQty[%v] 주문수[%d] ISWINDOW[%5v]\n\t\tMapName[%s]\n\t\tMapPoint[%14p],Hd[%14p]\n\t\tOrdList[%v]\n", 
		i, hoga.Price, hoga.SumQty, hoga.OrderCount, hoga.IsWindow, 
		hoga.MapName, hoga.MapPoint, hoga.Hd, hoga.OrdList)
	}
	fmt.Println("\t🔴===매수=======")
	for i, hoga := range p.Mesu {
	fmt.Printf("\t%3d(%13v) SumQty[%v] 주문수[%d] ISWINDOW[%5v]\n\t\tMapName[%s]\n\t\tMapPoint[%14p],Hd[%14p]\n\t\tOrdList[%v]\n", 
		i, hoga.Price, hoga.SumQty, hoga.OrderCount, hoga.IsWindow, 
		hoga.MapName, hoga.MapPoint, hoga.Hd, hoga.OrdList)
	}
}
//
func (p *MEMHOGA)Dump(){
	if p == nil {return}
	fmt.Printf("HOGA : %v, SumQty[%v] 주문수[%d] ISWINDOW[%v] MapName[%s] MapPoint[%p], Hd[%p] OrdList[%v]\n", 
	p.Price, p.SumQty, p.OrderCount, p.IsWindow, p.MapName, p.MapPoint, p.Hd, p.OrdList)
}
//
func (p *CoinInfor)Dump(dep int){
	if p == nil {return }
	var tabstr string
	for i:= 0 ; i < dep; i++{tabstr = tabstr + "\t"}
	fmt.Printf("%sCOIN INFOR : SYMBOL[%s], ItemCode[%d] CoinCode[%d] MKCODE[%d]\n",
		tabstr, p.GetSymbol(), p.ItemCode, p.CoinCode, p.MkCode)
}
//
func (p *OrderMasterInfor)Dump(dep int){
	if p == nil{return}

	var tabstr string
	for i:= 0 ; i < dep; i++{tabstr = tabstr + "\t"}

	fmt.Printf("%sDump OrderMasterInfor\n", tabstr)
	p.Coin.Dump(1)
	fmt.Printf("%sOrdNum(%d) ", tabstr, p.OrdNum)
	fmt.Printf("BfPrice(%v) ", p.BfPrice.Decimal())
	fmt.Printf("Bftime(%d)\n ", p.BfTime)
	fmt.Printf("%sMedoNameIdx(%d) MesuNameIdx(%d)\n", tabstr,
		p.MedoNameIdx, p.MesuNameIdx)
	fmt.Printf("%s------------- OrderMasterInfor\n\n",tabstr)	
}
