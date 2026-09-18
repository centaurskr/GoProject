package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Transaction 구조체는 트랜잭션 데이터를 저장합니다
type Transaction struct {
	Hash        string    `json:"hash"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	Amount      string    `json:"amount"`
	Timestamp   time.Time `json:"timestamp"`
	BlockNumber int64     `json:"block_number"`
	FromLabel   string    `json:"from_label,omitempty"`
	ToLabel     string    `json:"to_label,omitempty"`
	Token       string    `json:"token"`
}

// BTCApiCrawler 구조체
type BTCApiCrawler struct {
	client    *http.Client
	userAgent string
}

// NewBTCApiCrawler 새로운 BTC API 크롤러 생성
func NewBTCApiCrawler() *BTCApiCrawler {
	return &BTCApiCrawler{
		client:    &http.Client{Timeout: 30 * time.Second},
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
}

// Arkham API에서 BTC 최근 전송 내역 가져오기
func (c *BTCApiCrawler) CrawlBTCRecentTransfers() ([]Transaction, error) {
	fmt.Println("Arkham API에 직접 요청 시도 중...")

	// 가능한 API 엔드포인트들
	apiEndpoints := []string{
		"https://intel.arkm.com/api/v1/transfers?token=BTC&limit=50",
		"https://intel.arkm.com/api/v2/transfers/btc?limit=50",
		"https://api.arkm.com/v1/transfers?token=BTC&limit=50",
		"https://intel.arkm.com/api/transfers?token=BTC&limit=50",
	}

	var transactions []Transaction

	for i, endpoint := range apiEndpoints {
		fmt.Printf("API 엔드포인트 %d 시도: %s\n", i+1, endpoint)

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			fmt.Printf("  요청 생성 실패: %v\n", err)
			continue
		}

		// 헤더 설정
		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Accept-Language", "ko-KR,ko;q=0.9,en;q=0.8")
		req.Header.Set("Referer", "https://intel.arkm.com/")
		req.Header.Set("Origin", "https://intel.arkm.com")

		resp, err := c.client.Do(req)
		if err != nil {
			fmt.Printf("  요청 실패: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		fmt.Printf("  상태 코드: %d\n", resp.StatusCode)

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("  응답 읽기 실패: %v\n", err)
				continue
			}

			// 응답 내용 미리보기
			bodyStr := string(body)
			if len(bodyStr) > 200 {
				fmt.Printf("  응답 미리보기: %s...\n", bodyStr[:200])
			} else {
				fmt.Printf("  응답: %s\n", bodyStr)
			}

			// JSON 파싱 시도
			var apiData map[string]interface{}
			if err := json.Unmarshal(body, &apiData); err == nil {
				parsedTransactions := c.parseAPIResponse(apiData)
				if len(parsedTransactions) > 0 {
					fmt.Printf("  ✓ %d건의 트랜잭션 발견!\n", len(parsedTransactions))
					transactions = append(transactions, parsedTransactions...)
					break
				} else {
					fmt.Printf("  트랜잭션 데이터 없음\n")
				}
			} else {
				fmt.Printf("  JSON 파싱 실패: %v\n", err)
			}
		} else {
			errorBody, _ := io.ReadAll(resp.Body)
			fmt.Printf("  오류 응답: %s\n", string(errorBody))
		}
	}

	// 만약 API로 데이터를 못 찾으면, 샘플 데이터로 데모
	if len(transactions) == 0 {
		fmt.Println("\nAPI에서 데이터를 찾지 못했습니다. 샘플 데이터로 데모를 실행합니다.")
		transactions = c.getSampleData()
	}

	return transactions, nil
}

// API 응답 파싱
func (c *BTCApiCrawler) parseAPIResponse(data map[string]interface{}) []Transaction {
	var transactions []Transaction

	// 다양한 API 응답 형식 시도

	// 1. data.transfers 형식
	if dataObj, ok := data["data"].(map[string]interface{}); ok {
		if transfers, ok := dataObj["transfers"].([]interface{}); ok {
			for _, transfer := range transfers {
				if tx := c.parseTransferItem(transfer); tx != nil {
					transactions = append(transactions, *tx)
				}
			}
		}
	}

	// 2. transfers 형식
	if transfers, ok := data["transfers"].([]interface{}); ok {
		for _, transfer := range transfers {
			if tx := c.parseTransferItem(transfer); tx != nil {
				transactions = append(transactions, *tx)
			}
		}
	}

	// 3. results 형식
	if results, ok := data["results"].([]interface{}); ok {
		for _, result := range results {
			if tx := c.parseTransferItem(result); tx != nil {
				transactions = append(transactions, *tx)
			}
		}
	}

	// 4. 배열 형식
	if dataArray, ok := data["data"].([]interface{}); ok {
		for _, item := range dataArray {
			if tx := c.parseTransferItem(item); tx != nil {
				transactions = append(transactions, *tx)
			}
		}
	}

	return transactions
}

// 트랜잭션 아이템 파싱
func (c *BTCApiCrawler) parseTransferItem(item interface{}) *Transaction {
	if txMap, ok := item.(map[string]interface{}); ok {
		tx := Transaction{Token: "BTC"}

		// 해시
		if hash, ok := txMap["hash"].(string); ok {
			tx.Hash = hash
		} else if hash, ok := txMap["tx_hash"].(string); ok {
			tx.Hash = hash
		} else if hash, ok := txMap["transaction_hash"].(string); ok {
			tx.Hash = hash
		}

		// From 주소
		if from, ok := txMap["from"].(string); ok {
			tx.From = from
		} else if from, ok := txMap["from_address"].(string); ok {
			tx.From = from
		} else if from, ok := txMap["sender"].(string); ok {
			tx.From = from
		}

		// To 주소
		if to, ok := txMap["to"].(string); ok {
			tx.To = to
		} else if to, ok := txMap["to_address"].(string); ok {
			tx.To = to
		} else if to, ok := txMap["receiver"].(string); ok {
			tx.To = to
		}

		// 금액
		if amount, ok := txMap["amount"].(string); ok {
			tx.Amount = amount
		} else if amount, ok := txMap["value"].(string); ok {
			tx.Amount = amount
		} else if amount, ok := txMap["quantity"].(string); ok {
			tx.Amount = amount
		}

		// 타임스탬프
		if timestamp, ok := txMap["timestamp"].(string); ok {
			if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
				tx.Timestamp = t
			}
		} else if timestamp, ok := txMap["time"].(string); ok {
			if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
				tx.Timestamp = t
			}
		}

		// 블록 번호
		if block, ok := txMap["block"].(float64); ok {
			tx.BlockNumber = int64(block)
		} else if block, ok := txMap["block_number"].(float64); ok {
			tx.BlockNumber = int64(block)
		}

		// 라벨 정보
		if fromEntity, ok := txMap["from_entity"].(map[string]interface{}); ok {
			if name, ok := fromEntity["name"].(string); ok {
				tx.FromLabel = name
			}
		}

		if toEntity, ok := txMap["to_entity"].(map[string]interface{}); ok {
			if name, ok := toEntity["name"].(string); ok {
				tx.ToLabel = name
			}
		}

		// 토큰 정보
		if token, ok := txMap["token"].(map[string]interface{}); ok {
			if symbol, ok := token["symbol"].(string); ok {
				tx.Token = symbol
			}
		}

		// 유효한 트랜잭션만 반환
		if tx.Hash != "" && tx.From != "" && tx.To != "" {
			return &tx
		}
	}

	return nil
}

// 샘플 데이터 (데모용)
func (c *BTCApiCrawler) getSampleData() []Transaction {
	now := time.Now()
	return []Transaction{
		{
			Hash:        "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456",
			From:        "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			To:          "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2",
			Amount:      "0.12345678",
			Timestamp:   now.Add(-1 * time.Hour),
			BlockNumber: 823456,
			Token:       "BTC",
		},
		{
			Hash:        "f6e5d4c3b2a198765432109876543210fedcba9876543210fedcba987654321",
			From:        "3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy",
			To:          "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
			Amount:      "0.98765432",
			Timestamp:   now.Add(-2 * time.Hour),
			BlockNumber: 823455,
			Token:       "BTC",
		},
		{
			Hash:        "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			From:        "1FeexV6bYb9pCwBCLy5HQi7FvT3TtFfjZ6",
			To:          "1F1tAaz5x1HUXrCNLbtMDqcw6o5GNn4xqX",
			Amount:      "1.00000000",
			Timestamp:   now.Add(-3 * time.Hour),
			BlockNumber: 823454,
			Token:       "BTC",
		},
	}
}

// 결과를 JSON 파일로 저장
func (c *BTCApiCrawler) SaveToFile(transactions []Transaction, filename string) error {
	data, err := json.MarshalIndent(transactions, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 마샬링 실패: %v", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// 결과를 콘솔에 출력
func (c *BTCApiCrawler) PrintTransactions(transactions []Transaction) {
	fmt.Printf("\n=== BTC 최근 전송 내역 (%d건) ===\n\n", len(transactions))

	if len(transactions) == 0 {
		fmt.Println("수집된 데이터가 없습니다.")
		return
	}

	for i, tx := range transactions {
		fmt.Printf("[%d] 해시: %s\n", i+1, tx.Hash)
		fmt.Printf("    From: %s", tx.From)
		if tx.FromLabel != "" {
			fmt.Printf(" (%s)", tx.FromLabel)
		}
		fmt.Printf("\n    To: %s", tx.To)
		if tx.ToLabel != "" {
			fmt.Printf(" (%s)", tx.ToLabel)
		}
		fmt.Printf("\n    금액: %s %s\n", tx.Amount, tx.Token)
		if !tx.Timestamp.IsZero() {
			fmt.Printf("    시간: %s\n", tx.Timestamp.Format("2006-01-02 15:04:05"))
		}
		if tx.BlockNumber > 0 {
			fmt.Printf("    블록: %d\n", tx.BlockNumber)
		}
		fmt.Println()
	}
}

func main() {
	fmt.Println("BTC API 크롤러 시작...")
	fmt.Println("Arkham Intel API에서 BTC 최근 전송 내역을 수집합니다.\n")

	crawler := NewBTCApiCrawler()

	transactions, err := crawler.CrawlBTCRecentTransfers()
	if err != nil {
		log.Fatalf("데이터 수집 실패: %v", err)
	}

	// 결과 출력
	crawler.PrintTransactions(transactions)

	// 파일 저장
	if len(transactions) > 0 {
		filename := fmt.Sprintf("btc_transfers_%s.json", time.Now().Format("20060102_150405"))
		if err := crawler.SaveToFile(transactions, filename); err != nil {
			fmt.Printf("파일 저장 실패: %v\n", err)
		} else {
			fmt.Printf("결과가 %s 파일에 저장되었습니다.\n", filename)
		}
	}

	fmt.Println("\n크롤링 완료!")
}
