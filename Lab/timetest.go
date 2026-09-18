package main

import (
	"fmt"
	"time"
	"reflect"
	"strconv"
)

func main(){
	fmt.Println("123456789012345678901234567890123456789012345678901234567890")
	fmt.Println(time.Now())

	uxtime := time.Now()
	fmt.Println("Print Unix()      : ", uxtime.Unix())
	fmt.Println("Print UnixMilli() : ", uxtime.UnixMilli())
	fmt.Println("Print UnixMicro() : ", uxtime.UnixMicro())
	fmt.Println("Print UnixNano()  : ", uxtime.UnixNano())

	fmt.Println("UnixNano Return type :", reflect.TypeOf(uxtime.UnixNano()))
	//ipcstr := strconv.FormatInt(uxtime.UnixNano(), 10)
	ipcstr := "inproc://" + strconv.FormatInt(uxtime.UnixNano(), 10) + ".mon"

	fmt.Println(ipcstr)

	fmt.Println("GoString():", uxtime.GoString())
	
	fmt.Println("Timeonly+Nano : ", uxtime.Format(time.TimeOnly), uxtime.Nanosecond())
	
	seqstr := uxtime.Format("20060102") + "00000000000"
	seqint,_ := strconv.ParseUint(seqstr, 10, 64)
	fmt.Println("12345678901234567890")
	fmt.Println(seqstr)
	fmt.Println(seqint)

	seqstr = uxtime.Format("060102") + "-00000000000"
	fmt.Printf("%s\n", seqstr)
	
}
