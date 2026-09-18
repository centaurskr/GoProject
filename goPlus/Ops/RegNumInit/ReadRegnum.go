
//주문접수번호 Mapfile 조회  Process
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
	"unsafe"
)
func main(){
	Log.InitLogger("", Log.CONSOLE_MODE);defer Log.Close()	

	env := os.Getenv("PLATFORM_ENV")
	if env == "" {
		Log.Error("PLATFORM_ENV이 없습니다.")	
		os.Exit(0)
	}

	regmap, err := api.GetEnvValue(env, "REGNUM_MAP")
	if err != nil {
		Log.Error("%s에 REGNUM_MAP 변수가 없습니다. :%v", env, err)
		os.Exit(0)
	}
	pMap,_, err := api.AttachMapfile(regmap)
	if err != nil{
		Log.Error("AttachMapfile(%s) : %v", regmap, err)
		os.Exit(0)
	}
	defer api.UnMapfile(pMap)
	
	hd := (*Mb.MapHeader)(unsafe.Pointer(&pMap[0]))

	for i := 0; i < hd.Count; i++{
		offset := api.SizeOf(Mb.MapHeader{}) + api.SizeOf(Mb.RegNumber{}) * i
		reg := (*Mb.RegNumber)(unsafe.Pointer(&pMap[offset]))
/*
		reg := (*Mb.RegNumber)(unsafe.Add( unsafe.Pointer(&pMap[0]),
						uintptr(api.SizeOf(Mb.MapHeader{}) + 
						api.SizeOf(Mb.RegNumber{}) * i)))
*/
		fmt.Printf("Index [%3d] 접수번호[%d]\n", reg.Index, reg.RegNum)
	}

	fmt.Printf("==== 접수번호 Dump =====\n")
	fmt.Printf("Map File name : %s\n", regmap)
	fmt.Printf("생성 일 : %s, Buffer[%d] Count[%d]\n", 
		string(hd.Create[:]), hd.Buffer, hd.Count)
}
