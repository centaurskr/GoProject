// JSON String을 Structure로 변환
// 
package main
import (
	"fmt"
	"encoding/json"
	"log"
	"os"
	"reflect"
)

//
// json:"first" json:"FIRST" -> 중복 선언 안됨
//
type person struct {
	First string `json:"first"`
	Last  string `json:"last"`
}

func main() {
	js :=`[{"first":"neil", "last":"reduman"},{"first":"pro", "last":"tangonist"}, {"First":"cento", "last":"master"}]` 

	// JSON string -> people로 변화( Unmarshal() 사용 )
	bs := []byte(js)
	fmt.Printf("%T\n", js);
	fmt.Printf("%T\n", bs);

	people := []person{};

	err := json.Unmarshal(bs, &people)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(people)

	fmt.Printf("people type : %s, js type : %s\n", reflect.TypeOf(people), reflect.TypeOf(js))

	// people --> JSON을 변환 ( Encode() 사용 )
	// os.Stdout을 string stream으로 변환 방법 생각해 봐야함.
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(people)
}
