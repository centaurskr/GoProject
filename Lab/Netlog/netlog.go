///
/// @brief  
/// @file    netlogtest.go
/// @date    2025. 12. 22. (월) 11:31:52 KST
/// @author  Cento
///

package main

import(
	NetLog "centolib/netlog"
	"fmt"
	"path/filepath"	
	"os"
	"sync"
)
func main(){
	var wg sync.WaitGroup

	err := NetLog.InitNetLogout(filepath.Base(os.Args[0]), "")
	if err != nil {
		fmt.Printf("%s : %v\n", os.Args[0], err)
		os.Exit(0)
	}
	defer NetLog.Close()


	NetLog.Infor("INFO === 시작 == %s", os.Args[0])
	NetLog.Debug("Debug === WARN == %s", os.Args[0])
	NetLog.Error("Error === DEBUG == %s", os.Args[0])
	ch := make(chan int)

	wg.Add(1)
	go goFunction(&wg, ch)


	for i:= 0 ; i < 1000; i++{
		ch <- i
		fmt.Printf("INDEX : %d\n", i)
	}
	close(ch)
	wg.Wait()
}




