package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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
}

// BTCWebCrawler 구조체
type BTCWebCrawler struct {
	client    *http.Client
	userAgent string
}

// NewBTCWebCrawler 새로운 BTC 웹 크롤러 생성
func NewBTCWebCrawler() *BTCWebCrawler {
	return &BTCWebCrawler{
		client:    &http.Client{Timeout: 30 * time.Second},
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
}

// Arkham 웹사이트에서 BTC 최근 전송 내역 크롤링
func (c *BTCWebCrawler) CrawlBTCRecentTransfers() ([]Transaction, error) {
	url := "https://intel.arkm.com"

	fmt.Println("웹사이트 접속 중...")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("요청 생성 실패: %v", err)
	}

	// 브라우저처럼 보이게 헤더 설정
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ko-KR,ko;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Cache-Control", "max-age=0")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("요청 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP 상태 코드: %d", resp.StatusCode)
	}

	// gzip 압축 해제
	var reader io.Reader = resp.Body
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("gzip 압축 해제 실패: %v", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("HTML 파싱 실패: %v", err)
	}

	fmt.Println("HTML 분석 중...")

	var transactions []Transaction

	// 1. 먼저 JavaScript 데이터에서 추출 시도
	transactions = c.extractTransactionsFromScript(doc)
	if len(transactions) > 0 {
		fmt.Printf("JavaScript에서 %d건의 데이터를 찾았습니다.\n", len(transactions))
		return transactions, nil
	}

	// 2. API 엔드포인트 직접 시도
	transactions, err = c.tryDirectAPI()
	if err == nil && len(transactions) > 0 {
		fmt.Printf("API에서 %d건의 데이터를 찾았습니다.\n", len(transactions))
		return transactions, nil
	}

	// 3. 다양한 선택자로 시도
	fmt.Println("HTML 요소 검색 중...")
	selectors := []string{
		"table tr",
		".table tr",
		"[class*='transfer'] tr",
		"[class*='transaction'] tr",
		"[data-testid*='transfer'] tr",
		"[data-testid*='transaction'] tr",
		"tbody tr",
		".list-item",
		"[class*='row']",
	}

	foundElements := false
	for _, selector := range selectors {
		elements := doc.Find(selector)
		if elements.Length() > 0 {
			fmt.Printf("선택자 '%s'로 %d개 요소 발견\n", selector, elements.Length())
			foundElements = true

			elements.Each(func(i int, s *goquery.Selection) {
				// 디버깅: 요소 내용 출력
				elementText := strings.TrimSpace(s.Text())
				if elementText != "" {
					fmt.Printf("  요소 %d: %s\n", i+1, elementText)
				}

				tx := c.extractTransactionFromRow(s)
				if tx.Hash != "" && tx.From != "" && tx.To != "" {
					// 중복 확인
					isDuplicate := false
					for _, existingTx := range transactions {
						if existingTx.Hash == tx.Hash {
							isDuplicate = true
							break
						}
					}

					if !isDuplicate {
						transactions = append(transactions, tx)
						fmt.Printf("  ✓ 트랜잭션 발견: %s\n", tx.Hash[:16]+"...")
					}
				}
			})
		}
	}

	if !foundElements {
		fmt.Println("트랜잭션 테이블이나 리스트를 찾지 못했습니다.")
		fmt.Println("웹사이트가 JavaScript로 동적으로 로드할 수 있습니다.")

		// 페이지 내용 일부 출력
		fmt.Println("\n=== 페이지 내용 미리보기 ===")
		doc.Find("body").Each(func(i int, s *goquery.Selection) {
			text := s.Text()
			if len(text) > 1000 {
				text = text[:1000] + "..."
			}
			fmt.Println(text)
		})

		// 테이블 내용 출력
		fmt.Println("\n=== 테이블 내용 분석 ===")
		doc.Find("table").Each(func(i int, table *goquery.Selection) {
			fmt.Printf("테이블 %d:\n", i+1)
			table.Find("tr").Each(func(j int, row *goquery.Selection) {
				rowText := strings.TrimSpace(row.Text())
				if rowText != "" {
					fmt.Printf("  행 %d: %s\n", j+1, rowText)
				}
			})
		})
	}

	return transactions, nil
}

// API 엔드포인트 직접 시도
func (c *BTCWebCrawler) tryDirectAPI() ([]Transaction, error) {
	// Arkham의 가능한 API 엔드포인트들
	apiEndpoints := []string{
		"https://intel.arkm.com/api/v1/transfers",
		"https://intel.arkm.com/api/transfers",
		"https://api.arkm.com/v1/transfers",
	}

	for _, endpoint := range apiEndpoints {
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			continue
		}

		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Accept", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			// API 응답 파싱 시도
			var apiData map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&apiData); err == nil {
				return c.parseAPIResponse(apiData), nil
			}
		}
	}

	return nil, fmt.Errorf("API 시도 실패")
}

// API 응답 파싱
func (c *BTCWebCrawler) parseAPIResponse(data map[string]interface{}) []Transaction {
	var transactions []Transaction

	// 다양한 API 응답 형식 시도
	if transfers, ok := data["transfers"].([]interface{}); ok {
		for _, transfer := range transfers {
			if txMap, ok := transfer.(map[string]interface{}); ok {
				tx := Transaction{}
				if hash, ok := txMap["hash"].(string); ok {
					tx.Hash = hash
				}
				if from, ok := txMap["from"].(string); ok {
					tx.From = from
				}
				if to, ok := txMap["to"].(string); ok {
					tx.To = to
				}
				if amount, ok := txMap["amount"].(string); ok {
					tx.Amount = amount
				}

				if tx.Hash != "" && tx.From != "" && tx.To != "" {
					transactions = append(transactions, tx)
				}
			}
		}
	}

	return transactions
}

// 행에서 트랜잭션 추출
func (c *BTCWebCrawler) extractTransactionFromRow(s *goquery.Selection) Transaction {
	tx := Transaction{}

	// 셀 내용 추출
	s.Find("td, div, span").Each(func(j int, cell *goquery.Selection) {
		cellText := strings.TrimSpace(cell.Text())
		cellClass, _ := cell.Attr("class")

		// 해시 패턴 확인 (64자리 16진수)
		if len(cellText) == 64 && isHex(cellText) {
			tx.Hash = cellText
		}

		// 주소 패턴 확인 (BTC 주소)
		if len(cellText) >= 26 && len(cellText) <= 35 && (strings.HasPrefix(cellText, "1") || strings.HasPrefix(cellText, "3") || strings.HasPrefix(cellText, "bc1")) {
			if tx.From == "" {
				tx.From = cellText
			} else if tx.To == "" && cellText != tx.From {
				tx.To = cellText
			}
		}

		// 금액 패턴 확인
		if strings.Contains(cellText, "BTC") || strings.Contains(cellClass, "amount") || strings.Contains(cellClass, "value") {
			amount := strings.TrimSpace(strings.Replace(cellText, "BTC", "", -1))
			if amount != "" {
				tx.Amount = amount
			}
		}

		// 시간 정보 추출
		if strings.Contains(cellClass, "time") || strings.Contains(cellClass, "timestamp") {
			if parsedTime, err := parseTime(cellText); err == nil {
				tx.Timestamp = parsedTime
			}
		}

		// 블록 번호 추출
		if strings.Contains(cellClass, "block") {
			if blockNum, err := parseBlockNumber(cellText); err == nil {
				tx.BlockNumber = blockNum
			}
		}
	})

	return tx
}

// JavaScript 데이터에서 트랜잭션 추출
func (c *BTCWebCrawler) extractTransactionsFromScript(doc *goquery.Document) []Transaction {
	var transactions []Transaction

	// script 태그에서 데이터 추출
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		scriptContent := s.Text()

		// JSON 데이터 패턴 찾기
		if strings.Contains(scriptContent, "transfer") || strings.Contains(scriptContent, "transaction") {
			// 간단한 JSON 파싱 시도
			jsonData := extractJSONFromScript(scriptContent)
			if jsonData != "" {
				var txData []map[string]interface{}
				if err := json.Unmarshal([]byte(jsonData), &txData); err == nil {
					for _, tx := range txData {
						transaction := Transaction{}
						if hash, ok := tx["hash"].(string); ok {
							transaction.Hash = hash
						}
						if from, ok := tx["from"].(string); ok {
							transaction.From = from
						}
						if to, ok := tx["to"].(string); ok {
							transaction.To = to
						}
						if amount, ok := tx["amount"].(string); ok {
							transaction.Amount = amount
						}

						if transaction.Hash != "" && transaction.From != "" && transaction.To != "" {
							transactions = append(transactions, transaction)
						}
					}
				}
			}
		}
	})

	return transactions
}

// 결과를 JSON 파일로 저장
func (c *BTCWebCrawler) SaveToFile(transactions []Transaction, filename string) error {
	data, err := json.MarshalIndent(transactions, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 마샬링 실패: %v", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// 결과를 콘솔에 출력
func (c *BTCWebCrawler) PrintTransactions(transactions []Transaction) {
	fmt.Printf("=== BTC 최근 전송 내역 (%d건) ===\n\n", len(transactions))

	if len(transactions) == 0 {
		fmt.Println("수집된 데이터가 없습니다.")
		fmt.Println("웹사이트 구조가 변경되었을 수 있습니다.")
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
		fmt.Printf("\n    금액: %s BTC\n", tx.Amount)
		if !tx.Timestamp.IsZero() {
			fmt.Printf("    시간: %s\n", tx.Timestamp.Format("2006-01-02 15:04:05"))
		}
		if tx.BlockNumber > 0 {
			fmt.Printf("    블록: %d\n", tx.BlockNumber)
		}
		fmt.Println()
	}
}

// 보조 함수들
func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func parseTime(timeStr string) (time.Time, error) {
	timeStr = strings.TrimSpace(timeStr)

	// 다양한 시간 형식 시도
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04",
		"Jan 2, 2006 3:04 PM",
		"Jan 2, 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	// 상대적 시간 처리
	if strings.Contains(timeStr, "ago") {
		if strings.Contains(timeStr, "hour") {
			return time.Now().Add(-time.Hour), nil
		} else if strings.Contains(timeStr, "minute") {
			return time.Now().Add(-time.Minute), nil
		} else if strings.Contains(timeStr, "second") {
			return time.Now().Add(-time.Second), nil
		}
		return time.Now(), nil
	}

	return time.Time{}, fmt.Errorf("시간 파싱 실패: %s", timeStr)
}

func parseBlockNumber(blockStr string) (int64, error) {
	blockStr = strings.TrimSpace(strings.Replace(blockStr, ",", "", -1))
	blockStr = strings.Replace(blockStr, "#", "", -1)

	var blockNum int64
	_, err := fmt.Sscanf(blockStr, "%d", &blockNum)
	return blockNum, err
}

func extractJSONFromScript(script string) string {
	// 간단한 JSON 추출 (실제로는 더 정교한 파싱 필요)
	start := strings.Index(script, "{")
	if start == -1 {
		return ""
	}

	braceCount := 0
	for i := start; i < len(script); i++ {
		if script[i] == '{' {
			braceCount++
		} else if script[i] == '}' {
			braceCount--
			if braceCount == 0 {
				return script[start : i+1]
			}
		}
	}

	return ""
}

func main() {
	fmt.Println("BTC 웹 크롤러 시작...")
	fmt.Println("Arkham Intel 웹사이트에서 BTC 최근 전송 내역을 수집합니다.\n")

	crawler := NewBTCWebCrawler()

	fmt.Println("데이터 수집 중...")
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
