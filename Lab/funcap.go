///
/// @brief  : Golang function capture test
/// @file   : funcap.go
/// @date   : 2024. 04. 29. (월) 11:04:30 KST
/// @author : Cento
///
package main
import "fmt"

/// instance "i"를 직접 참조하기 때문에 "3"을 출력한다.
func CaptureLoop1(){
	f:= make([]func(), 3)
	fmt.Println("CaptureLoop1")

	for i:= 0; i < 3; i++{
		f[i] = func(){
			fmt.Println(i)
		}
	}
	for i:= 0 ; i<3;i++{
		f[i]()
	}
}
/// instance "v"를 참조
func CaptureLoop2(){
	f:= make([]func(), 3)
	fmt.Println("CaptureLoop2")

	for i:= 0; i < 3; i++{
		v := i
		f[i] = func(){
			fmt.Println(v)
		}
	}
	for i:= 0; i < 3; i++{
		f[i]()
	}
}

func main(){
	CaptureLoop1()
	CaptureLoop2()
}

