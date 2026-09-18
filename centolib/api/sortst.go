///
/// @brief   Structure Sort
/// @file    sortst.go
/// @date    2024. 10. 16. (수) 12:48:28 KST
/// @author  Cento
///
package api

import (
	"unsafe"
	"sort"
)

// Sort를 위한 slice
type SortSlice struct {
	StBase  uintptr      // Base point of structure list
	StCount int          // Count of structures
	StSize  uintptr      // Size of structure
	LessFunc func(unsafe.Pointer, unsafe.Pointer) bool // Compare function
}

type SearchMode int
const (
	SMODE_EQ SearchMode = iota // Equal
	SMODE_GT
	SMODE_LT
	SMODE_GE
	SMODE_LE
)
////////////////////////////////////////////////////////////////////////////////
/// 
/// @fn     : func (ss SortSlice)Len() int{
/// @param  :
/// @return :
////////////////////////////////////////////////////////////////////////////////
func (ss *SortSlice)Len() int{
	return ss.StCount
}
////////////////////////////////////////////////////////////////////////////////
///
/// @fn     : func (ss SortSlice)Less(i, j int) bool{
/// @param  :
/// @return :
////////////////////////////////////////////////////////////////////////////////
func (ss *SortSlice)Less(i, j int) bool{
	elem1 := unsafe.Pointer(ss.StBase + uintptr(i)*ss.StSize)
	elem2 := unsafe.Pointer(ss.StBase + uintptr(j)*ss.StSize)
	return ss.LessFunc(elem1, elem2)
}
////////////////////////////////////////////////////////////////////////////////
///
/// @fn     : func (ss SortSlice)Swap(i, j int){
/// @param  :
/// @return :
////////////////////////////////////////////////////////////////////////////////
func (ss *SortSlice)Swap(i, j int){
    // Get pointers to the elements
    elem1Ptr := unsafe.Pointer(ss.StBase + uintptr(i)*ss.StSize)
    elem2Ptr := unsafe.Pointer(ss.StBase + uintptr(j)*ss.StSize)

    // Allocate temporary storage to hold one element for the swap
    temp := make([]byte, ss.StSize)

    // Perform the swap in memory
    copy(temp, (*[1<<30]byte)(elem1Ptr)[:ss.StSize]) // Copy elem1 to temp
    copy((*[1<<30]byte)(elem1Ptr)[:ss.StSize], (*[1<<30]byte)(elem2Ptr)[:ss.StSize]) // Copy elem2 to elem1
    copy((*[1<<30]byte)(elem2Ptr)[:ss.StSize], temp) // Copy temp to elem2
}
////// getElement ///////////////////////////////////
func (ss *SortSlice)getElement(i int)unsafe.Pointer{
	return unsafe.Pointer(ss.StBase + uintptr(i)*ss.StSize)
}
////// Search ///////////////////////////////////
func (ss *SortSlice)Search(key unsafe.Pointer) unsafe.Pointer{
	idx := sort.Search(ss.StCount, func (i int)bool{
		return  !ss.LessFunc(ss.getElement(i), key)
	})
	if idx < ss.StCount && !ss.LessFunc(key, ss.getElement(idx)){
		return ss.getElement(idx)
	}
	return nil // Not found
}
func (ss *SortSlice)SearchMode(key unsafe.Pointer, mode SearchMode)unsafe.Pointer{
	idx := sort.Search(ss.StCount, func(i int) bool {
		return !ss.LessFunc(ss.getElement(i), key)
	})

	switch mode {
	case SMODE_EQ:
		if idx < ss.StCount && !ss.LessFunc(key, ss.getElement(idx)) {
			return ss.getElement(idx)
		}
	case SMODE_GT:
		if idx < ss.StCount && !ss.LessFunc(key, ss.getElement(idx)) {
			idx++
		}
		if idx < ss.StCount {
			return ss.getElement(idx)
		}
	case SMODE_GE:
		if idx < ss.StCount && !ss.LessFunc(key, ss.getElement(idx)) {
			return ss.getElement(idx) // Exact match found
		} else if idx < ss.StCount {
			return ss.getElement(idx) // Next greater element
		}
	case SMODE_LT:
		if idx > 0 {
			return ss.getElement(idx - 1)
		}
	case SMODE_LE:
		if idx < ss.StCount && !ss.LessFunc(key, ss.getElement(idx)) {
			return ss.getElement(idx) // Exact match
		} else if idx > 0 {
			return ss.getElement(idx - 1) // Closest lesser value
		}
	}
	return nil // Not found
}

//func (ss *SortSlice) Search(matchFunc func(unsafe.Pointer) bool) (int, unsafe.Pointer){

