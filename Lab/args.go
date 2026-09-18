package main
import ( 
	"fmt"
	"os"
)
func main(){

	fmt.Printf("Argument Len : %d\n", len(os.Args))
	for i, str:= range os.Args{
		fmt.Printf("[%2d] : %s\n",i, str)
	}
	
	var argstr string
	for i := 1; i < len(os.Args); i++{
		argstr = argstr + " " + os.Args[i]	
	}
	fmt.Printf("Arg String : [%s]\n", argstr)

}
