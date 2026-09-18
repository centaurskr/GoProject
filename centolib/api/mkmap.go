///
/// @brief   Memory mapped file lib
/// @file    mkmap.go
/// @date    2024. 10. 11. (금) 10:05:24 KST
/// @author  Cento
///
package api
import (
	"os"
	"syscall"
)
////////////////////////////////////////////////////////////////////////////////
/// 강제로 Map file을 만든다
/// @fn      func CreateMapfile(fname string, size int)([]byte, error)
/// @param  fname : map file name 
/// @param  size  : map file size 
/// @return []byte : memory mapped file pointer, error
////////////////////////////////////////////////////////////////////////////////
func CreateMapfile(fname string, size int)([]byte, error){
	file , err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}
	err = file.Truncate(int64(size))
	if err != nil {
		file.Close()
		return nil, err
	}
	mappoint, err := syscall.Mmap(int(file.Fd()), 0, size, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		file.Close()
		return nil, err
	}
	file.Close()

	return mappoint, nil
}

////////////////////////////////////////////////////////////////////////////////
/// Attach Memory mapped file 
/// @fn     : func AttachMapfile(fname string)(unsafe.Pointer, int, error)
/// @param  fname : map file name 
/// @return []byte : memory mapped file pointer
/// @return int    : mapfile size
/// @return err    : error
////////////////////////////////////////////////////////////////////////////////
func AttachMapfile(fname string)([]byte, int, error){
	file , err := os.OpenFile(fname, os.O_RDWR, 0666)
	if err != nil {
		return nil, -1, err
	}
	finfor, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, -1, err
	}
	fsize := finfor.Size()
	mappoint, err := syscall.Mmap(int(file.Fd()), 0, int(fsize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		file.Close()
		return nil, -1, err
	}
	file.Close()

	return mappoint, int(fsize), nil
}

////////////////////////////////////////////////////////////////////////////////
/// Unmap
/// @fn      func UnMapfile(data []byte){
/// @param  
/// @return
////////////////////////////////////////////////////////////////////////////////
func UnMapfile(data []byte){
	syscall.Munmap(data)
}
