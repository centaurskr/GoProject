/// 
/// @brief  : GOLANG defer Test(LIFO)
/// @file   : defer.go
/// @date   : 2024. 04. 19. (금) 16:09:55 KST
/// @author : Cento
///

package main
import "fmt"

func trace(s string) string{
	fmt.Println("entering:", s)
	return s
}

func un(s string){
	fmt.Println("\tleaving:", s)
}
func a(){
	defer un(trace("a"))
	fmt.Println("in a")
}
func b(){
	defer un(trace("b"))
	fmt.Println("in b")
	a()
}

func main(){
	b()
}
