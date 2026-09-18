//
// Process의 중복실행 방지를 위한 File Lock
// 
package api
import (
	"os"
	"syscall"
)
//
// 중복실행방지를 위한 Check Function
// -. return value 
// ture : 실행중임
// false : 실행중 아님 (Lock을 건다)
// -. name
//  현Host에서 유일한 이름으로.....
//
func IsRunning(name string)(bool, error){
	lockname := "/tmp/" + name + ".lock"
	file, err := os.OpenFile(lockname, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil{
		return true, err
	}
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil{
		return true, err
	}
	return false, nil
}
