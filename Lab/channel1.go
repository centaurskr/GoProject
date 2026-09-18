///
/// @brief  : GOLANG channel test-1
/// @file   : channel1.go
/// @date   : 2024. 04. 26. (금) 14:26:40 KST
/// @author : Cento
///
package main
import (
	"fmt"
	"sync"
	"time"
)
func square(wg *sync.WaitGroup, ch chan int){
	fmt.Println("Square start")
	n := <- ch
	fmt.Println("Square : ", n * n)
	wg.Done()
	fmt.Println("Square end")
}
func main(){
	var wg sync.WaitGroup
	ch := make (chan int)

	wg.Add(1)

	go square(&wg, ch)

	time.Sleep(time.Second)
	fmt.Print("1초기다리기\n")

	ch <- 3
	fmt.Print("3 전송\n")

	wg.Wait()
}
