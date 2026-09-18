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
	// Open the file for reading
	file, err := os.Open("ORDER_MAP.MAP")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Get the file size
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}
	fileSize := fileInfo.Size()

	// Calculate the number of Order structs in the file
	orderSize := int64(binary.Size(Order{}))
	numOrders := fileSize / orderSize
	fmt.Printf("Reading %d orders from the file...\n", numOrders)

	// Read each Order from the file
	for i := int64(0); i < numOrders; i++ {
		var order Order
		err := binary.Read(file, binary.LittleEndian, &order)
		if err != nil {
			fmt.Println("Error reading from file:", err)
			return
		}

		// Print the Order
		fmt.Printf("index %d: OrdSeq = %d OrdType = %d, Midx = %d, BuySell = %d, OrdTime = %d, RegNum = %d, CheSeq = %d, Price = %v\n",
			i, order.OrdSeq, order.OrdType, order.Midx, order.BuySell, order.OrdTime, order.RegNum, order.CheSeq, order.Price)
	}
}

