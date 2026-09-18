package main
import (
	"centolib/api"
	Mb "centolib/matchingboard"
	"fmt"
	"os"
	"unsafe"
	"strings"
)
// Order Count
func CheckOrderCount(path, name string)bool{
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
func CleanOrderMapFile(path, symbol string)error{

	// ./Work가 없으면 생성하고
	// MEDO.* MESU.* file을 ./Work/ 아래로 모두 이동
	workDir := fmt.Sprintf("%s/%s/Work", path,symbol)
    os.Mkdir(workDir, 0755)

	readDir := fmt.Sprintf("%s/%s", path, symbol)
    dirs , err := os.ReadDir(readDir)
    if err != nil{return err;}
    for _, file := range dirs{
        if strings.HasPrefix(file.Name(), "MEDO.") {
			fromPath := fmt.Sprintf("%s/%s", readDir, file.Name())
			newPath := fmt.Sprintf("%s/%s", workDir, file.Name())
            err = os.Rename(fromPath, newPath)
            if err != nil{ return err;}
        }else if strings.HasPrefix(file.Name(), "MESU."){
			fromPath := fmt.Sprintf("%s/%s", readDir, file.Name())
			newPath := fmt.Sprintf("%s/%s", workDir, file.Name())
            err = os.Rename(fromPath, newPath)
            if err != nil{ return err; }
        }
    }
	readDir = workDir
	dirs, err = os.ReadDir(readDir) // 작업 DIR : "COIN"/Work
	if err != nil {return err;}

	medoIdx := 0
	mesuIdx := 0

	// 주문이 있는 Map file을 새로운 SEQ를 부여(0부터 시작)
	// "../"로 이동한다
	for _, file := range dirs{
		if strings.HasPrefix(file.Name(), "MEDO."){
			if CheckOrderCount(readDir, file.Name()) == true{
				fromPath := fmt.Sprintf("%s/%s", readDir, file.Name())
				newPath := fmt.Sprintf("%s/%s/MEDO.%04d", path, symbol, medoIdx)
				err := os.Rename(fromPath, newPath)
				if err != nil {return err;}
				medoIdx ++
			}
		}
		if strings.HasPrefix(file.Name(), "MESU."){
			if CheckOrderCount(workDir, file.Name()) == true{
				fromPath := fmt.Sprintf("%s/%s", readDir, file.Name())
				newPath := fmt.Sprintf("%s/%s/MESU.%04d", path, symbol, medoIdx)
				err := os.Rename(fromPath, newPath)
				if err != nil {return err;}
				mesuIdx ++
			}

        }
	}

	// 3개 더 만든다
	workDir = fmt.Sprintf("%s/%s", path, symbol)
	for i:= medoIdx; i < (medoIdx + 3) ; i++{
		Mb.CreateHoga(workDir, symbol, Mb.MEDO, i)
	}
	for i:= mesuIdx; i < (mesuIdx + 3) ; i++{
		Mb.CreateHoga(workDir, symbol, Mb.MESU, i)
	}

	// ./Work에 있는 나머지 File을 delete한다
	workDir = fmt.Sprintf("%s/%s/Work", path, symbol)
	dirs, _ = os.ReadDir(workDir)
	for _, file := range dirs{
		delName := fmt.Sprintf("%s/%s", workDir, file.Name())
		os.RemoveAll(delName)
	}

	///// Mater Infor 변경
	mesuIdx = mesuIdx + 3
	medoIdx = medoIdx + 3
	mstName := fmt.Sprintf("%s/%s/%s.MASTER", path, symbol, symbol)
    pMap, _, err := api.AttachMapfile(mstName)
    if err != nil{ return err; }
    defer api.UnMapfile(pMap)

    pMaster := (*Mb.OrderMasterInfor)(unsafe.Pointer(&pMap[0]))
	pMaster.MedoNameIdx = medoIdx
	pMaster.MesuNameIdx = mesuIdx

	return nil
}

func Handler(memcoin *Mb.MEMCOIN, path, symbol string)(*Mb.MEMCOIN, bool){
	memcoin.UnLoadMEMCOIN()
	err := CleanOrderMapFile(path, symbol)
	if err != nil{
		return nil, false
	}
	newmemcoin := &Mb.MEMCOIN{}
	newmemcoin.LoadMEMCOIN(path, symbol)

	return newmemcoin, true
}
func main(){
    if len(os.Args) != 3{
        fmt.Printf("Clean 주문file & Reload  Test\n")
        fmt.Printf("ex) gocleanmap path BTC.USDT\n")
        fmt.Printf(" path/BTC.USDT에 있는 MEDO.*, MESU.*를 읽는다\n")
        os.Exit(0)
    }
	path := os.Args[1]
	symbol := os.Args[2]

	fmt.Printf("PATH : %s, Symbol : %s\n", path, symbol)

	memcoin := &Mb.MEMCOIN{}
	memcoin.LoadMEMCOIN(path, symbol)

	memcoin.MstInfor.Dump(1)
	
	memcoin, flag := Handler(memcoin, path, symbol)

	if flag == false{
		fmt.Printf("Error. Handler\n")
		os.Exit(0)
	}
	fmt.Printf("\nAFTER -----------------------\n")
	memcoin.MstInfor.Dump(1)
}
