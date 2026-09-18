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

// Transaction 구조체
type Transaction struct {
	Hash        string    `json:"hash"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	Amount      string    `json:"amount"`
	Timestamp   time.Time `json:"timestamp"`
	BlockNumber int64     `json:"block_number"`
	Token       string    `json:"token"`
}

// BTCPublicCrawler 구조체
type BTCPublicCrawler struct {
	client    *http.Client
	userAgent string
}

// NewBTCPublicCrawler 새로운 공개 API 크롤러 생성
func NewBTCPublicCrawler() *BTCPublicCrawler {
	return &BTCPublicCrawler{
		client:    &http.Client{Timeout: 30 * time.Second},
		userAgent: "Mozilla/5.0 (compatible; BTC-Crawler/1.0)",
	}
}

// 공개 API에서 BTC 최근 전송 내역 가져오기
func (c *BTCPublicCrawler) CrawlBTCRecentTransfers() ([]Transaction, error) {
	fmt.Println("공개 블록체인 API에서 BTC 데이터 수집 중...")
	
	// 여러 공개 API 시도
	apis := []struct {
		name string
		url  string
		parser func([]byte) ([]Transaction, error)
	}{
		{
			name: "BlockCypher",
			url:  "https://api.blockcypher.com/v1/btc/main/txs",
			parser: c.parseBlockCypher,
		},
		{
			name: "Blockstream",
			url:  "https://blockstream.info/api/blocks",
			parser: c.parseBlockstream,
		},
		{
			name: "Mempool",
			url:  "https://mempool.space/api/blocks",
			parser: c.parseMempool,
		},
	}
	
	for _, api := range apis {
		fmt.Printf("\n%s API 시도 중...\n", api.name)
		
		req, err := http.NewRequest("GET", api.url, nil)
		if err != nil {
			fmt.Printf("요청 생성 실패: %v\n", err)
			continue
		}
		
		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Accept", "application/json")
		
		resp, err := c.client.Do(req)
		if err != nil {
			fmt.Printf("요청 실패: %v\n", err)
			continue
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("HTTP 오류: %d\n", resp.StatusCode)
			continue
		}
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("응답 읽기 실패: %v\n", err)
			continue
		}
		
		transactions, err := api.parser(body)
		if err != nil {
			fmt.Printf("파싱 실패: %v\n", err)
			continue
		}
		
		if len(transactions) > 0 {
			fmt.Printf("✓ %s: %d건의 트랜잭션 수집 성공!\n", api.name, len(transactions))
			return transactions, nil
		}
	}
	
	// 모든 API 실패 시 샘플 데이터 반환
	fmt.Println("\n모든 공개 API 실패. 샘플 데이터를 반환합니다.")
	return c.getSampleData(), nil
}

// BlockCypher API 파싱
func (c *BTCPublicCrawler) parseBlockCypher(body []byte) ([]Transaction, error) {
	var response []map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	
	var transactions []Transaction
	for _, tx := range response {
		if hash, ok := tx["hash"].(string); ok {
			transaction := Transaction{
				Hash:  hash,
				Token: "BTC",
			}
			
			// 금액 계산 (간단화)
			if total, ok := tx["total"].(float64); ok {
				transaction.Amount = fmt.Sprintf("%.8f", total/100000000) // satoshi to BTC
			}
			
			// 시간
			if timeStr, ok := tx["received"].(string); ok {
				if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
					transaction.Timestamp = t
				}
			}
			
			// 블록
			if blockHeight, ok := tx["block_height"].(float64); ok {
				transaction.BlockNumber = int64(blockHeight)
			}
			
			// 주소들
			if addresses, ok := tx["addresses"].([]interface{}); ok && len(addresses) >= 2 {
				if from, ok := addresses[0].(string); ok {
					transaction.From = from
				}
				if to, ok := addresses[1].(string); ok {
					transaction.To = to
				}
			}
			
			transactions = append(transactions, transaction)
		}
	}
	
	return transactions, nil
}

// Blockstream API 파싱
func (c *BTCPublicCrawler) parseBlockstream(body []byte) ([]Transaction, error) {
	var blocks []map[string]interface{}
	if err := json.Unmarshal(body, &blocks); err != nil {
		return nil, err
	}
	
	var transactions []Transaction
	for _, block := range blocks[:3] { // 최근 3개 블록
		if id, ok := block["id"].(string); ok {
			// 블록 상세 정보 가져오기
			blockURL := fmt.Sprintf("https://blockstream.info/api/block/%s/txs", id)
			if txs, err := c.getBlockTransactions(blockURL); err == nil {
				transactions = append(transactions, txs...)
			}
		}
	}
	
	return transactions, nil
}

// Mempool API 파싱
func (c *BTCPublicCrawler) parseMempool(body []byte) ([]Transaction, error) {
	var blocks []map[string]interface{}
	if err := json.Unmarshal(body, &blocks); err != nil {
		return nil, err
	}
	
	var transactions []Transaction
	for _, block := range blocks[:3] { // 최근 3개 블록
		if id, ok := block["id"].(string); ok {
			// 블록 상세 정보 가져오기
			blockURL := fmt.Sprintf("https://mempool.space/api/block/%s/txs", id)
			if txs, err := c.getBlockTransactions(blockURL); err == nil {
				transactions = append(transactions, txs...)
			}
		}
	}
	
	return transactions, nil
}

// 블록 트랜잭션 가져오기
func (c *BTCPublicCrawler) getBlockTransactions(url string) ([]Transaction, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", c.userAgent)
	
	resp, err := c.client.Do(req)
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
	
	var transactions []Transaction
	for _, tx := range txs[:5] { // 블록당 5개 트랜잭션만
		if txid, ok := tx["txid"].(string); ok {
			transaction := Transaction{
				Hash:  txid,
				Token: "BTC",
			}
			
			// 금액과 주소 정보 추출 (간단화)
			if vout, ok := tx["vout"].([]interface{}); ok && len(vout) > 0 {
				if firstVout, ok := vout[0].(map[string]interface{}); ok {
					if value, ok := firstVout["value"].(float64); ok {
						transaction.Amount = fmt.Sprintf("%.8f", value)
					}
					if scriptPubKey, ok := firstVout["scriptpubkey"].(map[string]interface{}); ok {
						if address, ok := scriptPubKey["address"].(string); ok {
							transaction.To = address
						}
					}
				}
			}
			
			if vin, ok := tx["vin"].([]interface{}); ok && len(vin) > 0 {
				if firstVin, ok := vin[0].(map[string]interface{}); ok {
					if prevout, ok := firstVin["prevout"].(map[string]interface{}); ok {
						if scriptPubKey, ok := prevout["scriptpubkey"].(map[string]interface{}); ok {
							if address, ok := scriptPubKey["address"].(string); ok {
								transaction.From = address
							}
						}
					}
				}
			}
			
			// 블록 정보
			if blockHeight, ok := tx["status"].(map[string]interface{}); ok {
				if height, ok := blockHeight["block_height"].(float64); ok {
					transaction.BlockNumber = int64(height)
				}
			}
			
			transactions = append(transactions, transaction)
		}
	}
	
	return transactions, nil
}

// 샘플 데이터
func (c *BTCPublicCrawler) getSampleData() []Transaction {
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
	}
}

// 결과 저장
func (c *BTCPublicCrawler) SaveToFile(transactions []Transaction, filename string) error {
	data, err := json.MarshalIndent(transactions, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 마샬링 실패: %v", err)
	}
	
	return os.WriteFile(filename, data, 0644)
}

// 결과 출력
func (c *BTCPublicCrawler) PrintTransactions(transactions []Transaction) {
	fmt.Printf("\n=== BTC 최근 전송 내역 (%d건) ===\n\n", len(transactions))
	
	for i, tx := range transactions {
		fmt.Printf("[%d] 해시: %s\n", i+1, tx.Hash)
		fmt.Printf("    From: %s\n", tx.From)
		fmt.Printf("    To: %s\n", tx.To)
		fmt.Printf("    금액: %s %s\n", tx.Amount, tx.Token)
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
	fmt.Println("BTC 공개 API 크롤러 시작...")
	fmt.Println("여러 공개 블록체인 API에서 BTC 데이터를 수집합니다.\n")
	
	crawler := NewBTCPublicCrawler()
	
	transactions, err := crawler.CrawlBTCRecentTransfers()
	if err != nil {
		log.Fatalf("데이터 수집 실패: %v", err)
	}
	
	// 결과 출력
	crawler.PrintTransactions(transactions)
	
	// 파일 저장
	if len(transactions) > 0 {
		filename := fmt.Sprintf("btc_public_transfers_%s.json", time.Now().Format("20060102_150405"))
		if err := crawler.SaveToFile(transactions, filename); err != nil {
			fmt.Printf("파일 저장 실패: %v\n", err)
		} else {
			fmt.Printf("결과가 %s 파일에 저장되었습니다.\n", filename)
		}
	}
	
	fmt.Println("\n크롤링 완료!")
}
