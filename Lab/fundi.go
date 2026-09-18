///
/// @brief  : GOLANG DI(Dependency Injection) Test 
/// @file   : fundi.go
/// @date   : 2024. 04. 29. (월) 11:56:09 KST
/// @author : Cento
///
package main
import "fmt"

type PrintConsole func(string)

func PrintMessage(p PrintConsole){
	p("Hello, World!")
}

func main(){
	PrintMessage(func(msg string){
		fmt.Println(msg)
	})
}

