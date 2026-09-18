// main.go — goTbmon 진입점
//
// goTbmon: goTbnova 프로세스 관리 데몬용 터미널 인터랙티브 모니터
//
// 실행 흐름:
//   main() → NewModel() → tea.NewProgram().Run()
//
// 명령행 옵션:
//   -addr  goTbnova TCP 주소 (기본: 127.0.0.1:30001)
//   -v     버전 출력 후 종료
package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

const version = "1.0.0"

// main: CLI 플래그를 파싱하고 bubbletea 프로그램을 실행한다.
// 대체 화면(AltScreen)을 사용하여 기존 터미널 내용을 보존한다.
func main() {
	host := flag.String("h", "localhost", "goTbnova 호스트 (기본: localhost)")
	ver := flag.Bool("v", false, "버전 출력")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "사용법: goTbmon [옵션]\n\n옵션:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n조작키:\n")
		fmt.Fprintf(os.Stderr, "  f           목록 재갱신\n")
		fmt.Fprintf(os.Stderr, "  s / k       선택 프로세스 시작 / 종료\n")
		fmt.Fprintf(os.Stderr, "  t           goTbnova Process Table Reload\n")
		fmt.Fprintf(os.Stderr, "  h           프로세스 상세 정보\n")
		fmt.Fprintf(os.Stderr, "  g / d       그룹 트리 / 의존성 트리\n")
		fmt.Fprintf(os.Stderr, "  e           전역 환경변수 조회\n")
		fmt.Fprintf(os.Stderr, "  Space/a/u   개별/전체 선택·해제\n")
		fmt.Fprintf(os.Stderr, "  ↑↓ PgUp PgDn Home End  커서 이동\n")
		fmt.Fprintf(os.Stderr, "  /           ID·그룹·설명 검색\n")
		fmt.Fprintf(os.Stderr, "  ?           도움말\n")
		fmt.Fprintf(os.Stderr, "  q           종료\n")
	}
	flag.Parse()

	if *ver {
		fmt.Printf("goTbmon v%s\n", version)
		os.Exit(0)
	}

	addr := *host + ":30001"
	m := NewModel(*host, addr)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		os.Exit(1)
	}
}
