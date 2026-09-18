package main
import (
	"fmt"
	"strconv"
)
type Adder interface{
	Add(val interface{}) interface{}
}

type IntAdder int
type StrAdder string

func (i IntAdder)Add(val interface{})interface{}{
	switch v:=val.(type){
		case IntAdder :
			return int(i) + int(v)
		case StrAdder :
			ii, _:=strconv.Atoi(string(v))
			return int(i) + ii
		default :
			return -1
	}
}
func (s StrAdder)Add(val interface{})interface{}{
	switch v:=val.(type){
		case IntAdder :
			i, _ := strconv.Atoi(string(s))
			return i + int(v)
		case StrAdder :
			i, _ := strconv.Atoi(string(s))
			ii,_ := strconv.Atoi(string(v))
			return i + ii
		default :
			return -1
	}

}
func Add(a, b Adder) interface{}{
	return a.Add(b)
}
func main(){
	i := 10
	str := "20"

	fmt.Printf("int(%d) + string(%s) = %d\n", i, str, Add(IntAdder(i), StrAdder(str)))
	fmt.Printf("string(%s) + string(%s) = %d\n", str, str, Add(StrAdder(str), StrAdder(str)))
	fmt.Printf("string(%s) + int(%d) = %d\n", str, i, Add(StrAdder(str), IntAdder(i)))
	fmt.Printf("int(%d) + int(%d) = %d\n", i, i, Add(IntAdder(i), IntAdder(i)))
}
