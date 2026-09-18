///
/// @brief  : Golang 가변인수(Variable argument list) test
/// @file   : varg.go
/// @date   : 2024. 04. 29. (월) 10:39:19 KST
/// @author : Cento
///

package main
import "fmt"

func sum(nums ...int) int {
	fmt.Printf("Nums type:%T\n", nums)

	sum := 0
	for _,num:=range nums {
		sum += num
	}
	return sum
}

func main(){
	fmt.Println(sum(1,2,3,4,5))
	fmt.Println(sum(10, 20))
	fmt.Println(sum())
}
