package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
//	"os"
	"strconv"
	"strings"
	"time"
	"io/ioutil"
	//"github.com/PuerkitoBio/goquery"
	"github.com/schollz/progressbar/v3"
)

// UserAgents는 랜덤 User-Agent 목록입니다.
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36",
}

// getRandomUserAgent는 랜덤한 User-Agent를 반환합니다.
func getRandomUserAgent() string {
	rand.Seed(time.Now().UnixNano())
	return userAgents[rand.Intn(len(userAgents))]
}

// fetchExchangeRate는 investing.com에서 USD/KRW 환율과 거래 시간을 가져옵니다.
func fetchExchangeRate() (float64, string, error) {
	url := "https://kr.investing.com/currencies/usd-krw"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, "", fmt.Errorf("요청 생성 실패: %v", err)
	}

	// 헤더 설정
	req.Header.Set("User-Agent", getRandomUserAgent())
	req.Header.Set("Accept-Language", "ko-KR,ko;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("요청 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("HTTP 오류: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, "", fmt.Errorf("응답 읽기 실패: %v", err)
	}

	// 가격 정보 추출
	priceMarker := "data-test=\"instrument-price-last\">"
	priceStart := strings.Index(string(body), priceMarker)
	if priceStart == -1 {
		// 디버깅을 위해 HTML 일부를 로그로 남깁니다
		htmlSample := string(body)
		if len(htmlSample) > 1000 {
			htmlSample = htmlSample[:1000] + "..."
		}
		return 0, "", fmt.Errorf("가격 정보를 찾을 수 없습니다. HTML 샘플: %s", htmlSample)
	}

	priceStart = priceStart + len(priceMarker)
	// 닫는 태그를 찾을 때까지 순회
	var priceEnd int
	for i := priceStart; i < len(body) && i < priceStart+100; i++ {
		if body[i] == '<' {
			priceEnd = i
			break
		}
	}

	if priceEnd <= priceStart {
		return 0, "", fmt.Errorf("가격 정보 형식이 올바르지 않습니다 (닫는 태그를 찾을 수 없음)")
	}

	rateText := string(body[priceStart:priceEnd])
	rateText = strings.TrimSpace(rateText)
	rateText = strings.ReplaceAll(rateText, ",", "")

	// 거래 시간 정보 추출
	timeMarker := "data-test=\"trading-time-label\">"
	timeStart := strings.Index(string(body), timeMarker)
	var timeText string
	if timeStart != -1 {
		timeStart = timeStart + len(timeMarker)
		timeEnd := strings.Index(string(body[timeStart:]), "<")
		if timeEnd > 0 {
			timeText = string(body[timeStart : timeStart+timeEnd])
			timeText = strings.TrimSpace(timeText)
		}
	}

	rate, err := strconv.ParseFloat(rateText, 64)
	if err != nil {
		return 0, "", fmt.Errorf("가격 파싱 실패: %v", err)
	}

	return rate, timeText, nil
}

func main() {
	fmt.Println("USD/KRW 환율 모니터링을 시작합니다...")
	fmt.Println("종료하려면 Ctrl+C를 누르세요.")

	// 진행 상황 표시를 위한 프로그레스 바
	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription("다음 업데이트까지 대기 중..."),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionClearOnFinish(),
	)

	for {
		// 랜덤한 대기 시간 (30초 ~ 5분)
		waitTime := time.Duration(30+rand.Intn(270)) * time.Second
		
		// 진행 상황 바 업데이트
		bar.Reset()
		bar.ChangeMax(int(waitTime.Seconds()))
		
		// 대기 시간 동안 진행 상황 표시
		go func() {
			for i := 0; i < int(waitTime.Seconds()); i++ {
				time.Sleep(time.Second)
				bar.Add(1)
			}
		}()

		// 대기
		time.Sleep(waitTime)

		// 환율 조회
		rate, tradeTime, err := fetchExchangeRate()
		if err != nil {
			log.Printf("오류: %v\n", err)
		} else {
			if tradeTime != "" {
				fmt.Printf("[%s] USD/KRW: %.2f원 (기준: %s)\n", time.Now().Format("2006-01-02 15:04:05"), rate, tradeTime)
			} else {
				fmt.Printf("[%s] USD/KRW: %.2f원\n", time.Now().Format("2006-01-02 15:04:05"), rate)
			}
		}
	}
}
