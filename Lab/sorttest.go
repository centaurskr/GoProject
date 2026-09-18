package main

import (
	"fmt"
	"os"
	"sort"
	"syscall"
	"unsafe"
	"strings"
	"centolib/api"
)

// Order struct definition
type Order struct {
	Symbol  [32]byte
	OrdSeq  uint64
	OrdType int32
	Midx    int32
	BuySell int32
	OrdTime int64
	RegNum  uint64
	CheSeq  int32
	Price   [32]byte
}
var Symbol = []string{"BTC.USDT", "AAVE.USDT", "TTAD.USDT", "GOVE.USDT", "SSD.USDT"}

// Create or open the memory-mapped file
func createMapfile(filename string, size int) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}

	// Extend the file to the required size
	err = file.Truncate(int64(size))
	if err != nil {
		file.Close()
		return nil, err
	}

	return file, nil
}

// Attach to the memory-mapped file
func attachMapfile(filename string, size int) ([]byte, error) {
	file, err := os.OpenFile(filename, os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	// Map the file into memory
	data, err := syscall.Mmap(int(file.Fd()), 0, size, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		file.Close()
		return nil, err
	}

	return data, nil
}

// Convert the Order struct to a byte slice for writing to mmap
func orderToBytes(order *Order) []byte {
	size := unsafe.Sizeof(*order)
	orderBytes := (*[1 << 30]byte)(unsafe.Pointer(order))[:size:size]
	return orderBytes
}

// Convert a byte slice to an Order struct for reading from mmap
func bytesToOrder(data []byte) *Order {
	return (*Order)(unsafe.Pointer(&data[0]))
}

// Calculate offset for nth order in mmap
func getOrderOffset(index int, orderSize int) int {
	return index * orderSize
}

func main() {
	filename := "ORDER_MAP.MAP"
	orderSize := int(unsafe.Sizeof(Order{})) // Size of each Order struct
	numOrders := 5                           // Number of orders to store
	totalSize := orderSize * numOrders       // Total size of the mmap file

	// Step 1: Create the memory-mapped file
	file, err := createMapfile(filename, totalSize)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	file.Close()

	fmt.Printf("===============MAP File : %s==============\n", filename)
	// Step 2: Attach to the memory-mapped file
	data, err := attachMapfile(filename, totalSize)
	if err != nil {
		fmt.Println("Error attaching mmap:", err)
		return
	}
	defer syscall.Munmap(data) // Unmap when done

	// Step 3: Write multiple orders to mmap
	for i := 0; i < numOrders; i++ {
		ord := &Order{
			OrdSeq:  uint64(numOrders - i), // Reversed for testing sorting
			OrdType: int32(i + 1),
			Midx:    int32(i + 1),
			BuySell: int32(i % 2), // Alternate Buy/Sell
			OrdTime: 1234567890 + int64(i),
			RegNum:  56789 + uint64(i),
			CheSeq:  int32(i + 10),
		}
		copy(ord.Price[:], fmt.Sprintf("%d00.%d0", i+1, i)) // Example price
		copy(ord.Symbol[:], fmt.Sprintf("%s", Symbol[i]))

		// Calculate the offset for the ith order
		offset := getOrderOffset(i, orderSize)
		copy(data[offset:], orderToBytes(ord))
	}
	// Step 4 : Dump mapfile
	for i := 0; i < numOrders; i++{
		pord := (*Order)( unsafe.Pointer(&data[int(unsafe.Sizeof(Order{})) * i]))
		fmt.Printf("INDEX [%d] OrdSeq[%d] Symbol[%s]\n", i, pord.OrdSeq, string(pord.Symbol[:]))
	}

	byName := func (a, b unsafe.Pointer) bool{
			o1 := (*Order)(a)
			o2 := (*Order)(b)
			rtn := strings.Compare(string(o1.Symbol[:]), string(o2.Symbol[:]))
			return rtn < 0 // 오름차순 (중요  rtn < 0)
	}
	// 5. Sort by name
	sortOrder := api.SortSlice{
		StBase  : uintptr(unsafe.Pointer(&data[0])),
		StCount : numOrders,
		StSize  : unsafe.Sizeof(Order{}),
		LessFunc : byName,
	}

	sort.Sort(&sortOrder)

	fmt.Printf("\n============================Symbol Sort 후\n")
	for i := 0; i < numOrders; i++{
		pord := (*Order)( unsafe.Pointer(&data[int(unsafe.Sizeof(Order{})) * i]))
		fmt.Printf("INDEX [%d] OrdSeq[%d] Symbol[%s]\n", i, pord.OrdSeq, string(pord.Symbol[:]))
	}
	fmt.Println()

	var key Order

	sModes := []api.SearchMode{
		api.SMODE_EQ, api.SMODE_GT, api.SMODE_LT, api.SMODE_GE, api.SMODE_LE}
	sNames := []string{"EQ", "GT", "LT", "GE", "LE"}
	//copy(key.Symbol[:], "CBTC.USDT")
	//copy(key.Symbol[:], "TTE.USDT")
	copy(key.Symbol[:], "AAVE.USDT")
	item := sortOrder.Search(unsafe.Pointer(&key))
	if item == nil{
		fmt.Printf("Search [%s] 없습니다\n\n", string(key.Symbol[:]))
	}else{
		pord := (*Order)(item)
		fmt.Printf("Search [%s] OK\n", string(key.Symbol[:]))
		fmt.Printf("%+v\n\n", pord)
	}

	for i, mode := range sModes {
		fmt.Printf("SearchMode [%s] [%s] ", string(key.Symbol[:]), sNames[i])
		item := sortOrder.SearchMode(unsafe.Pointer(&key), mode)
		if item == nil {
			fmt.Printf(" NOT FOUND\n")
		} else {
			pord := (*Order)(item)
			fmt.Printf(" -> %s\n", string(pord.Symbol[:]))
		}
	}
	
	// 6. Sort By OrdSeq
	bySeq := func (a, b unsafe.Pointer) bool{
			o1 := (*Order)(a)
			o2 := (*Order)(b)
			return o1.OrdSeq < o2.OrdSeq
	}
	sortOrder.LessFunc = bySeq
	sort.Sort(&sortOrder)

	fmt.Printf("\n============================OrdSeq Sort 후\n")
	for i := 0; i < numOrders; i++{
		pord := (*Order)( unsafe.Pointer(&data[int(unsafe.Sizeof(Order{})) * i]))
		fmt.Printf("INDEX [%d] OrdSeq[%d] Symbol[%s]\n", i, pord.OrdSeq, string(pord.Symbol[:]))
	}
	fmt.Println()
	// 7.Search by Order Seq
	key = Order{OrdSeq:4}
	item = sortOrder.Search(unsafe.Pointer(&key))
	if item == nil{
		fmt.Printf("Search key.OrdSeq = 4 없습니다\n\n")
	}else{
		pord := (*Order)(item)
		fmt.Printf("Search key.OrdSeq = 4 OK\n")
		fmt.Printf("%+v\n\n", pord)
	}

}

