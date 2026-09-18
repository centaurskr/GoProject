///
/// @brief  : GOLANG Context TEST
/// @file   : context.go
/// @date   : 2024. 04. 19. (금) 15:13:29 KST
/// @author : Cento
///
package main
import(
	"context"
	"fmt"
	"sync"
	"time"
)
var wg sync.WaitGroup

func PrintTick(ctx context.Context){
	tick := time.Tick(time.Second)
	for{
		select {
			case <-ctx.Done():
				fmt.Println("Done :", ctx.Err())
				wg.Done()
				return
			case <-tick:
				fmt.Println("tick")
		}
	}
}
func main(){
	wg.Add(1)
	ctx,cancel := context.WithCancel(context.Background())

	go PrintTick(ctx)
	
	time.Sleep(5*time.Second)
	cancel()

	wg.Wait()
}

