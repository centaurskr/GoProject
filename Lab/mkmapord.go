package main

import (
	"encoding/binary"
	"syscall"
	"bytes"
	"fmt"
	"os"
)
const mapMode = syscall.PROT_READ | syscall.PROT_WRITE
// Define the structure for Order
type Order struct {
	OrdSeq  uint64
	OrdType int32
	Midx    int32
	BuySell int32
	OrdTime int64
	RegNum  uint64
	CheSeq  int32
	Price   [32]byte // Example Price field
}

func createMapfile(filename string, size int) ([]byte, error){
    _, err :=os.Stat(filename)
    if err == nil{
        return nil, fmt.Errorf("File Exist")
    }
    f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
    if err != nil {
        return nil, err
    }

    if err := f.Truncate(int64(size)); err != nil{
        return nil, err
    }
    mappoint , err := syscall.Mmap(int(f.Fd()), 0, size, mapMode, syscall.MAP_SHARED)
    if err != nil {
        return nil, err
    }
    f.Close()
    return mappoint, nil
}
func attachMapfile(filename string)([]byte, error){
	finfor, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	
    f, err := os.OpenFile(filename,os.O_RDWR, 0666)
    if err != nil {
        return nil, err
    }

    mappoint, err := syscall.Mmap(int(f.Fd()), 0, int(finfor.Size()), mapMode, syscall.MAP_SHARED)
    if err != nil {
		fmt.Println("Mmap()")
        return nil, err
    }
    f.Close()
    return mappoint, nil
}
func main() {
	// Sample orders to write into the file
	orders := []Order{
		{OrdType: 1, Midx: 100, BuySell: 1, OrdTime: 123456, OrdSeq: 5, RegNum: 500, CheSeq: 10, Price: [32]byte{1, 2, 3}},
		{OrdType: 1, Midx: 101, BuySell: 1, OrdTime: 123457, OrdSeq: 2, RegNum: 501, CheSeq: 11, Price: [32]byte{4, 5, 6}},
		{OrdType: 1, Midx: 102, BuySell: 1, OrdTime: 123458, OrdSeq: 8, RegNum: 502, CheSeq: 12, Price: [32]byte{7, 8, 9}},
		{OrdType: 1, Midx: 103, BuySell: 1, OrdTime: 123459, OrdSeq: 1, RegNum: 503, CheSeq: 13, Price: [32]byte{10, 11, 12}},
		{OrdType: 1, Midx: 104, BuySell: 1, OrdTime: 123460, OrdSeq: 7, RegNum: 504, CheSeq: 14, Price: [32]byte{13, 14, 15}},
	}
	filename := "Order.MAP"
	mappoint , err := createMapfile(filename, 1024)
	if err != nil {
		fmt.Println(err)
		return
	}
	syscall.Munmap(mappoint)
	fmt.Println("Create OK ")

	mappoint, err = attachMapfile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	ordsize := binary.Size(Order{})
	for index, order := range orders {
		offset := index * ordsize
		fmt.Printf("Index : %d, Offset : %d\n", index, offset)
		copy(mappoint[offset:], order)

		//buf := mappoint[offset:offset+ordsize]
		//if err := binary.Write(bytes.NewBuffer(buf), binary.LittleEndian, order); err != nil{
			//fmt.Printf("Line : %d , Error : %v\n", index, err)
			//break;
		//}
	}

	for i := 0 ; i < 5; i++{
		var ord Order
		offset := i * ordsize
		err := binary.Read(bytes.NewBuffer(mappoint[offset: offset+ordsize]), binary.LittleEndian, &ord)
		if err != nil {
			fmt.Println("ERROR : index ", i)
			break
		}
        fmt.Printf("index %d: OrdSeq = %d OrdType = %d, Midx = %d, BuySell = %d, OrdTime = %d, RegNum = %d, CheSeq = %d, Price = %v\n",
            i, ord.OrdSeq, ord.OrdType, ord.Midx, ord.BuySell, ord.OrdTime, ord.RegNum, ord.CheSeq, ord.Price)

	}
	syscall.Munmap(mappoint)
	fmt.Println("Write Ok")
}

