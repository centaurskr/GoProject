//주문 상태를 Check하는 Function
//
package main
import (
	"centolib/api"
	Mb "centolib/matchingboard"
	sc "centolib/screen"
	"unsafe"
)
func displayOrderCondition()bool{
	hLine := "────────────────────────────────────────────"
	clearArea(1, AppTop, MaxX - 1, AppBottom - AppTop)
	mapFile, err := api.GetPlatformEnv("ORDERCOND_MAP")
	if err != nil{
		messageLine("%v", err)
		return true
	}
	mapPoint, _, err := api.AttachMapfile(mapFile)
	if err != nil{
		messageLine("%v", err)
		return true
	}
	defer api.UnMapfile(mapPoint)

	header := (*Mb.MapHeader)(unsafe.Pointer(&mapPoint[0]))
	x :=10
	i := 0
	outStrXy(true,x, AppTop + i, "%s", "================ 주문상태 ==================");i++;
	outStrXy(true,x, AppTop + i, "Map File [%s]", mapFile); i++;
	outStrXy(true,x, AppTop + i, "생성일 : %s", string(header.Create[:])); i++;
	outStrXy(true,x, AppTop + i, "Buffer Count : %d, Data Count : %d",
					header.Buffer, header.Count); i++;
	outStrXy(true,x, AppTop + i, "%s", hLine); i++
	outStrXy(true,x, AppTop + i, "%s", "순번    COIN     GROUP    상태");i++;
	cList := uintptr(unsafe.Pointer(&mapPoint[unsafe.Sizeof(Mb.MapHeader{})]))
	for ii:= 0; ii < header.Count; ii++{
		list := (*Mb.OrderCondition)(unsafe.Pointer(cList + unsafe.Sizeof(Mb.OrderCondition{}) * uintptr(ii)))
		outStrXy(true,x, AppTop + i, "%2d    %20.20s       %d    %d", ii,
				string(list.Symbol[:]), 
				list.Group,
				list.Status);i++;
	}
	outStrXy(true,x, AppTop + i, "%s", hLine)
	return true
}
// OrderCondition은 주문 조건 화면을 표시하고 키 입력을 처리한다.
func OrderCondition()bool{
	displayOrderCondition()
	run := true
	for run {
		char , _:=sc.GetKey()
		switch {
			case char == "q" || char == "Q" :
				return true
			case char == sc.CTR_L :
				displayOrderCondition()
		}
	}
	return true
}

