///
/// @brief  : PostgreSQL TimescaleDB TEST
///           주어지 시간 안의 DATA를 생성한다.
/// @file   : timesc.go
/// @date   : 2025. 11. 14. (금) 14:57:35 KST
/// @author : Cento
///
package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"
	"sort"
	"github.com/lib/pq"
)

const (
	dbHost = "localhost"
	dbPort = 5432
	dbUser = "cento"
	dbPass = "cento"
	dbName = "CENTO"
)

var (
	currencys = []string{"EURO", "USDT"}

	minPrice = 1345.34
	maxPrice = 1567.98
)
type Row struct {
	dtime    time.Time
	currency string
	exrate   float64
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <start:YYYYMMDDHHMM> <end:YYYYMMDDHHMM> <count>\n", os.Args[0])
	os.Exit(1)
}

func main() {
	if len(os.Args) != 4 {
		usage()
	}

	startStr := os.Args[1]
	endStr := os.Args[2]
	countStr := os.Args[3]

	// 시각 파싱 (로컬 타임존, 포맷: YYYYMMDDHHMM)
	const inputLayout = "200601021504"

	startTime, err := time.ParseInLocation(inputLayout, startStr, time.Local)
	if err != nil {
		log.Fatalf("invalid start time: %v", err)
	}

	endTime, err := time.ParseInLocation(inputLayout, endStr, time.Local)
	if err != nil {
		log.Fatalf("invalid end time: %v", err)
	}

	if !startTime.Before(endTime) {
		log.Fatalf("start time must be before end time")
	}

	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 {
		log.Fatalf("invalid count: %v", err)
	}

	// Buffer 생성
	rows := make([]Row, 0, count)

	// 랜덤 시드 설정
	rand.Seed(time.Now().UnixNano())

	// DB 연결
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)
	fmt.Println("DB STRING : ", dsn)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping DB: %v", err)
	}

	log.Printf("Connected to PostgreSQL (TimescaleDB) at %s:%d DB=%s\n", dbHost, dbPort, dbName)
/*
public.TS_exrate (
    dtime     TIMESTAMP   NOT NULL,-- 시각 (타임존 없음)
    currency  VARCHAR(4)  NOT NULL,-- 통화 (USDT)
    exrate    FLOAT       NOT NULL,-- 환율
    timestr   VARCHAR(12)          -- 분단위 String (자동생성)
)
*/
	// INSERT 준비
	insertSQL := `INSERT INTO public.ts_exrate (dtime, currency, exrate) VALUES ($1, $2, $3)`
	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		log.Fatalf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	startUnix := startTime.Unix()
	endUnix := endTime.Unix()
	if endUnix <= startUnix {
		log.Fatalf("invalid time range")
	}
	rangeSec := endUnix - startUnix

	// Data 생성
	for i := 0; i < count; i++ {
		// 1) dtime 생성: 범위 내 랜덤 초
		//    0 <= delta <= rangeSec
		delta := rand.Int63n(rangeSec + 1)
		// 랜덤 1/10000초 (0 ~ 9999) → 나노초(100,000ns 단위)
		sub   := rand.Int63n(10000) * 100000
		dtime := time.Unix(startUnix+delta, sub)

		// 2) symbol 랜덤 선택
		currency := currencys[rand.Intn(len(currencys))]

		// 3) exrate 랜덤 생성 (minPrice ~ maxPrice)
		exrate := minPrice + rand.Float64()*(maxPrice-minPrice)
		// 소수점 둘째 자리로 반올림
		exrate = math.Round(exrate*100) / 100

		// 표준 출력: YYYYMMDDHHMMSS,symbol,price,flag
		rows = append(rows, Row{dtime, currency, exrate})
	}

	// sort
	sort.Slice(rows, func(i,j int) bool{
			return rows[i].dtime.Before(rows[j].dtime)})

	for _, r:= range rows {
		fmt.Printf("%s,%s,%.2f\n",
			r.dtime.Format("20060102150405"),
			r.currency,
			r.exrate)
		// DB INSERT
		_, err := stmt.Exec(r.dtime, r.currency, r.exrate)
		if err != nil {
			// 중복 (PRIMARY KEY 위반) 인지 확인
			if pgErr, ok := err.(*pq.Error); ok {
				if pgErr.Code == "23505" { // unique_violation
					log.Printf("duplicate(dtime=%s, symbol=%s). exit.\n",
						r.dtime.Format(time.RFC3339), r.currency)
					return
				}
			}
			log.Fatalf("failed to insert row: %v", err)
		}
	}
	log.Printf("Successfully generated and inserted %d records.\n", count)
}

