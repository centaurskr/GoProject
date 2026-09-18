//
// Free plan : 하루에 한번 Update된 환율, API조회
// 구독금액 : https://app.exchangerate-api.com/plan
//

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// ExchangeRateResponse는 환율 API 응답 구조체입니다.
type ExchangeRateResponse struct {
	Result             string             `json:"result"`
	Documentation      string             `json:"documentation"`
	TermsOfUse         string             `json:"terms_of_use"`
	TimeLastUpdateUnix int64              `json:"time_last_update_unix"`
	TimeLastUpdateUTC  string             `json:"time_last_update_utc"`
	TimeNextUpdateUnix int64              `json:"time_next_update_unix"`
	TimeNextUpdateUTC  string             `json:"time_next_update_utc"`
	BaseCode           string             `json:"base_code"`
	ConversionRates    map[string]float64 `json:"conversion_rates"`
}

// getAvailableCurrencies는 사용 가능한 통화 목록을 문자열로 반환합니다.
func getAvailableCurrencies(rates map[string]float64) string {
	currencies := make([]string, 0, len(rates))
	for k := range rates {
		currencies = append(currencies, k)
	}
	return strings.Join(currencies, ", ")
}

func main() {
	fmt.Println("USD/KRW 환율 조회를 시작합니다...")
	fmt.Println("종료하려면 Ctrl+C를 누르세요.")

	// API 키는 실제 사용 시 환경 변수 등으로 관리하는 것이 좋습니다.
	apiKey := "acaf0c55cfec9fd48d6858b7" // 실제 API 키로 교체 필요
	baseCurrency := "USD"
	targetCurrency := "KRW"
	
	// API 키 유효성 검사 (간단한 형식 체크)
	if len(apiKey) != 24 {
		log.Fatal("올바르지 않은 API 키 형식입니다. 24자리 문자열이어야 합니다.")
	}

	for {
		rate, lastUpdate, err := getExchangeRate(apiKey, baseCurrency, targetCurrency)
		if err != nil {
			log.Printf("오류: %v\n", err)
		} else {
			fmt.Printf("[%s] %s/%s: %.2f (기준: %s)\n",
				time.Now().Format("2006-01-02 15:04:05"),
				baseCurrency,
				targetCurrency,
				rate,
				lastUpdate,
			)
		}

		// 1분마다 업데이트
		time.Sleep(1 * time.Minute)
	}
}

// getExchangeRate는 API를 통해 환율을 조회하고, 환율과 마지막 업데이트 시간을 반환합니다.
func getExchangeRate(apiKey, base, target string) (float64, string, error) {
	// ExchangeRate-API (https://www.exchangerate-api.com/) 사용 예시
	// 무료 플랜에서는 base currency를 USD로만 사용 가능
	url := fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/latest/USD", apiKey)

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

	var result ExchangeRateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, "", fmt.Errorf("JSON 파싱 실패: %v", err)
	}

	// KRW 환율 조회 (대소문자 구분 없이 처리)
	target = strings.ToUpper(target)
	rate, exists := result.ConversionRates[target]
	if !exists {
		return 0, "", fmt.Errorf("지원하지 않는 통화 코드: %s (사용 가능한 통화: %v)", 
			target, getAvailableCurrencies(result.ConversionRates))
	}

	// 마지막 업데이트 시간 포맷팅
	lastUpdate := "알 수 없음"
	if result.TimeLastUpdateUTC != "" {
		// 예: "Wed, 20 Aug 2025 04:30:00 +0000" -> "2025-08-20 13:30:00"
		t, err := time.Parse(time.RFC1123Z, result.TimeLastUpdateUTC)
		if err == nil {
			// 한국 시간(KST)으로 변환 (UTC+9)
			loc, _ := time.LoadLocation("Asia/Seoul")
			lastUpdate = t.In(loc).Format("2006-01-02 15:04:05")
		}
	}

	return rate, lastUpdate, nil
}
