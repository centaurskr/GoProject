package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

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

func main() {
	// Sample orders to write into the file
	orders := []Order{
		{OrdSeq: 5, OrdType: 1, Midx: 100, BuySell: 1, OrdTime: 123456, RegNum: 500, CheSeq: 10, Price: [32]byte{1, 2, 3}},
		{OrdSeq: 2, OrdType: 1, Midx: 101, BuySell: 1, OrdTime: 123457, RegNum: 501, CheSeq: 11, Price: [32]byte{4, 5, 6}},
		{OrdSeq: 8, OrdType: 1, Midx: 102, BuySell: 1, OrdTime: 123458, RegNum: 502, CheSeq: 12, Price: [32]byte{7, 8, 9}},
		{OrdSeq: 1, OrdType: 1, Midx: 103, BuySell: 1, OrdTime: 123459, RegNum: 503, CheSeq: 13, Price: [32]byte{10, 11, 12}},
		{OrdSeq: 7, OrdType: 1, Midx: 104, BuySell: 1, OrdTime: 123460, RegNum: 504, CheSeq: 14, Price: [32]byte{13, 14, 15}},
	}

	// Create or overwrite the file
	file, err := os.Create("Order.MAP")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Write orders to the file in binary format
	for _, order := range orders {
		err := binary.Write(file, binary.LittleEndian, order)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
	}

	fmt.Println("Order.MAP file created successfully!")
}

