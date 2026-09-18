///
/// @brief  : GOLANG Pointer test
/// @file   : pointer1.go
/// @date   : 2024. 04. 22. (월) 15:27:44 KST
/// @author : Cento
///
package main
import "fmt"
import "math"

type Vertex struct {
	X, Y float64
}

func (v Vertex)Abs() float64{
	return math.Sqrt(v.X * v.X + v.Y * v.Y)
}

func (v *Vertex) ScaleP(f float64){
	v.X = v.X * f
	v.Y = v.Y *f
}

func (v Vertex) Scale(f float64){
	v.X = v.X * f
	v.Y = v.Y *f
}
func scaleFun(v *Vertex, f float64){
	v.X = v.X * f
	v.Y = v.Y * f
}
func newFun(a, b float64) *Vertex{
	vert := &Vertex{
			X:a,
			Y:b,
		}
	return vert
}
func main(){
	v := Vertex{3,4}

	fmt.Println("V is " , v)

	vp := newFun(1.1, 2.1)	
	fmt.Println("Vertext pointer return Test-1 : ", *vp)

	fmt.Println("\tbefore ScaleP(10)", v)
	v.ScaleP(10)
	fmt.Println("\taftter ScaleP(10)", v)

	fmt.Println("V is " , v)

	fmt.Println("\tbefore Scale(10)", v)
	v.Scale(10)
	fmt.Println("\taftter Scale(10)", v)

	fmt.Println("V is " , v)

	fmt.Println("\tbefore ScaleP(10)", v)
	v.ScaleP(10)
	fmt.Println("\taftter ScaleP(10)", v)

	fmt.Println("v.Abs() is : ", v.Abs())

	fmt.Println("Scale function Text ----");
	scaleFun(&v, 10)
	fmt.Println("scaleFun(v, 10): ", v)

	v1 := Vertex{5,6}

	v1.ScaleP(5)
	fmt.Println("v1.ScaleP(5) ", v1)
	p := &v1

	p.ScaleP(10)
	fmt.Println("v1.ScaleP(10) ", v1)

	vp.X = 3.1
	vp.Y = 4.1
	fmt.Println("Vertext pointer return Test-2 : ", *vp)


}
