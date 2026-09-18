///
/// @brief  :GOLANG channel select Test
/// @file   : select.go
/// @date   : 2024. 04. 26. (금) 15:56:47 KST
/// @author : Cento
///
package main
import (
	"fmt"
	"sync"
	"time"
)

func square(wg *sync.WaitGroup, ch chan int){
	tick:= time.Tick(time.Second)
	terminate := time.After(10*time.Second)
	for{
		select{
		case <-tick :
			fmt.Println("TICK")
		case <-terminate:
			fmt.Println("Terminate")
			wg.Done()
			return
		case n:=<-ch:
			fmt.Printf("%d Square: %d\n",n, n * n)
			time.Sleep(time.Second)	
		}
	}
}

func main(){
	var wg sync.WaitGroup
	ch  := make(chan int)

	wg.Add(1)
	go square(&wg, ch)
	
	for i:= 0; i < 10; i++{
		ch <- i*2
	}
	close(ch)

	wg.Wait()
}

