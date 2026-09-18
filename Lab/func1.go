///
/// @brief  : GOLANG function Type
/// @file   : func1.go
/// @date   : 2024. 04. 29. (월) 10:52:01 KST
/// @author : Cento
///
package main

import "fmt"

type OperatorFunc func(int, int) int

func add(a, b int) int {
	return a + b
}

func multiple(a, b int) int{
	return a * b
}

func getOperator(op string) OperatorFunc {
	if op == "+"{
		//return add
		return func(a, b int) int{
			return a + b
		}
	}else if op == "*" {
		// return multiple
		return func(a, b int) int{
			return a * b
		}
	}else{
		return nil
	}
}

func main(){
	//var op func(int, int) int
	var op OperatorFunc

	op = getOperator("+")
	result := op(10, 20)
	fmt.Println(result)
	
	op = getOperator("*")
	result = op(10, 20)
	fmt.Println(result)
}
