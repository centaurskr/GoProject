///
/// @brief  : slice test
/// @file   : slice.go
/// @date   : 2024. 04. 22. (월) 14:09:35 KST
/// @author : Cento
///
package main
import (
	"fmt"
	"reflect"
	"unsafe"
)

func printSlice(s []int){
	fmt.Printf("len = %d cap = %d %v\n", len(s), cap(s), s)
}
func main(){
	s := [] int {1,2,3,4,5}
	printSlice(s)	
	for idx,v := range s{
		fmt.Printf("Slice Index [%d]- %v\n", idx, v)
	}

	s = s[:0]
	printSlice(s)	

	fmt.Println("\tbefor :4", s)	
	s = s[:4]
	fmt.Println("\tafter :4", s)	
	printSlice(s)	

	fmt.Println("\tbefore 2:", s)	
	s = s[2:]
	fmt.Println("\tafter 2:", s)	
	printSlice(s)	

	slice1 := []int{1, 2, 3, 4, 5}
	slice2 := slice1
	fmt.Println(slice2)
	slice1[0] = 100
	fmt.Println(slice2)
	slHeader1 := (*reflect.SliceHeader)(unsafe.Pointer(&slice1))
	slHeader2 := (*reflect.SliceHeader)(unsafe.Pointer(&slice2))
	fmt.Println()
	fmt.Println("*-*-*-*-* slice1 *-*-*-*-*")
	fmt.Printf("slHeader1.Data : %v\n", slHeader1.Data)
	fmt.Printf("slHeader1.Len : %v\n", slHeader1.Len)
	fmt.Printf("slHeader1.Cap : %v\n", slHeader1.Cap)
	fmt.Println()
	fmt.Println("*-*-*-*-* slice2 *-*-*-*-*")
	fmt.Printf("slHeader2.Data : %v\n", slHeader2.Data)
	fmt.Printf("slHeader2.Len : %v\n", slHeader2.Len)
	fmt.Printf("slHeader2.Cap : %v\n", slHeader2.Cap)

	slice := []string{"apple", "banana", "orange"}
	fmt.Printf("ORG : %v\n", slice)
	slice = slice[1:]
	fmt.Printf("[1:] : %v\n", slice)
}


