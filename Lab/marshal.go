package main
import (
	"fmt"
	"log"
	"encoding/json"
)
type person struct {
	First string `json:"FIRST"`
	Last  string `json:"LAST"`
}

func main() {
	p1 := person{First: "neil", Last: "reduman"}
	p2 := person{First: "pro", Last: "tagonist"}

	people := []person{p1, p2}

	// struct를 marshalling
	b1, err := json.Marshal(people)
	if err != nil {
		log.Fatalln(err) // kill the program
	}
	fmt.Println(b1)
	fmt.Println(string(b1))
}
