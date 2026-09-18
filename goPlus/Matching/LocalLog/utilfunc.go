// Matching 프로세스에서 사용하는 유틸리티 함수 모음.
package main

import (
	"centolib/api"
	Mb "centolib/matchingboard"
	"strings"
	"unsafe"
	"os"
	"fmt"
)

// IsExistOrderDir은 ORDPATH 환경변수 경로 아래에 symbol 디렉토리가 있는지 검사한다.
// symbol은 "BTC.USDT" 형식으로 입력한다.
// 디렉토리가 존재하면 true, 없으면 false를 반환한다.
func IsExistOrderDir(symbol string)(bool, error){
	path, err := api.GetPlatformEnv("ORDPATH")
	if err != nil{return false, err;}
	path = fmt.Sprintf("%s/%s", strings.TrimSuffix(path, "/") , symbol)
    _, err = os.Stat(path)
	if err != nil { return false, nil;}
	return true, nil
}
