// 신규 코인등록시 주문 관련 Map file을 만드는 Process
// 1. mkdir $(ORDPATH)/${coin}
// 2. make $(ORDPATH)/${coin}/${coin}.MASTER
// 3. make $(ORDPATH)/${coin}/MEDO.00 ~ 03
// 4. make $(ORDPATH)/${coin}/MESU.00 ~ 03
package main
import (
	Mb "centolib/matchingboard"
	Log "centolib/logout"
	"centolib/api"
	"os"
	"unsafe"

	"fmt"
	"strings"

)

func main(){
	Log.InitLogger("", Log.ALL_MODE);defer Log.Close()

	if len(os.Args) != 2{
		Log.Error("신규 Coin 거래를 위한 주문정보 초기 설정 PGM")
		Log.Error("Using)%s <coin>", os.Args[0])
		Log.Error("     <coin> :거래 코인이름(ex : BTC.KRW)")
		os.Exit(1)
	}
	envfile := os.Getenv("PLATFORM_ENV")
	if envfile == ""{
		Log.Error("환경변수 PLATFORM_ENV가 없습니다")
		os.Exit(1)
	}
	path,err := api.GetEnvValue(envfile, "ORDPATH")
	if err != nil {
		Log.Error("%s에 ORDPATH가 없습니다", envfile)
		os.Exit(2)
	}
	name := os.Args[1]
	Log.Infor("ORDPATH : %s", path)

	dbyaml,err := api.GetEnvValue(envfile, "DBCONFIG_YAML")
	if err != nil{
		Log.Error("%s에 DBCONFIG_YAML이 없습니다", envfile)
		os.Exit(3)
	}
	dbctx, err := api.ConnectDb(dbyaml)
	if err != nil{
		Log.Error("ConnectDb(%s) : %v", dbyaml, err)
		os.Exit(4)
	}
	defer dbctx.Close()
	
	cinfor, err := Mb.GetCoinInfor(dbctx, name)	
	if err != nil{
		Log.Error("GetCoinInfor()\n%v", err)
		os.Exit(5)
	}
	// Path 검사
	tpath := fmt.Sprintf("%s/%s", strings.TrimSuffix(path, "/") , name)
	fmt.Printf("main path : %q\n", tpath)
	_, err = os.Stat(tpath)
	if err != nil{
		Log.Warn("%s는 없습니다. 생성합니다 mode(rwxrw_rw-).", tpath)
		err := os.MkdirAll(tpath, 0777)
		if err != nil{
			Log.Error("%s 생성실패\n\t%v", tpath, err)
			os.Exit(6) 
		}
	}
	
	// Make Order Master Information
	err = Mb.CreateOrderMasterInfor(path, cinfor)
	if err != nil{
		Log.Error("CreateOrdermasterInfor(%s, %s)\n\t%v", path, name, err)
		os.Exit(7) 
	}

	// get
	pmap,err := Mb.GetOrderMasterInfor(path, name)
	if err != nil{
		Log.Error("GetOrderMasterInfor(%s, %s)\n\t%v", path, name, err)
		os.Exit(8)
	}
	defer api.UnMapfile(pmap)

	// Path 재설정
	path = path + "/"+ name
	mst := (*Mb.OrderMasterInfor)(unsafe.Pointer(&pmap[0]))

	// Master Name index update
	mst.MedoNameIdx = Mb.ORDER_COUNT
	mst.MesuNameIdx = Mb.ORDER_COUNT

	mst.Dump(0)
	Log.Infor("COIN Default Path : %s", path)
	Log.Infor("Symbol[%s] ItemCode[%d] CoinCoide[%d] MkCode[%d]", 
		string(cinfor.Symbol[:]),
		cinfor.ItemCode, 		
		cinfor.CoinCode, 		
		cinfor.MkCode) 		

	Log.Infor("Master Infor")
	Log.Infor("\tOrdNum[%d] BfPrice[%s] BfTime[%d] MedoNameIdx[%d] MesuNameIdx[%d]", 
	mst.OrdNum, mst.BfPrice.String(), mst.BfTime, 
	mst.MedoNameIdx,mst.MesuNameIdx)

	// 매도 HOGA 생성
	for i := 0; i < Mb.ORDER_COUNT; i ++{
		err:=Mb.CreateHoga(path, name, Mb.MEDO, i)
		if err != nil{
			Log.Error("CreateHoga(%s, %s, MEDO, %d):\n\t%v", path, name, i, err)
			os.Exit(9)
		}else{
			Log.Infor("CreateHoga(%s, %s, MEDO, %d)", path, name, i)
		}
	}
	// 매수 HOGA 생성
	for i := 0; i < Mb.ORDER_COUNT; i ++{
		err:=Mb.CreateHoga(path, name, Mb.MESU, i)
		if err != nil{
			Log.Error("CreateHoga(%s, %s, MESU, %d):\n\t%v", path, name, i, err)
			os.Exit(10)
		}else{
			Log.Infor("CreateHoga(%s, %s, MESU, %d)", path, name, i)
		}
	}
	Log.Infor("%s 주문 초기구조 구축완료", name)
	os.Exit(0)
}
