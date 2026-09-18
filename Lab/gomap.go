///
/// @brief  : GOLANG map test
/// @file   : gomap.go
/// @date   : 2024. 04. 29. (월) 09:39:22 KST
/// @author : Cento
///
package main

import "fmt"

func main(){
	m:= make(map[string]string)
	m["name"]="John"
	m["country"] = "Korea"
	fmt.Println(m)

	m["city"]="Seoul"
	fmt.Println(m)

	fmt.Printf("No Key :%s / %T\n\n", m["language"], m["language"])


	s:= make(map[string]string)
	s["name"]=""
	fmt.Printf("name:%s / %T\n", s["name"], s["name"])
	fmt.Printf("country:%s / %T\n\n", s["country"], s["country"])

	v,ok := m["name"]
	fmt.Println(v, ok)

	v,ok = s["country"]
	fmt.Println(v, ok)
}

