// Log를 읽기 좋게....
//
package main
import (
	"fmt"
	"encoding/json"
	"bufio"
	"bytes"
	"os"
)
func main(){
	if len(os.Args) != 2 {
		fmt.Printf("Json file을 보는 Program\n")
		os.Exit(0)
	}
	file, err := os.Open(os.Args[1])
	if err != nil{
		fmt.Printf("Error : Open(%s) : %v\n", os.Args[1], err)
		os.Exit(0)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var outBuff bytes.Buffer
		line := scanner.Text()
		err := json.Indent(&outBuff, []byte(line), "", "\t")
		if err != nil{
			fmt.Printf("Error : %s\n%s\n", err, line)
			continue
		}
		fmt.Println(outBuff.String())
	}
}
