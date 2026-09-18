// ansi.go — ANSI 이스케이프 시퀀스 상수 및 헬퍼 함수
//
// 터미널 색상·커서 제어에 필요한 이스케이프 시퀀스를 상수와 함수로 제공한다.
// bubbletea가 렌더링 문자열을 그대로 터미널에 쓰므로, View() 내부에서
// 이 패키지의 상수를 조합해 ANSI 색상 코드를 삽입한다.
//
// 주요 카테고리:
//   커서 제어 (moveTo, cursorHome, cursorHide …)
//   색상     (fg256, bg256, fgRed, bgBlue …)
//   화면     (clearScreen, clearLine …)
package main

import "fmt"

// ANSI 이스케이프 시퀀스 상수
const (
	esc = "\x1b["

	// 커서 제어
	cursorHome    = esc + "H"
	cursorSave    = esc + "s"
	cursorRestore = esc + "u"
	cursorHide    = esc + "?25l"
	cursorShow    = esc + "?25h"

	// 화면 지우기
	clearScreen = esc + "2J"
	clearLine   = esc + "2K"
	clearDown   = esc + "J"

	// 대체 화면
	altScreenOn  = esc + "?1049h"
	altScreenOff = esc + "?1049l"

	// 색상 리셋
	reset = esc + "0m"

	// 굵기
	bold = esc + "1m"

	// 전경색 (256 색상)
	fgBlack   = esc + "30m"
	fgRed     = esc + "31m"
	fgGreen   = esc + "32m"
	fgYellow  = esc + "33m"
	fgBlue    = esc + "34m"
	fgMagenta = esc + "35m"
	fgCyan    = esc + "36m"
	fgWhite   = esc + "37m"
	fgDefault = esc + "39m"

	// 배경색
	bgBlack   = esc + "40m"
	bgRed     = esc + "41m"
	bgGreen   = esc + "42m"
	bgYellow  = esc + "43m"
	bgBlue    = esc + "44m"
	bgMagenta = esc + "45m"
	bgCyan    = esc + "46m"
	bgWhite   = esc + "47m"
	bgDefault = esc + "49m"

	// 밝은 전경색
	fgBrightBlack   = esc + "90m"
	fgBrightRed     = esc + "91m"
	fgBrightGreen   = esc + "92m"
	fgBrightYellow  = esc + "93m"
	fgBrightBlue    = esc + "94m"
	fgBrightMagenta = esc + "95m"
	fgBrightCyan    = esc + "96m"
	fgBrightWhite   = esc + "97m"

	// 밝은 배경색
	bgBrightBlack = esc + "100m"
	bgBrightWhite = esc + "107m"

	// 반전
	reverseVideo = esc + "7m"
)

// 커서를 지정 위치로 이동 (1-based)
func moveTo(row, col int) string {
	return fmt.Sprintf("%s%d;%dH", esc, row, col)
}

// 256색 전경
func fg256(n int) string {
	return fmt.Sprintf("%s38;5;%dm", esc, n)
}

// 256색 배경
func bg256(n int) string {
	return fmt.Sprintf("%s48;5;%dm", esc, n)
}

// 컬럼 수만큼 수평선 문자 반복
func hline(width int) string {
	s := ""
	for i := 0; i < width; i++ {
		s += "─"
	}
	return s
}
