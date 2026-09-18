package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"syscall"
	"unsafe"
)

// Define the structure for Order
type Order struct {
	OrdType int32
	Midx    int32
	BuySell int32
	OrdTime int64
	OrdSeq  uint64
	RegNum  uint64
	CheSeq  int32
	Price   [32]byte // Example Price field
}

// Memory map a file and return the byte slice that represents the file contents.
func mmapFile(filename string, length int) ([]byte, func(), error) {
	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		return nil, nil, err
	}

	// Memory-map the file into a byte slice
	data, err := syscall.Mmap(int(file.Fd()), 0, length, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, nil, err
	}

	// Unmap function to release memory-mapped region
	unmap := func() {
		err := syscall.Munmap(data)
		if err != nil {
			log.Fatalf("Error unmapping file: %v", err)
		}
		file.Close()
	}

	return data, unmap, nil
}

// getOrderPointer returns a pointer to the nth Order struct in the memory-mapped region.
func getOrderPointer(mappedData []byte, index int, orderSize int) unsafe.Pointer {
	offset := index * orderSize
	return unsafe.Pointer(&mappedData[offset])
}

// sortOrders sorts the orders in the memory-mapped file by OrdSeq field using unsafe pointers.
func sortOrders(mappedData []byte, numOrders int) {
	orderSize := int(unsafe.Sizeof(Order{})) // Size of each Order struct

	// Use sort.Slice with unsafe pointer to directly manipulate memory-mapped orders
	sort.Slice(make([]struct{}, numOrders), func(i, j int) bool {
		orderPtrI := (*Order)(getOrderPointer(mappedData, i, orderSize))
		orderPtrJ := (*Order)(getOrderPointer(mappedData, j, orderSize))

		// Compare OrdSeq fields for sorting
		return orderPtrI.OrdSeq < orderPtrJ.OrdSeq
	})
}

func main() {
	// Assume the file has 5 orders stored in it and is already memory-mapped
	filename := "orders_mmap.dat"
	numOrders := 5
	orderSize := int(unsafe.Sizeof(Order{}))

	// Open and map the file to memory
	mappedData, unmap, err := mmapFile(filename, numOrders*orderSize)
	if err != nil {
		log.Fatalf("Error mapping file: %v", err)
	}
	defer unmap() // Ensure the file is unmapped at the end

	// Read orders before sorting
	fmt.Println("Before sorting by OrdSeq:")
	for i := 0; i < numOrders; i++ {
		orderPtr := (*Order)(getOrderPointer(mappedData, i, orderSize))
		fmt.Printf("Order %d: OrdSeq = %d\n", i, orderPtr.OrdSeq)
	}

	// Sort the orders in the memory-mapped file by OrdSeq
	sortOrders(mappedData, numOrders)

	// Read orders after sorting
	fmt.Println(mappedData)
	fmt.Println("\nAfter sorting by OrdSeq:")
	for i := 0; i < numOrders; i++ {
		orderPtr := (*Order)(getOrderPointer(mappedData, i, orderSize))
		//fmt.Printf("Order %d: OrdSeq = %d\n", i, orderPtr.OrdSeq)
		fmt.Printf("Order %d: %v\n", i, orderPtr)
	}
}

