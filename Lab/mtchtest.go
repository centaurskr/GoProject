package main
import (
	Log "centolib/logout"
	"time"
)
func FuncA(ch <-chan int){
	for n := range ch{
		Log.Infor("A : %d",n)
	}
}

func FuncB(ch <-chan int){
	for n := range ch{
		Log.Infor("B : %d",n)
	}
}
func main(){
	Log.InitLogger("")
	defer Log.Close()
	
	ch := make(chan int)
	go FuncA(ch)
	go FuncB(ch)
	
	for i := 0; i < 10; i++{
		ch <- i
		time.Sleep(1*time.Second)
	}
	close(ch)

}
