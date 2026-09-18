package main

import (
	"fmt"
	"bufio"
	"os"
	"strings"
	"strconv"
	"slices"
)
func AddNewCoin(fileName, newCoin string) error{
	file, err := os.Open(fileName)
	if err != nil{return err; }
	var lines, coins []string
	var index, coinCount int
	scanner := bufio.NewScanner(file)
	coinCount = 9999999 // 초기값
	lineNum := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "ORD_"){
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2{
				list := strings.Split(strings.Trim(parts[1], " :"), ":")
				cnt := len(list)
				if coinCount > cnt {
					coinCount = cnt
					index = lineNum
				}
				coins = append(coins, list...)
			}
			lines = append(lines, line)
		}else{
			lines = append(lines, line)
		}
		lineNum ++
	}
	file.Close()

	// 이미 있는가?
	if exist := slices.Contains(coins, newCoin); exist{
		return fmt.Errorf("AddNewCoin() %s는 이미 등록된 코인입니다", newCoin)
	}

	// Line 마지막에 #이 있는가 검사
	parts := strings.SplitN(lines[index], "#", 2)
	if len(parts) < 2{
		lines[index] = strings.Trim(lines[index], " :") + ":"+newCoin
	}else{
		lines[index] = strings.Trim(parts[0] , " :") + ":" + newCoin + " #" + parts[1]
		fmt.Printf("Lines [%s]\n", lines[index])
	}

	os.Rename(fileName, fileName+".BACK")
	err = os.WriteFile(fileName, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil{
		os.Rename(fileName+".BACK", fileName)
		return err
	}

	return nil
}
func GetCoinListByIndex(fileName string, idx int)([]string, error){

	var coins []string

	file, err := os.Open(fileName)
	if err != nil{return nil, err; }
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan(){
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "ORD_"){
			str := strings.SplitN(line, "#", 2) // Line 중 # 제거
			parts := strings.SplitN(str[0], "=", 2)
			if len(parts) != 2{continue;}
			name := strings.SplitN(strings.Trim(parts[0], " "), "_", 2) // Index 구하기 
			num,_ := strconv.Atoi(name[1])
			if num != idx {continue;}
			lists :=strings.Split( strings.Trim(parts[1], " :"), ":")
			for _, item := range(lists){
				coins = append(coins, strings.Trim(item, " "))
			}
		}
	}

	if len(coins) == 0 {
		return nil, fmt.Errorf("ORD[%d]는 등록된 코인이 없습니다.",idx)
	}
	return coins, nil
}
func GetAllCoins(fileName string)([]string, error){
	var coins []string
	file, err := os.Open(fileName)
	if err != nil{
		return nil, fmt.Errorf("GetAllCoins() : %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan(){
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "ORD_"){continue;}
		str := strings.SplitN(line, "#", 2) // Line중 #
		parts := strings.SplitN(str[0], "=", 2) // ORD_??= 분리
		if len(parts) != 2{continue;}
		lists := strings.Split( strings.Trim(parts[1], " :"), ":") // Coin 분리
		for _, item := range(lists){
			coins = append(coins, strings.Trim(item, " "))
		}
	}
	if len(coins) == 0 {
		return nil, fmt.Errorf("등록된 코인이 없습니다.")
	}
	return coins, nil
}
func main(){
	fileName := "COIN.LIST"
	err := AddNewCoin(fileName, "ETH.USDT")
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}
	for i:= 1; i < 12; i++{
		coins, err := GetCoinListByIndex(fileName, i)
		if err != nil{
			fmt.Printf("GetCoinList() %v\n", err)
			continue
		}
		fmt.Printf("index [%d] coins %v\n", i, coins)
	}
	coins, err := GetAllCoins(fileName)
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}
	fmt.Printf("All Coins : %v", coins)
}
