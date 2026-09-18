///
/// @brief  : Test Struct array sort
/// @file   : sortarray.go
/// @date   : 2024. 10. 14. (월) 09:44:56 KST
/// @author : Cento
///
package main

import (
	"fmt"
	"reflect"
	"sort"
	"unsafe"
)

type Order struct {
	OrdSeq int32
	Index  int32
	Name   string 
}
///////////////////////////////////// Order Sort를 위한 struct
type SortSlice struct {
    base    uintptr     // 구조체의 Base point
    count   int         // 갯수
    size    uintptr     // 구조체 size
}
// Len returns the number of element
func (ss SortSlice)Len() int{
    return ss.count
}
// Less compares the OrdSeq of two orders and determines if the first is less than the second
func (ss SortSlice)Less(i, j int)bool{
    ord1 := (*Order)(unsafe.Pointer(ss.base + uintptr(i)*ss.size))
    ord2 := (*Order)(unsafe.Pointer(ss.base + uintptr(j)*ss.size))
    return ord1.OrdSeq < ord2.OrdSeq
}

// Swap swaps the memory locations of two orders in the memory-mapped file
func (ss SortSlice) Swap(i, j int) {
    // Get pointers to the orders
    ord1Ptr := (*Order)(unsafe.Pointer(ss.base + uintptr(i)*ss.size))
    ord2Ptr := (*Order)(unsafe.Pointer(ss.base + uintptr(j)*ss.size))

    // Temporary variable to hold one order while swapping
    var temp Order

    // Perform the swap in memory
    temp = *ord1Ptr
    *ord1Ptr = *ord2Ptr
    *ord2Ptr = temp
}
/////////////////////////////////

func main(){
	/// Make aray
	point := []Order{
		{OrdSeq : 1, Index :  1, Name : "1" },
		{OrdSeq : 5, Index :  5, Name : "5" },
		{OrdSeq : 18,Index : 18, Name : "18"},
		{OrdSeq : 11,Index : 11, Name : "11"},
		{OrdSeq : 9, Index :  9, Name : "9" },
		{OrdSeq : 3, Index :  3, Name : "3" },
		{OrdSeq : 7, Index :  7, Name : "7" },
	}
	pp := &point
	fmt.Println(point, reflect.TypeOf(point))
	fmt.Println(pp, reflect.TypeOf(pp))
	fmt.Println(*pp)

	// sort-1
	// Test sort struct
	// i < j --> 오름차순, i > j --> 내림차순
	sort.Slice(point, 
		func(i, j int)bool{
			return point[i].OrdSeq < point[j].OrdSeq // 오름차순
		})			

	fmt.Println(point)	

	// Test Sort struct pointer
	sort.Slice(*pp, 
		func(i, j int)bool{
			return (*pp)[i].OrdSeq > (*pp)[j].OrdSeq // 내림차순
		})			
	fmt.Println(point)	
	
	// Test Search by OrdSeq
	// sorting 된 배열이 오름차순, 내림차순에 따라 부등호가 바뀜
	targetSeq := int32(11)
	idx := sort.Search(len(point), func(i int) bool{
		return point[i].OrdSeq <= targetSeq // 내림차순일 때 Search 부등호 방향
	})
	
	if idx < len(point) && point[idx].OrdSeq == targetSeq {
		fmt.Printf("Found Order at index %d : %+v\n", idx, point[idx])
	}else{
		fmt.Printf("Order not found!idx:%d, OrdSeq:%d\n", idx, targetSeq)
	}

	// sort-2
	/// Make aray
	point1 := []Order{
		{OrdSeq : 1, Index :  1, Name : "1" },
		{OrdSeq : 5, Index :  5, Name : "5" },
		{OrdSeq : 18,Index : 18, Name : "18"},
		{OrdSeq : 11,Index : 11, Name : "11"},
		{OrdSeq : 9, Index :  9, Name : "9" },
		{OrdSeq : 3, Index :  3, Name : "3" },
		{OrdSeq : 7, Index :  7, Name : "7" },
	}
	fmt.Println("Before Sorting-2 : ", point1)
	sortslice := SortSlice{
		base : uintptr(unsafe.Pointer(&point1[0])),
		count : len(point1),
		size  : uintptr(unsafe.Sizeof(Order{})),
	}
	sort.Sort(sortslice)
	fmt.Println("After Sorting-2  : ", point1)
}
