package main
import (
	"centolib/api"
	"unsafe"
	"fmt"
	"os"
	Mb "centolib/matchingboard"
)

func main(){
	mapFile, err := api.GetPlatformEnv("ORDERCOND_MAP")
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}

	mapPoint,_, err := api.AttachMapfile(mapFile)
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}
	header   := (*Mb.MapHeader)(unsafe.Pointer(&mapPoint[0]))
	fmt.Printf("======= 주문 상태 Dump [%s] =========\n", mapFile)
	fmt.Printf("생성일  : %s\nBuffer Count[%d] Data Count[%d]\n", 
		string(header.Create[:]), header.Buffer, header.Count)
	fmt.Printf("───────────────────────────────────────────────────────────────\n")
	fmt.Printf("IDX\tCOIN\t\tGROUP\tSTATUS\n")

	coinList := uintptr(unsafe.Pointer(&mapPoint[unsafe.Sizeof(Mb.MapHeader{})]))
	for i := 0; i < header.Count; i++{
		list := (*Mb.OrderCondition)(unsafe.Pointer(coinList + unsafe.Sizeof(Mb.OrderCondition{}) * uintptr(i)))
		fmt.Printf("%02d    %-10.10s            %d    %d\n", 
			i, string(list.Symbol[:]), list.Group, list.Status)
	}
	fmt.Printf("───────────────────────────────────────────────────────────────\n")
	api.UnMapfile(mapPoint)
}
