///
/// @brief  : GOLANG greeting package test
/// @file   : greetingmain.go
/// @date   : 2024. 04. 29. (월) 14:37:36 KST
/// @author : Cento
///
package main

import (
	"fmt"
	"github.com/dev-yakuza/study-golang/package/greeting"
)

func main(){
	fmt.Println(greeting.PI)
	greeting.PrintGreeting()
}

