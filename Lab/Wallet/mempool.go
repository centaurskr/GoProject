//
// mempool txid 수집
// 60초마다:
//    mempool txid 목록 조회
//    신규 txid만 필터링
//    tx 상세 조회
//    input address 추출
//
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const baseURL = "https://mempool.space/api"

var seenTx = make(map[string]bool)

type MempoolTx struct {
	TxID string `json:"txid"`
}

type TxDetail struct {
	TxID string `json:"txid"`
	Vin  []struct {
		Prevout struct {
			ScriptPubKeyAddress string `json:"scriptpubkey_address"`
			Value               int64  `json:"value"`
		} `json:"prevout"`
	} `json:"vin"`
}

func fetchRecentTxs() ([]MempoolTx, error) {
	resp, err := http.Get(baseURL + "/mempool/recent")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var txs []MempoolTx
	err = json.NewDecoder(resp.Body).Decode(&txs)
	return txs, err
}

func fetchTxDetail(txid string) (*TxDetail, error) {
	url := fmt.Sprintf("%s/tx/%s", baseURL, txid)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tx TxDetail
	err = json.NewDecoder(resp.Body).Decode(&tx)
	return &tx, err
}

func processTx(txid string) {

	if seenTx[txid] {
		return
	}
	seenTx[txid] = true

	tx, err := fetchTxDetail(txid)
	if err != nil {
		log.Println("fetch error:", err)
		return
	}

	fmt.Printf("\n[신규 거래 포착] 시간: %s\n", time.Now().Format("15:04:05"))
	fmt.Printf("TX: %s\n", tx.TxID)

	for _, vin := range tx.Vin {
		addr := vin.Prevout.ScriptPubKeyAddress
		value := vin.Prevout.Value
		fmt.Printf("  FROM: %s (%d sats)\n", addr, value)
	}
	fmt.Println("--------------------------------------------------")
}

func main() {

	for {
		txs, err := fetchRecentTxs()
		if err != nil {
			log.Println("mempool fetch error:", err)
			time.Sleep(30 * time.Second)
			continue
		}

		for _, tx := range txs {
			processTx(tx.TxID)
		}

		time.Sleep(60 * time.Second)
	}
}
