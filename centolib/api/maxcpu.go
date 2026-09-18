///
/// @brief  Set the maximum number of CPUs to use
/// @file   maxcpu.go
/// @date    2025. 01. 24. (금) 16:25:00 KST
/// @author  Cento
///
package api
import "runtime"

func UsingMaxProcess(){
	runtime.GOMAXPROCS(runtime.NumCPU())
}
