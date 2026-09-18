///
/// @brief   Test Golan enum
/// @file   : enum.go
/// @date   : 2024. 10. 14. (월) 12:36:02 KST
/// @author : Cento
///
package main
import "fmt"

type Month int

const (
	January Month = 1 + iota
	February
	March
	April
	May
	June
	July
	August
	September
	October
	November
	December
)

var months = [...]string{
	"January",
	"February",
	"March",
	"April",
	"May",
	"June",
	"July",
	"August",
	"September",
	"October",
	"November",
	"December",
}
func (m Month) String() string { return months[(m-1)%12] }

type BuySell int
const ( 
	sell BuySell = 1
	buy 
)
var buysells = [...] string {"MEDO", "MESU",}

func (bs BuySell)String() string {return buysells[bs-1]}

func main() {
	month := December

	if month == December {
		fmt.Println("Found a December")
	}

	month = month + Month(2)
	fmt.Println(month, Month(2))

	month = January + Month(2)
	fmt.Println(month)

	month++
	fmt.Println(month)

	day := 34
	month = Month(day % 31)
	fmt.Println(month)

	val := int(month) + 4
	fmt.Println(val)

	month = Month(val) + 1
	fmt.Println(month)

	fmt.Println(BuySell(1), BuySell(2))

	var bins []byte
	bins = []byte("String")
	fmt.Println(bins, "String")

	var bins2 [32]byte
	copy(bins2[:], fmt.Sprintf("MEDO.%04d", 2))

	fmt.Println(bins2)
	fmt.Println(string(bins2[:]))

	//
	str := "Copy Test 2"
	copy(bins2[:],  str[:])
	fmt.Println(bins2)
	fmt.Println(string(bins2[:]))

}
