///
/// @brief  : GOLANG local module test
/// @file   : greetinglocal.go
/// @date   : 2024. 04. 29. (월) 16:00:50 KST
/// @author : Cento
///
package main

import (
	gt "module/greeting"
	"fmt"
)

func main(){
	fmt.Println(gt.PI)
	gt.PrintGreeting()
}

