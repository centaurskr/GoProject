package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	// Chrome 실행 컨텍스트 생성
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// 실행 타임아웃 (10초)
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 결과 저장 변수
	var pageTitle string

	// chromedp 작업 실행
	err := chromedp.Run(ctx,
		// 페이지 열기
		chromedp.Navigate(`https://centaurskr.tistory.com`),
		//chromedp.Navigate(`https://example.com`),

		// 페이지 완전히 로드될 때까지 대기
		chromedp.WaitReady("body", chromedp.ByQuery),

		// 페이지 제목 가져오기
		chromedp.Title(&pageTitle),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("방문한 페이지 제목:", pageTitle)
}

