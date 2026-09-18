package main
import (
	"fmt"
	"time"
	"math/rand"
)
func main(){
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 100; i++{
		idx := rand.Intn(10)	
		fmt.Printf("I[%s] Rand.Intn(10) - [%d]\n", i, idx)
	}
}
