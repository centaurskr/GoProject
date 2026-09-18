///
/// @brief  : GoLang map test-2
/// @file   : gomap2.go
/// @date   : 2024. 04. 29. (월) 10:26:45 KST
/// @author : Cento
///
package main
import "fmt"

type Product struct{
	Name  string
	Price int
}
func main(){
	m:= make(map[int]Product)

	m[19] = Product{Name:"TV", Price : 3000}
	m[18] = Product{Name:"Phone", Price : 1200}
	m[17] = Product{Name:"PC", Price : 500}
	m[16] = Product{Name:"Table", Price : 2000}

	for key, value := range m{
		fmt.Println(key, value)
	}	
}

