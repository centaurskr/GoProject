// 
// 거래소 Hot Wallet 주소 White List 구축 Project
// Blockchain.info의 WebSocket을 사용하여 실시간으로 새로운 블록 해시를 
// 받아오고, 해당 블록의 상세 정보를 조회하여 
// 업비트(Upbit)로 추정되는 주소를 필터링하는  구조
//
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const wsURL = "wss://ws.blockchain.info/inv"

// 멤풀 트랜잭션 구조체 (필요한 주소 정보만 정의)
type MempoolTx struct {
	Op string `json:"op"`
	X  struct {
		Inputs []struct {
			PrevOut struct {
				Addr string `json:"addr"` // 보낸 주소 (From)
			} `json:"prev_out"`
		} `json:"inputs"`
		Out []struct {
			Addr string `json:"addr"` // 받는 주소 (To)
		} `json:"out"`
	} `json:"x"`
}

func main() {
	// 1. 웹소켓 연결
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal("연결 실패:", err)
	}
	defer conn.Close()

	// 2. 미확정 거래(Mempool) 구독 설정
	subscribeMsg := `{"op":"unconfirmed_sub"}`
	if err := conn.WriteMessage(websocket.TextMessage, []byte(subscribeMsg)); err != nil {
		log.Fatal("구독 요청 실패:", err)
	}

	fmt.Println("📡 비트코인 멤풀 모니터링 시작... (60초에 한 건씩 분석)")
	fmt.Println("--------------------------------------------------")

	for {
		// 웹소켓으로부터 메시지 수신
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("수신 에러:", err)
			return
		}

		// JSON 파싱
		var tx MempoolTx
		if err := json.Unmarshal(message, &tx); err != nil {
			continue
		}

		// 3. 'From' 주소 추출 및 출력
		if len(tx.X.Inputs) > 0 {
			fmt.Printf("\n[신규 거래 포착] 시간: %s\n", time.Now().Format("15:04:05"))
			fmt.Printf("Input : %v\n", tx.X.Inputs)
			fmt.Println("From 주소 목록:")
			
			for i, input := range tx.X.Inputs {
				fromAddr := input.PrevOut.Addr
				if fromAddr == "" {
					fromAddr = "(알 수 없는 주소 - 주로 코인베이스 거래 등)"
				}
				fmt.Printf("  [%d] %s\n", i+1, fromAddr)
			} // for i

		} // IF

		// 사용자 요청대로 하나씩 확인하기 위해 60초 대기
		fmt.Println("--------------------------------------------------")
		fmt.Println("☕ 다음 거래 분석을 위해 60초 대기 중...")
		time.Sleep(60 * time.Second)
	} // Main Loop
}
