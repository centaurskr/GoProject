// 주문 Map file을 dump한다
//
package main
import (
	"fmt"
	"centolib/api"
	Mb "centolib/matchingboard"
	"os"
	"unsafe"
)
func getOrdByIndex(base []byte, idx int)(*Mb.Order, int){
	hd := (*Mb.OrderHeader)(unsafe.Pointer(&base[0]))
	if hd.BufCount <= idx {return nil, 0}
	pord :=(*Mb.Order)(unsafe.Pointer(&base[unsafe.Sizeof(Mb.OrderHeader{}) + unsafe.Sizeof(Mb.Order{}) * uintptr(idx)]))
	return pord, 1
}
func getOrdPointerByIndex(base []byte, idx int)(uintptr, int){
	hd := (*Mb.OrderHeader)(unsafe.Pointer(&base[0]))
	if hd.BufCount <= idx {return uintptr(0), 0}
	ptr :=uintptr(unsafe.Pointer(&base[unsafe.Sizeof(Mb.OrderHeader{}) + unsafe.Sizeof(Mb.Order{}) * uintptr(idx)]))
	return ptr, 1
}
func main(){
	if len(os.Args) != 3 {
		fmt.Println("Order Mapfile Dump")
		fmt.Println("goDumpOrdmap <mapname> <mode>")
		fmt.Println(" <mode> 1: Dump Full, 2:JSON은 Dump 하지않음")
		return 
	}
	mode := os.Args[2]

	pMap,_, err := api.AttachMapfile(os.Args[1])
	if err != nil { fmt.Println(err); return }
	defer api.UnMapfile(pMap)

	pHd := (*Mb.OrderHeader)(unsafe.Pointer(&pMap[0]))

	szHd := api.SizeOf(Mb.OrderHeader{})
    szOrd := api.SizeOf(Mb.Order{})

	fmt.Printf("=== Order Map file Dump : %s, SZHd (%d) szOrd(%d) mode(%s)\n", 
		os.Args[1], szHd, szOrd, mode)

	for i := 0; i <pHd.BufCount; i++{
		ptr, ecode:=getOrdPointerByIndex(pMap, i)
		if ecode != 1{break}
		pord := (*Mb.Order )(unsafe.Pointer(ptr))
		pord.Dump(i)
	}

	pHd.Dump()

	//Log.Infor("가격 MAX(%v) MIN(%v)", Mb.MAX_PRICE, Mb.MIN_PRICE)
}
