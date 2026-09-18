package main
import (
	"fmt"
	"strings"
	"centolib/api"
	Mb "centolib/matchingboard"
	"unsafe"
	"path/filepath"
	"os"
)

// Order Count
func CheckOrderMapfile(path, name string)bool{
	fullName := path + "/"+ name
	pMap, _, err := api.AttachMapfile(fullName)
	if err != nil{
		fmt.Printf("Fail to attach(%s) : %v\n", fullName, err)
		return false
	}
	defer api.UnMapfile(pMap)
	pHd := (*Mb.OrderHeader)(unsafe.Pointer(&pMap[0]))
	if pHd.OrderCount == 0{
		return false
	}
	return true
}
func CleanOrderFile(pMst *Mb.OrderMasterInfor, path, symbol string)error{
	var medoCnt, mesuCnt int

	workDir := fmt.Sprintf("%s/%s/Work", path, symbol)
	err := os.Chdir(workDir)
	if err != nil{return err;}

	dirs, err := os.ReadDir(".")
	if err != nil{return err;}
	medoCnt = 0
	mesuCnt = 0

	for _, file := range dirs{
		if strings.HasPrefix(file.Name(), "MEDO."){
			if CheckOrderMapfile(".", file.Name()) == true {
				newPath := fmt.Sprintf("../MEDO.%04d", medoCnt)
	fmt.Printf("Move %s to %s\n", file.Name(), newPath)
				err := os.Rename(file.Name(), newPath)
				if err != nil{return err;}
				medoCnt ++
			}
		}

		if strings.HasPrefix(file.Name(), "MESU."){
			if CheckOrderMapfile(".", file.Name()) == true {
				newPath := fmt.Sprintf("../MESU.%04d", mesuCnt)
	fmt.Printf("Move %s to %s\n", file.Name(), newPath)
				err := os.Rename(file.Name(), newPath)
				if err != nil{return err;}
				mesuCnt ++
			}
		}
	}
	fmt.Printf("MEDO Count : %d, MESU Count : %d\n", medoCnt, mesuCnt)
	// 기본 3개
	if medoCnt < 3 {
		for i := medoCnt; i < 3; i++ {
			Mb.CreateHoga("..", symbol, Mb.MEDO, i)
		}
		medoCnt = 3
	}
	if mesuCnt < 3 {
		for i := mesuCnt; i < 3; i++ {
			Mb.CreateHoga("..", symbol, Mb.MESU, i)
		}
		mesuCnt = 3
	}
	fmt.Printf("MEDO Count : %d, MESU Count : %d\n", medoCnt, mesuCnt)
	pMst.MedoNameIdx = medoCnt
	pMst.MesuNameIdx = mesuCnt

	// 나머지 File 지우기
	dirs, _ = os.ReadDir(".")
	for _, file := range dirs {
		path := filepath.Join(".", file.Name())
		os.RemoveAll(path)
	}
	return  nil
}
// 
// 1. Make ./Work DIR
// 2. Move MEDO.* & MESU.* to ./Work/
// 
func MoveOrderFile(path, symbol string)error{
	saveDir, err := os.Getwd() // Get working directory
	if err != nil{ return err; }
	workDir := fmt.Sprintf("%s/%s", path, symbol)	
	err = os.Chdir(workDir)
	// fmt.Printf("Change Dir to : %s\n", workDir)

	if err != nil{ return err; }
	os.Mkdir("./Work", 0755)
	dirs , err := os.ReadDir(".")
	if err != nil{return err;}
	for _, file := range dirs{
		if strings.HasPrefix(file.Name(), "MEDO.") {
			newPath := "./Work/" + file.Name()
        	err = os.Rename(file.Name(), newPath)
			if err != nil{ return err;}
		}else if strings.HasPrefix(file.Name(), "MESU."){
			newPath := "./Work/" + file.Name()
        	err = os.Rename(file.Name(), newPath)
			if err != nil{ return err; }
		}
	}
	os.Chdir(saveDir)
	//fmt.Printf("Change Dir to %s\n", saveDir)
	return nil
}
func main(){
	if len(os.Args) != 2{
		fmt.Printf("주문 File 읽기 Test\n")
		fmt.Printf("ex) goReadOrd BTC.USDT\n")
		fmt.Printf(" $(ORDPATH)/BTC.USDT에 있는 MEDO.*, MESU.*를 읽는다\n")
		os.Exit(0)
	}
/*
	path := os.Getenv("ORDPATH")
	if path == "" {
		fmt.Printf("환경 변수에 ORDPATH가 없음\n")
		os.Exit(0)
	}
*/
	path := "."

	// Master Infor
	mstName := path + "/" + os.Args[1] + "/" + os.Args[1] + ".MASTER"	
	pMap, _, err := api.AttachMapfile(mstName)
	if err != nil{
		fmt.Printf("ERROR : Master(%s) : %v\n", mstName, err)
		os.Exit(0)
	}
	defer api.UnMapfile(pMap)

	pMaster := (*Mb.OrderMasterInfor)(unsafe.Pointer(&pMap[0]))
	pMaster.Dump(1)

	// [COIN]/Work로 모두 이동
	err = MoveOrderFile(path, os.Args[1])
	if err != nil{
		fmt.Printf("MoveOrderFile(%s,%s) : %v\n", path, os.Args[1], err)
		return
	}

	// [COIN]/Work에 있는 MEDO. MESU. 중 주문이 있는 것만 [COIN]으로 이동
	err = CleanOrderFile(pMaster, path, os.Args[1])	
	if err != nil{
		fmt.Printf("MoveOrderFile(%s,%s) : %v\n", path, os.Args[1], err)
		return
	}
	pMaster.Dump(1)
}
