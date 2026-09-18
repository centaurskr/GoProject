package main
import (
	"fmt"
	"strings"
)

func main(){
	var strList []string
	str1 := "  :TEST1:TEST2:TEST3"
	str2 := ":ARG1:ARG2:ARG3"
	//strList = strings.Split(strings.TrimSuffix(str1, ":"), ":")
	strList = strings.Split(strings.Trim(str1, " :"), ":")
	strTmp := strings.Split(strings.Trim(str2, " :"), ":")

	fmt.Printf("strList : %v, len(%d)\n", strList, len(strList))
	fmt.Printf("strTmp  : %v, len(%d)\n", strTmp, len(strTmp))
	strList = append(strList, strTmp...)
	fmt.Printf("strList : %v, len(%d)\n", strList, len(strList))
}
