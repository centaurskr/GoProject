package main

import (
	"fmt"
	"runtime"
	"strings"
	"centolib/api"
)
func getLineInfor(full bool) (string, int, string){
    pc, file, line, ok := runtime.Caller(1)
    if ok == false{
        return "?", 0, "?"
    }
	
	if full == false {
		strs := strings.Split(file, "/")
		file = strs[len(strs)-1]
	}

    fn := runtime.FuncForPC(pc)
    if fn == nil {
        return file, line, "?"
    }

    return file, line, fn.Name()
}
func functionA() {
	file, line, fname:= getLineInfor(false)
	fmt.Printf("Local, false : %s-%s-%d\n", file, fname, line)
}
func main() {
	functionA()
	file, line, fname := getLineInfor(true)
	fmt.Printf("Local, true : %s-%s-%d\n", file, fname, line)
	file, line, fname = api.GetLineInfor(false)
	fmt.Printf("API, false : %s-%s-%d\n", file, fname, line)
	file, line, fname = api.GetLineInfor(true)
	fmt.Printf("API, true : %s-%s-%d\n", file, fname, line)
}
