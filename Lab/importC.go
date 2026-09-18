///
/// @brief  : import C TEST
/// @file   : importC.go
/// @date   : 2024. 04. 25. (목) 10:18:38 KST
/// @author : Cento
///
package main

/*
#cgo CFLAGS: -Werror -I/usr/include
#cgo LDFLAGS: -L/usr/lib64 -L/usr/lib/x86_64-linux-gun -ldl
#include <zmq.h>
#include <stdio.h>
#include <stdlib.h>
char* print_env(void){
	printf("C> PAHT : '%s'\n", getenv("PATH"));
	return getenv("PATH");
}
*/
import "C"

import "fmt"
import "os"

func main(){
	user:= C.getenv(C.CString("USER"))
	fmt.Println(C.GoString(user))
	fmt.Println(os.Getenv("USER"))
	env:= C.print_env()
	fmt.Println(C.GoString(env))
}
