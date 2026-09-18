///
/// @brief  : GLANG Interface 사용 Test #1
/// @file   : interface-1.go
/// @date   : 2024. 04. 18. (목) 13:52:36 KST
/// @author : Cento
///
package main
import  "fmt"
import "math"

type Shape interface {
	area() float64
	perimeter() float64	
}

type Rect struct{
	width, height float64
}

type Circle struct{
	radius float64
}

/// Rect의 Shape interface 구현
func (r Rect) area() float64{
	return r.width * r.height
}
func (r Rect) perimeter() float64 {
	return 2 * r.width * r.height
}

/// Circle의 Shape interface 구현
func (c Circle) area() float64{
	return math.Pi * c.radius * c.radius
}
func (c Circle) perimeter() float64{
	return 2 * math.Pi * c.radius
}

/// showArea()
func showArea(shapes ...Shape){
	for _, s:= range shapes{
		a:= s.area()
		fmt.Println(a, s)
	}
}
func main(){
	r := Rect{10., 20.}
	c := Circle{10}

	showArea(r, c)
}
