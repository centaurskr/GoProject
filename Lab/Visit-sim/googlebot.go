package main

import (
	"os"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	// 1. Chrome 옵션: headless=false로 하면 실제 창이 뜸
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // 디버깅용으로 창 보이게
		chromedp.Flag("disable-gpu", false),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 2. 검색어
	keyword := "연결지향통신"

	// 3. 검색 결과 링크 저장 변수
	var firstResultURL string

	// 4. chromedp 작업 정의
	tasks := chromedp.Tasks{
		// Naver 접속
		chromedp.Navigate("https://www.naver.com"),

		// 검색창 보일 때까지 대기
		chromedp.WaitVisible(`input#query`, chromedp.ByQuery),

		// 검색어 입력 + Enter
		chromedp.SendKeys(`input#query`, keyword+"\n", chromedp.ByQuery),

		// 검색결과 로드 대기
		chromedp.Sleep(2 * time.Second),

		// 첫 번째 결과 링크 가져오기
		chromedp.AttributeValue(`a.api_txt_lines`, "href", &firstResultURL, nil, chromedp.ByQuery),

		// 첫 번째 결과 클릭
		chromedp.Click(`a.api_txt_lines`, chromedp.ByQuery),

		// 페이지 로드 대기
		chromedp.Sleep(2 * time.Second),
	}

	// 5. 작업 실행
	if err := chromedp.Run(ctx, tasks); err != nil {
		log.Fatal(err)
		os.Exit(0)
	}

	// 6. 결과 출력
	fmt.Println("첫 번째 결과 URL:", firstResultURL)
}

