package main
import (
	"fmt"
	Dc "github.com/shopspring/decimal"
)
func main() {
	var d1, d2, d3 Dc.Decimal

	d1,_ = Dc.NewFromString("1.1")
	d2,_ = Dc.NewFromString("2.1")
	d2 = d1.Add(d2)
	d3 = d2
	fmt.Printf("D1[%v] D2[%v] D3[%v]\n", d1, d2,d3)
}
