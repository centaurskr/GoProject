///
/// @brief   structure의 Byte Stream을 돌려준다
/// @file    getbytes.go
/// @date    2024. 10. 11. (금) 14:36:49 KST
/// @author  Cento
///
package api

import (
	"unsafe"
	"bytes"
	"reflect"
)
////////////////////////////////////////////////////////////////////////////////
func CleanByte(input []byte) []byte{
	return  bytes.ReplaceAll(input[:], []byte{0x00}, nil)
}
////////////////////////////////////////////////////////////////////////////////
// byte에서 0x00을 빼고 String을 return
// 사용법 :
// var item [100]byte
// str := api.CleanString(item[:])
///////////////////////////////////////////////////////////////////////////////
func CleanString(input []byte) string{
	clean := bytes.ReplaceAll(input[:], []byte{0x00}, nil)
	return string(clean)
	//return string(bytes.Trim(input, "\x00"))
}
////////////////////////////////////////////////////////////////////////////////
/// Any type의 Byte stream을 돌려준다
/// @fn     : func GetBytes[T any](v *T)[]byte
/// @param  :
/// @return :
////////////////////////////////////////////////////////////////////////////////
func GetBytes[T any](v *T)[]byte{
	size := unsafe.Sizeof(*v)
	anybytes := (*[1 << 30]byte)(unsafe.Pointer(v))[:size:size]
	return anybytes
}
// Size 구하기
func SizeOf[T any](v T)int{return int(unsafe.Sizeof(v))}

// Type알아보기
func GetTypeName(a interface {})string{return reflect.TypeOf(a).String()}
