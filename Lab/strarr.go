package main
import (
	"fmt"
)

func main(){
	var strarr []string

	for i := 0; i < 10; i++ {
		str := fmt.Sprintf("{\"INDEX\":%d, \"NAME\":\"TEST%d\"}", i ,i)
		strarr = append(strarr, str)
	}
	for i, s:= range strarr{
		fmt.Println(i, s)
	}
}
