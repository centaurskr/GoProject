//
// yaml file test
// rdyaml.go
// 2024. 12. 04. (수) 10:14:46 KST
// Author Cento
package main
import (
	"fmt"
	"os"
	"gopkg.in/yaml.v3"
)
type ProxyYaml struct {
	Comment         string `yaml:"comment"`
	ProxyType       string `yaml:"proxytype"`
	FrontEndPoint   string `yaml:"front"`
	BackEndPoint    string `yaml:"back"`
	ControlEndPoint string `yaml:"control"`
	CaptureEndPoint string `yaml:"capture"`
}
func main(){
	data, err := os.ReadFile("proxy.yaml")	
	if err != nil { fmt.Println(err);os.Exit(0)}

	var objects map[string]ProxyYaml
	err = yaml.Unmarshal(data, &objects)
	if err != nil {fmt.Println(err);os.Exit(0)}

	for _, obj := range objects{
		fmt.Printf(`
		Comment : %s
		ProxyType : %s
		FrontEndPoint: %s
		BackEndPoint: %s
		ControlEndPoint: %s
		CaptureEndPoint: %s

		`,
		obj.Comment         ,
		obj.ProxyType       ,
		obj.FrontEndPoint   ,
		obj.BackEndPoint    ,
		obj.ControlEndPoint ,
		obj.CaptureEndPoint )
	}
	var node yaml.Node
	err = yaml.Unmarshal(data, &node)
	if err != nil{
		fmt.Println(err)
		os.Exit(0)
	}
	
	var names []string	
	root := node.Content[0]
	for i := 0; i < len(root.Content) ; i+=2{
		names = append(names, root.Content[i].Value)
	}
	fmt.Printf("%v\n", names)
}
