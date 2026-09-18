//
// Alpha Vantage API를 사용한 환율 조회 프로그램
// - 무료 API 키는 하루 25회, 1시간에 1회만 조회 가능
// - 1시간마다 자동으로 환율을 가져와 출력
//
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)


// 전역 변수로 API 호출 횟수 관리
var apiCallCount int32

func main() {
	fmt.Println("Alpha Vantage 환율 조회를 시작합니다...")
	fmt.Println("하루 최대 25회, 1시간에 1회만 조회합니다.")
	fmt.Println("종료하려면 Ctrl+C를 누르세요.")

	// Alpha Vantage API 키
	apiKey := "CGKF7AX2H13T4UHN"
	fromCurrency := "USD"
	toCurrency := "KRW"

	// API 키 유효성 검사
	if len(apiKey) < 10 {
		log.Fatal("올바르지 않은 API 키 형식입니다.")
	}

	// 채널 생성 (에러 및 종료 신호용)
	errChan := make(chan error)
	done := make(chan bool)

	// 1시간마다 실행되는 티커 생성
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// 즉시 한 번 실행
	go fetchAndDisplay(apiKey, fromCurrency, toCurrency, errChan)

	// 메인 루프
	for {
		select {
		case err := <-errChan:
			log.Printf("오류 발생: %v\n", err)
		case <-ticker.C:
			go fetchAndDisplay(apiKey, fromCurrency, toCurrency, errChan)
		case <-done:
			return
		}
	}
}

func fetchAndDisplay(apiKey, fromCurrency, toCurrency string, errChan chan<- error) {
	// 하루 최대 25회로 제한
	if atomic.LoadInt32(&apiCallCount) >= 25 {
		errChan <- fmt.Errorf("일일 API 호출 한도(25회)에 도달했습니다.")
		return
	}

	// API 호출 횟수 증가
	count := atomic.AddInt32(&apiCallCount, 1)

	rate, lastRefreshed, err := getExchangeRate(apiKey, fromCurrency, toCurrency)
	if err != nil {
		errChan <- fmt.Errorf("환율 조회 실패: %v", err)
		return
	}

	// 한국 시간으로 변환
	loc, _ := time.LoadLocation("Asia/Seoul")
	utcTime, err := time.ParseInLocation("2006-01-02 15:04:05", lastRefreshed, time.UTC)
	if err != nil {
		errChan <- fmt.Errorf("시간 파싱 오류: %v", err)
		return
	}
	kstTime := utcTime.In(loc)

	fmt.Printf("\n[%d/25][%s] %s/%s: %.2f (기준: %s)\n",
		count,
		time.Now().Format("2006-01-02 15:04:05"),
		fromCurrency,
		toCurrency,
		rate,
		kstTime.Format("2006-01-02 15:04:05 MST"))
}

// AlphaVantageResponse is the structure for Alpha Vantage API response
type AlphaVantageResponse struct {
	RealtimeCurrencyExchangeRate struct {
		FromCurrencyCode string `json:"1. From_Currency Code"`
		FromCurrencyName string `json:"2. From_Currency Name"`
		ToCurrencyCode   string `json:"3. To_Currency Code"`
		ToCurrencyName   string `json:"4. To_Currency Name"`
		ExchangeRate     string `json:"5. Exchange Rate"`
		LastRefreshed    string `json:"6. Last Refreshed"`
		TimeZone         string `json:"7. Time Zone"`
		BidPrice         string `json:"8. Bid Price"`
		AskPrice         string `json:"9. Ask Price"`
	} `json:"Realtime Currency Exchange Rate"`
}

// getExchangeRate는 Alpha Vantage API를 통해 환율을 조회합니다.
func getExchangeRate(apiKey, fromCurrency, toCurrency string) (float64, string, error) {
	// Alpha Vantage API 엔드포인트
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=CURRENCY_EXCHANGE_RATE&from_currency=%s&to_currency=%s&apikey=%s",
		fromCurrency, toCurrency, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return 0, "", fmt.Errorf("API 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("API 응답 오류: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, "", fmt.Errorf("응답 읽기 실패: %v", err)
	}
	

	var result AlphaVantageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, "", fmt.Errorf("JSON 파싱 실패: %v", err)
	}

	fmt.Printf("RESP-%+v\n", resp)
	fmt.Println(string(body))
	fmt.Printf("RESULT-%+v\n", result)

	// 환율이 비어있는지 확인
	if result.RealtimeCurrencyExchangeRate.ExchangeRate == "" {
		// API 호출 횟수 감소 (실패한 경우)
		atomic.AddInt32(&apiCallCount, -1)
		return 0, "", fmt.Errorf("환율 정보를 찾을 수 없습니다")
	}
	// 문자열을 float64로 변환
	rate := 0.0
	fmt.Sscanf(result.RealtimeCurrencyExchangeRate.ExchangeRate, "%f", &rate)

	return rate, result.RealtimeCurrencyExchangeRate.LastRefreshed, nil
}
