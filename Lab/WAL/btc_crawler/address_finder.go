package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// AddressFinder 구조체
type AddressFinder struct {
	client    *http.Client
	userAgent string
}

// NewAddressFinder 새로운 주소 찾기 도구 생성
func NewAddressFinder() *AddressFinder {
	return &AddressFinder{
		client:    &http.Client{Timeout: 30 * time.Second},
		userAgent: "Mozilla/5.0 (compatible; Address-Finder/1.0)",
	}
}

// 최신 트랜잭션에서 활성 주소 찾기
func (f *AddressFinder) FindActiveAddresses() (map[string][]string, error) {
	fmt.Println("최신 BTC 트랜잭션에서 활성 주소 찾기...")

	// 최신 트랜잭션 가져오기
	url := "https://api.blockcypher.com/v1/btc/main/txs?limit=50"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", f.userAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP 오류: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var txs []map[string]interface{}
	if err := json.Unmarshal(body, &txs); err != nil {
		return nil, err
	}

	// 주소 분류
	exchangeCandidates := make(map[string][]string)

	for _, tx := range txs {
		if addresses, ok := tx["addresses"].([]interface{}); ok && len(addresses) >= 2 {
			fromAddr := fmt.Sprintf("%v", addresses[0])
			toAddr := fmt.Sprintf("%v", addresses[1])

			// 큰 금액 트랜잭션은 거래소일 가능성 높음
			if total, ok := tx["total"].(float64); ok {
				amountBTC := total / 100000000 // satoshi to BTC

				if amountBTC > 1.0 { // 1 BTC 이상
					exchangeCandidates["large_transactions"] = append(exchangeCandidates["large_transactions"], fromAddr, toAddr)
				}

				if amountBTC > 10.0 { // 10 BTC 이상
					exchangeCandidates["very_large_transactions"] = append(exchangeCandidates["very_large_transactions"], fromAddr, toAddr)
				}
			}

			// 여러 입력/출력이 있는 트랜잭션은 거래소일 가능성
			if len(addresses) > 3 {
				exchangeCandidates["multi_address_transactions"] = append(exchangeCandidates["multi_address_transactions"], fromAddr, toAddr)
			}
		}
	}

	// 중복 제거
	for category := range exchangeCandidates {
		exchangeCandidates[category] = removeDuplicates(exchangeCandidates[category])
	}

	return exchangeCandidates, nil
}

// 주소 상세 정보 확인
func (f *AddressFinder) AnalyzeAddress(address string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.blockcypher.com/v1/btc/main/addrs/%s", address)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", f.userAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP 오류: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var addrInfo map[string]interface{}
	if err := json.Unmarshal(body, &addrInfo); err != nil {
		return nil, err
	}

	return addrInfo, nil
}

// 중복 제거 함수
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

func main() {
	fmt.Println("=== BTC 거래소 주소 찾기 도구 ===\n")

	finder := NewAddressFinder()

	// 활성 주소 찾기
	candidates, err := finder.FindActiveAddresses()
	if err != nil {
		log.Fatalf("주소 찾기 실패: %v", err)
	}

	fmt.Println("발견된 거래소 후보 주소들:")

	for category, addresses := range candidates {
		fmt.Printf("\n[%s] (%d개 주소):\n", category, len(addresses))
		for i, addr := range addresses {
			if i >= 10 { // 각 카테고리별 10개만 표시
				fmt.Printf("  ... 그 외 %d개 주소\n", len(addresses)-10)
				break
			}
			fmt.Printf("  %d. %s\n", i+1, addr)

			// 주소 상세 정보 분석
			if info, err := finder.AnalyzeAddress(addr); err == nil {
				if balance, ok := info["balance"].(float64); ok {
					balanceBTC := balance / 100000000
					fmt.Printf("     잔액: %.8f BTC\n", balanceBTC)
				}

				if nTx, ok := info["n_tx"].(float64); ok {
					fmt.Printf("     트랜잭션 수: %.0f\n", nTx)

					// 거래소 특징 판단
					if nTx > 1000 {
						fmt.Printf("     → 거래소 가능성 높음 (많은 트랜잭션)\n")
					}
				}
				if balance, ok := info["balance"].(float64); ok && balance > 1000000000 { // 10 BTC 이상
					fmt.Printf("     → 거래소 가능성 높음 (큰 잔액)\n")
				}
			}
			fmt.Println()
		}
	}

	fmt.Println("\n=== 추천 ===")
	fmt.Println("위 주소들 중 트랜잭션이 많고 잔액이 큰 주소들이 거래소일 가능성이 높습니다.")
	fmt.Println("이 주소들을 exchange_crawler.go의 getExchangeAddresses() 함수에 추가하세요.")
}
