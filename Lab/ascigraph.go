// 먼저 해야할것
// go get github.com/guptarohit/asciigraph
//
package main

import (
	"fmt"
	"github.com/guptarohit/asciigraph"
)

func main() {
	fmt.Println("hello")

	data := []float64{3, 4.5, 9.3, 6, 2, 4, 5, 8, 5, 10, 2, 7, 2, 5, 6}
	graph := asciigraph.Plot(data)
	fmt.Println(graph)

}
