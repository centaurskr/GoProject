package main
import (
	"fmt"
)

func main(){
	var flag int
	const (
		SORT_HOGA = 0x01
		SORT_ORDER = 0x10
	)
	flag |= SORT_HOGA
	flag |= SORT_ORDER
	if flag&SORT_HOGA != 0{
		fmt.Println("SORT_HOGA")
	}
	flag &^= SORT_HOGA
	if flag&SORT_HOGA == 0{
		fmt.Println("Clear SORT_HOGA")
	}
}
