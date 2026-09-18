///
/// @brief  : GOLANG interface Duck typing
/// @file   : interface-duc.go
/// @date   : 2024. 04. 18. (목) 14:12:55 KST
/// @author : Cento
///
package main
import "fmt"

/// Interface 정의
type Quacker interface{
	quack()
	feathers()
}
/// Duck
type Duck struct{
}
func (d Duck) quack(){ fmt.Println("꽥!!!!") }
func (d Duck) feathers(){fmt.Println("오리는 흰색과 회색 털을 가지고 있습니다")}
/// Person
type Person struct{
}
func (d Person) quack(){ fmt.Println("사람은 오리 흉내를 냅니다. 꽥!!!!") }
func (d Person) feathers(){fmt.Println("사람은 기털을 주워보여줍니다")}

func inTheForest(q Quacker){
	q.quack()
	q.feathers()
}

func main(){
	var donal Duck
	var john  Person

	inTheForest(donal)
	inTheForest(john)
}
