///
/// @brief  : GLANG string test
/// @file   : string.go
/// @date   : 2024. 04. 30. (화) 10:52:43 KST
/// @author : Cento
///
package main

import(
	"fmt"
	"reflect"
	"unsafe"
	"strings"
	)

func main(){
	str1:="Hello\nWorld!"
	fmt.Println(str1)

	str2:=`Hello
	World`
	fmt.Println(str2)

	fmt.Printf("\"%s\" 를 하나씩 출력\n", str1)
	for i:=0;i<len(str1);i++{
		fmt.Printf("%c", str1[i])
	}
	fmt.Printf("\n\n")

	str3:="Hello 월드!"
	fmt.Printf("\"%s\"를 하나씩 출력(한글출력)\n", str3)
	arr := []rune(str3)
	for i:= 0; i< len(arr); i++{
		fmt.Printf("%c ", arr[i])
	}
	fmt.Printf("\n\n")

	str4:="a"
	str5:="b"

	fmt.Println("a b")
	fmt.Println("==", str4==str5)
	fmt.Println("!=", str4!=str5)
	fmt.Println("<",  str4<str5)
	fmt.Println(">",  str4>str5)

	strHeader1 := (*reflect.StringHeader)(unsafe.Pointer(&str1))
	strHeader2 := (*reflect.StringHeader)(unsafe.Pointer(&str2))
	fmt.Printf("String Header 출력\n")
	fmt.Println("str1 :", strHeader1)
	fmt.Println("str2 :", strHeader2, "\n")
	
	str1 = "String Bytes"
	strbyte := []byte(str1)
	fmt.Printf("String Byte 출력 : \"%s\"\n", str1)
	fmt.Printf("len(%d): %v\n\n", len(strbyte), strbyte)

	var orglen int
	orglen = 10240
	fmt.Printf("Size(%d) -> 4 bytes로 변경\n", orglen)
	var lenbyte [4]byte
	lenbyte[0] = byte(orglen >> 24)
	lenbyte[1] = byte(orglen >> 16)
	lenbyte[2] = byte(orglen >> 8)
	lenbyte[3] = byte(orglen)
	fmt.Printf("LenByte : %v\n", lenbyte)

	var relen int
	relen = int(lenbyte[0]) << 24 |
		int(lenbyte[1]) << 16 |
		int(lenbyte[2]) << 8 |
		int(lenbyte[3]) 
	fmt.Printf("len (int) <- lenbyte\n")
	fmt.Printf("Result : %d\n\n", relen)

	// 문자열 변경
	// slice 변수에 할당 후 변경
	var slice []byte = []byte(str1)

	slice[0] = "h"[0]
	fmt.Println("str1 : ", str1)
	fmt.Println("slice : ", slice)
	fmt.Printf("slice : %s\n", slice)
	
	fmt.Println()

	str:="Hello, World!"
	fmt.Println(strings.ToUpper(str))
	fmt.Println(strings.Contains(str, "Hello"))
	fmt.Println(strings.Index(str, "W"))
	fmt.Println(strings.Compare("a", "b"))

	// 문자 지우기
	fmt.Println("\n문자열중 마지막 \",\" 지우기")
	//str = `"S":[{"P":"123.123", "C":"10", "Q":"456.456"},]`
	str = `"S":[{"P":"123.123", "C":"10", "Q":"456.456"},`
	//str = `"S":[,]`
	//str = `"S":[]`
	//str = `"S":[`
	fmt.Printf("SOURCE     : %s\n", str)
	idx := strings.LastIndex(str, ",")
	if idx != -1{
		if idx == len(str)-1 || strings.Contains("}]", string(str[idx+1])){
			str = str[:idx] + str[idx+1:]
		}
	}
	fmt.Printf("After Trim : %s\n", str)
	fmt.Println()

	str = "2006010200000000000"
	fmt.Printf("%s[0:8] : %s\n", str, str[0:8])
}
