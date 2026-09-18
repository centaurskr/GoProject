//
// 10 분단위 환율 
// kr.investing.com에서 환율을 가져온 후 10단위로 환율을 저장한다
// 2분에서 최대 5분 지연 환율
//
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"centolib/api"
	//zmq "github.com/pebbe/zmq4"
	"os"
	"strconv"
)

func getExchangeRate(nation1, nation2 string) (string, string, error) {
	var timeStr, exrate string

	url := fmt.Sprintf("https://kr.investing.com/currencies/%s-%s", nation1, nation2)
	client := &http.Client{}
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil { return timeStr, exrate, err; }
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil { return timeStr, exrate, err; }
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil { return timeStr, exrate, err; }
	// Parse HTML and extract price (using naive approach for simplicity)
	priceMarker := "data-test=\"instrument-price-last\">"
	priceStart := strings.Index(string(body), priceMarker) + len(priceMarker)
	if priceStart == -1 { return timeStr, exrate, fmt.Errorf("Price start Err"); }

	priceEnd := strings.Index(string(body)[priceStart:], "</div>")
	priceStr := string(body)[priceStart : priceStart+priceEnd]
	exrate = strings.Replace(priceStr, ",", "", -1)

	if len(exrate) > 10 {return timeStr,exrate, fmt.Errorf("html Error");}

	timeMarker := "data-test=\"trading-time-label\">"
	timeStart := strings.Index(string(body), timeMarker) + len(timeMarker)
	if timeStart == -1 { return timeStr, exrate, fmt.Errorf("time data Not found");}
	timeEnd := strings.Index(string(body)[timeStart:], "</time>")

	//fmt.Printf("START[%d] END[%d]\n", timeStart, timeEnd)

	timeStr = string(body)[timeStart:timeStart+timeEnd]
	timeStr = strings.Replace(timeStr, ":", "", -1)
	timeStr = fmt.Sprintf("%d%s", api.GetNumYyyymmdd(), timeStr)
	return timeStr, exrate, nil
}
//////////////// Main ///////////////////
func main() {

var lastRate string
var updateTime int


	for{
		timeStr, exrate, err := getExchangeRate("usd", "krw")
		if err != nil{
			fmt.Printf("Error : %v\n", err)
			continue
		}
		if updateTime == 0{
			updateTime,_ = strconv.Atoi(timeStr)
			lastRate = exrate
			fmt.Printf("환율 : [%d]  %s\n", updateTime / 100, lastRate)
			continue
		}
		gettime , _ :=strconv.Atoi(timeStr)
		if  gettime  > updateTime {
			fmt.Printf("\nSave[%d] Rcv[%s]\t", updateTime, timeStr[:12])
			updateTime = gettime
			lastRate = exrate
			fmt.Printf("환율 : [%d]  %s  ", updateTime / 100, lastRate)
		}else{
			fmt.Printf(".")
			//fmt.Printf("\t[%d] Skip [%s][%s]\n", updateTime, timeStr, exrate)
		}
		time.Sleep(10 * time.Second)
	}
}

