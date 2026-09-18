///
/// @brief  : GOLANG channel test-3
/// @file   : channel3.go
/// @date   : 2024. 04. 26. (금) 14:45:46 KST
/// @author : Cento
///
package main
import (
	"fmt"
	"sync"
	"time"
)

func square(wg *sync.WaitGroup, ch chan int){
	tick := time.Tick(time.Second)
	term := time.After(10 * time.Second)
	for {
		select{
		case <-tick :
			fmt.Println("Tick")
		case <-term :
			fmt.Println("Terminated!")
			wg.Done()
			return
		case n := <-ch:
			fmt.Printf("Square:%d\n", n*n)
			time.Sleep(time.Second)
		}
	}
}

func main(){
	var wg sync.WaitGroup
	ch := make (chan int)
	
	wg.Add(1)
	go square(&wg, ch)
	
	for i:=0; i < 10; i++{
		ch <- i*2
	}
	close(ch) // daedlock 방지를 위함
	
	wg.Wait()
}



