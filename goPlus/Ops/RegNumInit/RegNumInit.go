//주문접수번호 Mapfile을 만드는 Process
// 접수번호       :  yymmdd+00+00000000000  (11자리-100억)
//                   (yymmdd + line no + Seq)
//MAP_FILE_NAME은  ${ENV_HOME}/Platform.env에 REGNUM_MAP 사용
// ┌───────┬───────────┐
// │ INDEX │ 접수번호  │
// ├───────┼───────────┤
// │  int  │  uint64   │
// └───────┴───────────┘
//
package main
import (
	"fmt"
	"os"
	Mb "centolib/matchingboard"
	Log "centolib/logout"
	"centolib/api"
	"strconv"
	"unsafe"
	"time"
)
// Using은 프로세스 사용법을 로그에 출력한다.
func Using(){
	Log.Error("❗㉽ 접수번호 Map file 생성 Programm")
	Log.Error("Using) ")
	Log.Error("\t강제 생성 🡆 %s force", os.Args[0])
	Log.Error("\t초기 생성 🡆 %s", os.Args[0])
}
func main(){
	//Log.InitLogger("", Log.ALL_MODE);defer Log.Close()	
	Log.InitLogger("", Log.CONSOLE_MODE);defer Log.Close()	
	if len(os.Args) > 2{
		Using()
		os.Exit(0)
	}
	if len(os.Args) == 2 && os.Args[1] != "force"{
		Using()
		os.Exit(0)
	}

	env := os.Getenv("PLATFORM_ENV")
	if env == "" {
		Log.Error("PLATFORM_ENV이 없습니다.")	
		os.Exit(0)
	}
	if len(os.Args) == 1{
		file, err := os.Open(env)
		if err == nil{
			Log.Error("%s는 이미 존재합니다", env)
			Using()
			os.Exit(0)
		}
		defer file.Close()
	}

	regmap, err := api.GetEnvValue(env, "REGNUM_MAP")
	if err != nil {
		Log.Error("%s에 REGNUM_MAP 변수가 없습니다. :%v", env, err)
		os.Exit(0)
	}
	Log.Infor("주문접수번호 Map File : %s, Max Count : %d", 
			regmap, Mb.MAX_LINE_COUNT)
	size := api.SizeOf(Mb.MapHeader{}) + api.SizeOf(Mb.RegNumber{}) * int(Mb.MAX_LINE_COUNT)
	pMap,err := api.CreateMapfile(regmap, size)
	if err != nil{
		Log.Error("CreateMapFile(%s):%v", err)
		os.Exit(0)
	}
	defer api.UnMapfile(pMap)
	now := time.Now()

	// Header
	hd := (*Mb.MapHeader)(unsafe.Pointer(&pMap[0]))
	nowStr := fmt.Sprintf("%v", now)
	copy(hd.Create[:], []byte(nowStr))
	hd.Buffer = int(Mb.MAX_LINE_COUNT)
	hd.Count  = int(Mb.MAX_LINE_COUNT)
	
	hdsize := api.SizeOf(Mb.MapHeader{})
	size = api.SizeOf(Mb.RegNumber{})
	for i:= 0; i < Mb.MAX_LINE_COUNT; i++{
		seqstr := fmt.Sprintf("%s%02d00000000000", now.Format("060102"), i)
		seqInt, _ := strconv.ParseUint(seqstr, 10, 64)
		pRegNum := (*Mb.RegNumber)( unsafe.Add(unsafe.Pointer(&pMap[0]), uintptr(hdsize + size * i)))
			
		pRegNum.Index  = i
		pRegNum.RegNum = seqInt
		fmt.Printf("Index [%3d] RegNum[%d]\n", 
			pRegNum.Index,
			pRegNum.RegNum)
	}

}
