//
// 시간 관련 Library
//
package api
import (
	"time"
	"strconv"
)
// 20060102-150405.000
// yyyymmdd -- 숫자
func GetNumYyyymmdd()int{
	now := time.Now()
	str := now.Format("20060102")	
	tint, _ := strconv.Atoi(str)
	return tint
}
// yyyymmdd -- String
func GetStrYyyymmdd()string{
	now := time.Now()
	str := now.Format("20060102")	
	return str
}

// yymmdd -- 숫자
func GetNumYymmdd()int{
	now := time.Now()
	str := now.Format("060102")	
	tint, _ := strconv.Atoi(str)
	return tint
}
// yymmdd -- String
func GetStrYymmdd()string{
	now := time.Now()
	str := now.Format("060102")	
	return str
}
