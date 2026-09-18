///
/// @brief   centolib/api를 이용한 Memory mapped file test
/// @file    rdmaplib.go
/// @date    2024. 10. 11. (금) 10:21:06 KST
/// @author  Cento
///
package main

import (
	"fmt"
	"centolib/api"
	"unsafe"
	"os"
)

// Order struct definition
type Order struct {
    OrdSeq  uint64
    OrdType int32
    Midx    int32
    BuySell int32
    OrdTime int64
    RegNum  uint64
    CheSeq  int32
    Price   [32]byte
}

func main(){
	if len(os.Args) != 2 {
		fmt.Println("centolib/api를 이용한 mapfile reader")
		fmt.Println("Map file structure")
		fmt.Println("  type Order struct {")
    	fmt.Println("        OrdSeq  uint64")
    	fmt.Println("        OrdType int32")
    	fmt.Println("        Midx    int32")
    	fmt.Println("        BuySell int32")
    	fmt.Println("        OrdTime int64")
    	fmt.Println("        RegNum  uint64")
    	fmt.Println("        CheSeq  int32")
    	fmt.Println("        Price   [32]byte")
		fmt.Println("}")
		fmt.Printf("Using) %s <Mapfilename>\n\n", os.Args[0])
		return
	}
	mappoint, fsize, err := api.AttachMapfile(os.Args[1])
	if err != nil{
		fmt.Println(err)
		return
	}
	fmt.Printf("%s Attach ok! size[%d]\n", os.Args[1], fsize)
	defer api.UnMapfile(mappoint)

	ordsize   := int64(unsafe.Sizeof(Order{}))
	numorders := int64(fsize) / ordsize

	fmt.Println("Numbers of Order is ", numorders)
	for i:= int64(0); i < numorders; i++{
		offset := i*ordsize
		order := (*Order)(unsafe.Pointer(&mappoint[offset]))
        fmt.Printf("index %d: OrdSeq = %d, OrdType = %d, Midx = %d, BuySell = %d, OrdTime = %d, RegNum = %d, CheSeq = %d, Price = %s\n",
            i, order.OrdSeq, order.OrdType, order.Midx, order.BuySell, order.OrdTime, order.RegNum, order.CheSeq, order.Price)

	}
	//Update Test
	order := (*Order)(unsafe.Pointer(&mappoint[0]))
	order.OrdSeq = 100
}
