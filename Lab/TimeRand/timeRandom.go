///
/// @brief  : 일력된 시작, 종료 시각 내 시간 문자를 
///           YYYYMMDDHHMMSS.MSEC 형태로 생성하는 프로그램
/// @file   : timeRandom.go
/// @date   : 2025. 11. 14. (금) 11:09:47 KST
/// @author : Cento
///
package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

func parseTime(s string) (time.Time, error) {
	// "YYYYMMDDHHMM" → Go 시간 파싱
	layout := "200601021504"
	return time.ParseInLocation(layout, s, time.Local)
}

func main() {
	if len(os.Args) != 4 {
		fmt.Printf("Usage: %s <START:YYYYMMDDHHMM> <END:YYYYMMDDHHMM> <COUNT>\n", os.Args[0])
		return
	}

	startStr := os.Args[1]
	endStr := os.Args[2]
	countStr := os.Args[3]

	startTime, err1 := parseTime(startStr)
	endTime, err2 := parseTime(endStr)
	if err1 != nil || err2 != nil {
		fmt.Println("Error: Cannot parse start/end time")
		return
	}

	count := 0
	fmt.Sscanf(countStr, "%d", &count)
	if count <= 0 {
		fmt.Println("Error: count must be > 0")
		return
	}

	if startTime.After(endTime) {
		fmt.Println("Error: start > end")
		return
	}

	rand.Seed(time.Now().UnixNano())

	diff := endTime.Sub(startTime).Seconds()
	if diff < 0 {
		fmt.Println("Time range error")
		return
	}

	// -------------------------
	// 랜덤 시각 생성 N개
	// -------------------------
	for i := 0; i < count; i++ {

		// 범위 내 랜덤 offset (초 단위)
		var offsetSec int64
		if diff > 0 {
			offsetSec = rand.Int63n(int64(diff) + 1)
		} else {
			offsetSec = 0
		}

		t := startTime.Add(time.Duration(offsetSec) * time.Second)

		// 랜덤 밀리초 0~999
		msec := rand.Intn(1000)
		t = t.Add(time.Duration(msec) * time.Millisecond)

		// 출력 형식: YYYYMMDDHHMMSS.mmm
		fmt.Println(t.Format("20060102150405.000"))
	}
}

