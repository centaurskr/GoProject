package main

import (
	//"encoding/binary"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// Order struct definition
type Order struct {
	OrdSeq  uint64
	OrdType int32
	Midx    int32
	BuySell int32
	OrdTime int64
	RegNum  uint64
	CheSeq  int32
	Price   [32]byte
}

func main() {
	// Open the file for reading
	file, err := os.OpenFile("ORDER_MAP.MAP", os.O_RDWR, 0600)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	//defer file.Close()

	// Get the file size
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}
	fileSize := fileInfo.Size()

	// Calculate the size of Order struct using unsafe.Sizeof
	orderSize := int64(unsafe.Sizeof(Order{}))
	numOrders := fileSize / orderSize
	fmt.Printf("Reading %d orders from the file...\n", numOrders)

	// Memory map the file
	data, err := syscall.Mmap(int(file.Fd()), 0, int(fileSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		fmt.Println("Error memory mapping the file:", err)
		return
	}
	file.Close()

	defer syscall.Munmap(data)

	// Read each Order from the memory-mapped file
	for i := int64(0); i < numOrders; i++ {
		// Calculate the offset for the ith Order
		offset := i * orderSize
		orderData := data[offset : offset+orderSize]

		// Convert the byte slice to an Order struct
		order := (*Order)(unsafe.Pointer(&orderData[0]))

		// Print the Order
		fmt.Printf("index %d: OrdSeq = %d, OrdType = %d, Midx = %d, BuySell = %d, OrdTime = %d, RegNum = %d, CheSeq = %d, Price = %s\n",
			i, order.OrdSeq, order.OrdType, order.Midx, order.BuySell, order.OrdTime, order.RegNum, order.CheSeq, order.Price)
	}
}

