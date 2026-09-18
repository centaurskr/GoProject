package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
)

// HAR 분석으로 확인된 실제 API:
//   GET https://api.arkm.com/transfers
//     ?base=upbit&flow=all&usdGte=1&sortKey=time&sortDir=desc&limit=16&offset=N
//
// 특이사항:
//   - x-timestamp, x-payload 헤더 필요 (브라우저가 JS로 자동 생성)
//   - payload는 HMAC-SHA256 추정이나 secret key 미확인
//   - Playwright로 브라우저를 그대로 실행하면 헤더 자동 처리됨

type Transaction struct {
	ID              string
	TransactionHash string
	BlockTimestamp  string
	Chain           string
	BlockNumber     string
	FromAddress     string
	FromLabel       string
	FromEntity      string
	ToAddress       string
	ToLabel         string
	ToEntity        string
	TokenSymbol     string
	TokenName       string
	UnitValue       string
	HistoricalUSD   string
}

var (
	allTx   []Transaction
	seenIDs = make(map[string]bool)
	mu      sync.Mutex
)

func main() {
	if err := playwright.Install(); err != nil {
		log.Fatalf("playwright install: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("playwright.Run: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false), // 확인 후 true로 전환 가능
		Args:     []string{"--no-sandbox", "--disable-setuid-sandbox"},
	})
	if err != nil {
		log.Fatalf("browser launch: %v", err)
	}
	defer browser.Close()

	bctx, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 " +
				"(KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36",
		),
	})
	if err != nil {
		log.Fatalf("NewContext: %v", err)
	}

	page, err := bctx.NewPage()
	if err != nil {
		log.Fatalf("NewPage: %v", err)
	}

	// ── API 응답 인터셉트 ─────────────────────────────────────────
	page.On("response", func(resp playwright.Response) {
		if !strings.Contains(resp.URL(), "api.arkm.com/transfers") {
			return
		}
		if resp.Status() != 200 {
			log.Printf("[skip] status=%d url=%s", resp.Status(), resp.URL())
			return
		}
		body, err := resp.Body()
		if err != nil {
			log.Printf("[err] resp.Body: %v", err)
			return
		}
		n := parseAndAppend(body)
		mu.Lock()
		total := len(allTx)
		mu.Unlock()
		log.Printf("[+%d tx] total=%d  url=%s", n, total, resp.URL())
	})

	// ── 페이지 이동 ──────────────────────────────────────────────
	const target = "https://intel.arkm.com/explorer/entity/upbit"
	log.Printf("Goto %s", target)
	if _, err := page.Goto(target, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(30000),
	}); err != nil {
		log.Fatalf("Goto: %v", err)
	}
	time.Sleep(3 * time.Second)

	// ── 페이지네이션 자동 클릭 ───────────────────────────────────
	// selector 목록: 실제 페이지 Elements에서 확인 후 맞는 것 앞에 배치
	nextSelectors := []string{
		"button[aria-label='Next page']",
		"button[aria-label='next']",
		"button[aria-label='Go to next page']",
		"[data-testid='next-page']",
		"button:has-text('Next')",
		".pagination-next button",
		"li.next > a",
	}

	for pageNum := 1; ; pageNum++ {
		mu.Lock()
		total := len(allTx)
		mu.Unlock()
		log.Printf("── page %d | %d tx collected ──", pageNum, total)

		clicked := false
		for _, sel := range nextSelectors {
			btn, err := page.QuerySelector(sel)
			if err != nil || btn == nil {
				continue
			}
			disabled, _ := btn.GetAttribute("disabled")
			ariaDisabled, _ := btn.GetAttribute("aria-disabled")
			if disabled != "" || ariaDisabled == "true" {
				log.Println("Next button disabled → last page")
				goto done
			}
			if err := btn.Click(); err != nil {
				log.Printf("click err (%s): %v", sel, err)
				continue
			}
			time.Sleep(2 * time.Second)
			clicked = true
			break
		}
		if !clicked {
			log.Println("No next button → last page")
			break
		}
	}

done:
	mu.Lock()
	total := len(allTx)
	mu.Unlock()
	log.Printf("Done. %d unique transactions collected.", total)

	if err := saveCSV("upbit_transactions.csv", allTx); err != nil {
		log.Fatalf("saveCSV: %v", err)
	}
	log.Println("Saved → upbit_transactions.csv")
}

// ── JSON 파싱 ────────────────────────────────────────────────────
// HAR 확인 응답 구조:
//
//	transfers[i] = {
//	  id, transactionHash, blockTimestamp, chain, blockNumber,
//	  tokenSymbol, tokenName, unitValue, historicalUSD,
//	  fromAddress: { address, arkhamLabel:{name}, arkhamEntity:{name} },
//	  toAddress:   { address, arkhamLabel:{name}, arkhamEntity:{name} },
//	}
func parseAndAppend(body []byte) int {
	var envelope struct {
		Transfers []struct {
			ID              string  `json:"id"`
			TransactionHash string  `json:"transactionHash"`
			BlockTimestamp  string  `json:"blockTimestamp"`
			Chain           string  `json:"chain"`
			BlockNumber     int64   `json:"blockNumber"`
			TokenSymbol     string  `json:"tokenSymbol"`
			TokenName       string  `json:"tokenName"`
			UnitValue       float64 `json:"unitValue"`
			HistoricalUSD   float64 `json:"historicalUSD"`
			FromAddress     addrInfo `json:"fromAddress"`
			ToAddress       addrInfo `json:"toAddress"`
		} `json:"transfers"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		log.Printf("[parse err] %v", err)
		return 0
	}

	added := 0
	for _, t := range envelope.Transfers {
		mu.Lock()
		if seenIDs[t.ID] {
			mu.Unlock()
			continue
		}
		seenIDs[t.ID] = true
		allTx = append(allTx, Transaction{
			ID:              t.ID,
			TransactionHash: t.TransactionHash,
			BlockTimestamp:  t.BlockTimestamp,
			Chain:           t.Chain,
			BlockNumber:     fmt.Sprintf("%d", t.BlockNumber),
			FromAddress:     t.FromAddress.Address,
			FromLabel:       t.FromAddress.labelName(),
			FromEntity:      t.FromAddress.entityName(),
			ToAddress:       t.ToAddress.Address,
			ToLabel:         t.ToAddress.labelName(),
			ToEntity:        t.ToAddress.entityName(),
			TokenSymbol:     t.TokenSymbol,
			TokenName:       t.TokenName,
			UnitValue:       fmt.Sprintf("%g", t.UnitValue),
			HistoricalUSD:   fmt.Sprintf("%.2f", t.HistoricalUSD),
		})
		mu.Unlock()
		added++
	}
	return added
}

type addrInfo struct {
	Address     string `json:"address"`
	ArkhamLabel *struct {
		Name string `json:"name"`
	} `json:"arkhamLabel"`
	ArkhamEntity *struct {
		Name string `json:"name"`
	} `json:"arkhamEntity"`
}

func (a addrInfo) labelName() string {
	if a.ArkhamLabel != nil {
		return a.ArkhamLabel.Name
	}
	return ""
}

func (a addrInfo) entityName() string {
	if a.ArkhamEntity != nil {
		return a.ArkhamEntity.Name
	}
	return ""
}

// ── CSV 저장 ─────────────────────────────────────────────────────
func saveCSV(filename string, txs []Transaction) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	_ = w.Write([]string{
		"BlockTimestamp", "TransactionHash", "Chain", "BlockNumber",
		"FromAddress", "FromLabel", "FromEntity",
		"ToAddress", "ToLabel", "ToEntity",
		"TokenSymbol", "TokenName", "UnitValue", "HistoricalUSD",
	})
	for _, tx := range txs {
		_ = w.Write([]string{
			tx.BlockTimestamp, tx.TransactionHash, tx.Chain, tx.BlockNumber,
			tx.FromAddress, tx.FromLabel, tx.FromEntity,
			tx.ToAddress, tx.ToLabel, tx.ToEntity,
			tx.TokenSymbol, tx.TokenName, tx.UnitValue, tx.HistoricalUSD,
		})
	}
	return w.Error()
}
