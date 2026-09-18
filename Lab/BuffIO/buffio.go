//
// channel을 이용한 buffered IO
//
package main
import (
	"fmt"
	"math/rand"
    "time"
)
type Message struct {
	seq int
	data []byte
}
func goProcess(ch chan Message){
for{
	select {
		case msg := <- ch :
			fmt.Printf("SEQ[%d] msg[%v]\n", msg.seq, msg.data)
	}
}
}
func main(){
	ch := make (chan Message, 100)
   go goProcess(ch)

    rand.Seed(time.Now().UnixNano())

    i := 0
    for {
		if i > 10000 {break;}
        length := rand.Intn(20) + 1 // Random length between 1 and 20
        data := make([]byte, length)
        for j := range data {
            data[j] = byte(rand.Intn(256)) // Fill with random bytes
        }

        msg := Message{
            seq:  i,
            data: data,
        }
        ch <- msg
        i++
		if i % 100 == 0 {
			time.Sleep(1 * time.Second)
		}
    }
	time.Sleep(5 * time.Second)
}
