///
/// @brief  : GOLANG INTERFACE Test
/// @file   : interface.go
/// @date   : 2024. 04. 18. (목) 11:03:35 KST
/// @author : Cento
///

package main
import (
	"fmt"
	"math"
)
// Interface 정의
type Abser interface{
	Abs() float64
}

// Abs 정의
type MyFloat float64
func (f MyFloat) Abs() float64{
	if f < 0 {
		return float64(-f)
	}
	return float64(f)
}

type Vertex struct{
	X,Y float64
}
func (v *Vertex) Abs() float64{
	return math.Sqrt(v.X * v.X + v.Y * v.Y)
}

func main(){
	var a Abser
	f := MyFloat(-math.Sqrt2)
	v := Vertex{3, 4}

	a = f
	fmt.Println(a.Abs())

	a = &v
	fmt.Println(a.Abs())

	a = v
	fmt.Println(a.Abs())

}

