package main
import (
	"fmt"
	"github.com/bwmarrin/snowflake"
	"sync"
	//"time"
)
var Node *snowflake.Node

func goProc(idx int, wg *sync.WaitGroup){
 	defer wg.Done()

	//Node, _ := snowflake.NewNode(int64(idx)) 
	for i := 0; i < 100; i++{
		uniqKey := Node.Generate()
		fmt.Printf("INDEX [%3d] key[%d]\n", idx, uniqKey)
	}

}

// 같은 Process의 다수 goroutine에서 Unique Key 생성 
func main() {
	var wg sync.WaitGroup
	Node, _ = snowflake.NewNode(1)
	for i:= 0; i < 10; i++{
		wg.Add(1) 
		go goProc(i, &wg)	
	}
	wg.Wait() // Wait for all goroutines to finish
}

