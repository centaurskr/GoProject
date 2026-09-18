///
/// @brief  : Test centolib/bindecimal 
/// @file   : bindeclib.go
/// @date   : 2024. 10. 08. (화) 15:17:26 KST
/// @author : Cento
///
package main

import (
	"fmt"
	Bdc "centolib/bindecimal"
)
func main(){

	d1,_ := Bdc.NewFromString("988888888888.123123123123")
    fmt.Printf("Return NewFromString(988888888888.123123123123) : %v\n%s\n",  d1, d1.String())
	d1.SetFromString("888.123123123123")
    fmt.Printf("Return SetFromString(888.123123123123) : %v\n%s\n",  d1, d1.String())

	d1.SetFromString("1.1")
	d2,_ := d1.Add("1.2", Bdc.NOROUND, 0)
	fmt.Printf("d1 = 1.1, d1.Add(\"1.2\")= %s\n", d2.String())

}

