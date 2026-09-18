///
/// @brief   test Memory mapped file
/// @file    tmmpa.go
/// @date    2024. 10. 08. (화) 17:17:04 KST
/// @author  Cento
///
package main
import (
	"fmt"
	"os"
	"encoding/binary"
	"golang.org/x/exp/mmap"
)
type DATA struct {
	key1 int32
	key2 int32
}

func createMapfile(filename string, size int) (*mmap.ReaderAt, error){
	_, err :=os.Stat(filename)
	if err == nil{
		return nil, fmt.Errorf("File Exist")
	}
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	if err := f.Truncate(int64(size)); err != nil{
		return nil, err
	}
	f.Close()
	mappoint , err := mmap.Open(filename)
	if err != nil {
		return nil, err
	}
	return mappoint, nil
}
func attachMapfile(filename string)(*mmap.ReaderAt, error){
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	f.Close()
	mappoint, err := mmap.Open(filename)
	if err != nil {
		return nil, err
	}
	return mappoint, nil
}
func main(){
	var mappoint *mmap.ReaderAt
	var err error
	var datas []DATA
	for i:= 0; i < 5; i++{
		data := DATA{
			key1 : int32(i + 1),
			key2 : int32(i + 2),
		}
		datas = append(datas, data)
	}


	mappoint, err = createMapfile("MAPTEST.MAP", 1024)
	if err != nil {
		mappoint, err = attachMapfile("MAPTEST.MAP")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Attach OK")
	}else{
		fmt.Println("create OK")
	}
	// Size := int64(len(orders)) * int64(binary.Size(Order{}))
	defer mappoint.Close()

	datasize := binary.Size(DATA{})
	fmt.Println("data size : ", datasize)
	for index, data := range datas{
		fmt.Println(index, data)
	}

	/// Write Memory mapped file
	for index, data := range datas{
		offset := datasize * index
		mdata := (*DATA)(mappoint.Data()[offset])
		mdata = data
		fmt.Println(mdata)
	}
}

