package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// UserAgents는 랜덤 User-Agent 목록입니다.
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
}

func main() {
	fmt.Println("USD/KRW 실시간 환율 모니터링을 시작합니다...")
	fmt.Println("종료하려면 Ctrl+C를 누르세요.")

	// 초기 대기 시간 (30초)
	pollingInterval := 30 * time.Second

	for {
		rate, timeText, err := fetchExchangeRate()
		if err != nil {
			log.Printf("오류: %v\n", err)
		} else {
			if timeText != "" {
				fmt.Printf("[%s] USD/KRW: %.2f원 (기준: %s)\n", 
					time.Now().Format("2006-01-02 15:04:05"), 
					rate, 
					timeText,
				)
			} else {
				fmt.Printf("[%s] USD/KRW: %.2f원\n", 
					time.Now().Format("2006-01-02 15:04:05"), 
					rate,
				)
			}
		}

		// 다음 업데이트까지 대기
		time.Sleep(pollingInterval)
	}
}

// getRandomUserAgent는 랜덤한 User-Agent를 반환합니다.
func getRandomUserAgent() string {
	rand.Seed(time.Now().UnixNano())
	return userAgents[rand.Intn(len(userAgents))]
}

// fetchExchangeRate는 investing.com에서 USD/KRW 환율을 가져옵니다.
func fetchExchangeRate() (float64, string, error) {
	url := "https://kr.investing.com/currencies/usd-krw"

	// HTTP 클라이언트 생성
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

	// 요청 보내기
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("요청 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("HTTP 오류: %s", resp.Status)
	}

	// HTML 파싱
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0, "", fmt.Errorf("HTML 파싱 실패: %v", err)
	}

	// 가격 추출
	priceText := doc.Find("span[data-test=instrument-price-last]").Text()
	priceText = strings.ReplaceAll(priceText, ",", "")
	price, err := strconv.ParseFloat(strings.TrimSpace(priceText), 64)
	if err != nil {
		return 0, "", fmt.Errorf("가격 파싱 실패: %v", err)
	}

	// 시간 추출
	timeText := doc.Find("time[data-test=instrument-price-date]").Text()
	timeText = strings.TrimSpace(timeText)

	return price, timeText, nil
}
