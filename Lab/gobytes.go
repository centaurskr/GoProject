///
/// @brief   Convert "go structure" <--> "bytes"
/// @file    gobytes.go
/// @date    2024. 10. 07. (월) 16:20:50 KST
/// @author  Cento
///
package main
import (
	"encoding/binary"
	"bytes"
	//"io"
	"reflect"
	"fmt"
	Dc "github.com/shopspring/decimal"
)
type BinDecimal struct {
	Bin   [32]byte
	Binlen    int32
}
type Order struct {
	OrdType int32
	Midx    int32
	BuySell int32
	OrdTime int64
	OrdSeq  uint64
	RegNum  uint64
	CheSeq  int32

	Price   BinDecimal
}
func main() {
	var prc Dc.Decimal
	var buf bytes.Buffer
	var ord, decode Order

	// Assign values to the Order fields
	ord.OrdType = 1
	ord.Midx = 1
	ord.BuySell = 1
	ord.OrdTime = 123
	ord.OrdSeq = 123
	ord.RegNum = 123
	ord.CheSeq = 0

	// Convert string price to decimal.Decimal
	strprc := "999999999999.999999999999"
	prc, _ = Dc.NewFromString(strprc)

	// Marshal decimal price to binary
	binprc, _ := prc.MarshalBinary()

	// Store the binary price in the fixed-size array and its length in Binlen
	copy(ord.Price.Bin[:], binprc)
	ord.Price.Binlen = int32(len(binprc))

	fmt.Println("Orignal Ord : ", ord)
	fmt.Println("Original decimal price:", prc)

	// Encode the entire Order struct into the binary buffer
	err := binary.Write(&buf, binary.LittleEndian, ord)
	if err != nil {
		fmt.Println("Error during binary.Write:", err)
		return
	}
	
	fmt.Println("Buff type : ", reflect.TypeOf(buf))
	fmt.Println("After Write, buf: ", buf)

	// Decode the binary buffer back into the Order struct
	err = binary.Read(bytes.NewReader(buf.Bytes()), binary.LittleEndian, &decode)
	if err != nil {
		fmt.Println("Error during binary.Read:", err)
		return
	}

	fmt.Println("Decoded Order struct:", decode)

	// Unmarshal the price from the decoded struct's binary Price field
	var prc2 Dc.Decimal
	err = prc2.UnmarshalBinary(decode.Price.Bin[:decode.Price.Binlen])
	if err != nil {
		fmt.Println("Error during UnmarshalBinary:", err)
		return
	}

	// Output the unmarshalled decimal price
	fmt.Println("Unmarshalled decimal price:", prc2)
}
