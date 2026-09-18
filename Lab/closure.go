package main
import (
	Log "centolib/logout"
)
func nextValue() func() int{
	i := 0
	return func() int {
		i++
		return i
	}
}

func main() {
	Log.InitLogger("")
	Log.Infor("Closer Test 1")
	next := nextValue()

	println("next()" , next())
	println("next()" , next())
	println("next()" , next())

	anotherNext := nextValue()
	println("anotherNext() ", anotherNext())
	println("anotherNext() ", anotherNext())

	println("next()" , next())

	Log.Infor("Clostr Test 2")

	funcList := make([]func(), 3)
	v := 1

	for i:= 0;i< 3; i++{
		Log.Infor("Inserting %v", i)
		funcList[i] = func(){
			Log.Infor("\t%v", v)
		}
	}
	Log.Debug("Current funcList's V %v", v)
	for i:= 0; i < len(funcList); i++{
		funcList[i]()
	}

	v = 4
	Log.Debug("Current V is %v", v)
	for i:= 0; i < len(funcList); i++{
		funcList[i]()
	}

	v = 19999
	Log.Debug("Current V is %v", v)
	for i:= 0; i < len(funcList); i++{
		funcList[i]()
	}
}



