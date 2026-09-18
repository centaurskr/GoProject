///
/// @brief  Netlog test gofunction
/// @file    gofun.go
/// @date    2025. 12. 29. (월) 11:31:09 KST
/// @author  Cento
///
package main

import (
	"fmt"
	"sync"
	"time"
	NetLog "centolib/netlog"
)

func goFunction(wg *sync.WaitGroup, ch chan int){
	NetLog.Debug("Hello I am go function")
	term := time.After(100 * time.Second)
	for {
		select {
			case n := <-ch :
				NetLog.Infor("Integer : %d\n", n)
				fmt.Printf("Integer : %d\n", n)
				time.Sleep(time.Second)
			case <-term :
				NetLog.Infor("Terminated!")
				fmt.Printf("Terminated!")
				wg.Done()
				return
		}
	}
}

