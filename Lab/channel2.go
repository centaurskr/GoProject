///
/// Channel deadlock!
/// @brief  : GOLANG channel test-2
/// @file   : channel2.go
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
	for n:= range ch{
		fmt.Printf("%d square : %d\n",n,  n *n)
		time.Sleep(time.Second)
	}
	fmt.Println("Done")
	wg.Done()
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



