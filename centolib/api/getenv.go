///
/// @brief   KEY=VALUE env file을 읽는 Function
/// @file    getenv.go
/// @date    2024. 09. 20. (금) 18:13:58 KST
/// @author  Cento
///
package api

import (
	"os"
	"bufio"
	"strings"
	"errors"
)

func GetEnvValue(name string, inkey string) (string, error) {
	
	file , err := os.Open(name)
	defer file.Close()
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		line  = strings.Split(line, "#")[0]
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2{
			continue
		}
		key := strings.TrimSpace(parts[0])
		if key != inkey {
			continue
		}else{
			value := strings.TrimSpace(parts[1])
			return value, nil
		}
	}
	errstr := name + " 에 " + inkey + "가 없습니다."
	return "", errors.New(errstr)
}
///
/// PLATFORM_ENV에서 환경 변수 읽기 
///
func GetPlatformEnv(name string)(string, error){
	envfile := os.Getenv("PLATFORM_ENV")
	if envfile 	== ""{
		return "", errors.New("환경변수에 PLATFORM_ENV가 없습니다")		
	}
	val, err := GetEnvValue(envfile, name)
	if err != nil{
		return "", errors.New(envfile + "에 "+ name + "가(이) 없습니다.")
	}
	return val, nil
}
