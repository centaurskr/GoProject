///
/// @brief  : GOLANG 방향성 channel
/// @file   : channeldir.go
/// @date   : 2024. 04. 26. (금) 16:09:05 KST
/// @author : Cento
///
package main
import (
	"fmt"
	"sync"
	"time"
)

func square(wg *sync.WaitGroup, ch <-chan int){
	for n:= range ch{
		fmt.Println("Square:", n *n)
		time.Sleep(time.Second)
	}
	fmt.Println("Done")
	wg.Done()
}
func main(){
	var wg sync.WaitGroup
	ch := make(chan int)

	wg.Add(1)
	go square(&wg, ch)
	
	for i:=0; i < 10; i++{
		ch <- i*2	
	}
	close(ch)
	wg.Wait()
}
