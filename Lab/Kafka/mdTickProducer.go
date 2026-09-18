///
/// @brief  : Kafka md.tick을 자동생성하는 Producer Program
///           kafka server : 192.168.0.6 (single node)
/// @file   : mdTickProducer.go
/// @date   : 2025. 11. 20. (목) 14:57:57 KST
/// @author : Cento
///
package main

import(
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

type Tick struct{
	Symbol   string  `jsong:"S"`
	Price    string  `json:"P"`
	Qty      string  `json:"Q"`
	Timestamp string  `json:"T"`
}

func main(){

	brokers := []string{"192.168.0.6:9092"}
	topic := "md.tick"

	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)

	// Producer 설정
	config := sarama.NewConfig()
	config.Producer.RequiredAcks    = sarama.WaitForAll  // 모든 replica ACK
	config.Producer.Retry.Max       = 3      // 재시도 횟수
	config.Producer.Return.Successes = true   // 성공 결과 수신
	config.Producer.Partitioner = sarama.NewHashPartitioner // key 기반 파티션닝

	// Sync Producer 생성
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Failed to start Sarama producer : %v", err)
	}

	defer func() {
		if err := producer.Close(); err != nil{
			log.Printf("Failed to close producer : %v", err)
		}
	}()

	log.Println("Kafka md.tick 테스트 Producer 시작")

	// 종료 시그널 처리 (CTRL-C)
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)

	// 랜덤 seed
	rand.Seed(time.Now().UnixNano())

	// Symbol List
	symbols := []string{"BTC.USDT", "ETH.USDT", "XRP,USDT", "SOL.USDT", "BNB.USDT"}
	
	ticker := time.NewTicker(1 * time.Second) // 1초에 한 번
	defer ticker.Stop()

loop :
	for {
		select {
		case t := <-ticker.C:
			sym := symbols[rand.Intn(len(symbols))]
			price := 10000 + rand.Float64() * 1000
			qty := int64(1 + rand.Intn(100))

			tick := Tick{
				Symbol:    sym,
				Price:     fmt.Sprintf("%.8f", price),
				Qty:       fmt.Sprintf("%d", qty),
				Timestamp: t.Format("20060102150405.000000"),
			}

			jsonstr, err := json.Marshal(tick)
			if err != nil{
				log.Printf("Json.Marshal error : %v", err)	
				continue
			}

			// symbol을 key로 사용( 같은 심볼은 같은 파티션으로)
			msg := &sarama.ProducerMessage{
				Topic: topic,
				Key: sarama.StringEncoder(sym),
				Value: sarama.ByteEncoder(jsonstr),
			}

			partition, offset, err := producer.SendMessage(msg)
			if err != nil {
				log.Printf("SendMessage error : %v", err)
				continue
			}


			log.Printf("Produced tick : %s | price : %.2f qty = %d | partitions=%d, offset=%d",
				sym,price,qty, partition, offset)

		case <-sigchan :
			log.Println("종료 시그널 수신, Producer 종료")
			break loop
		} // switch 
	} // For loop 
	fmt.Println("프로그램 종료")
}
