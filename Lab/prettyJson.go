package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func main() {
	raw := []byte(`{"name":"Alice","age":25,"city":"London"}`)
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, raw, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("%T\n%v\n", prettyJSON, prettyJSON.String())

	var strData map[string]interface{}

	err = json.Unmarshal([]byte(prettyJSON.String()), &strData)
	if err != nil{
		fmt.Println("Error : ", err)
	}else{
		for name, value := range strData{
			fmt.Printf("%s : %v \n", name, value)
		}
	}
}

