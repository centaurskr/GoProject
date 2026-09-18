# centolib/screen

터미널 raw mode 전환과 ANSI 이스케이프 시퀀스 기반 화면/커서/색상 제어 유틸리티.

## Import

```go
import "centolib/screen"
```

## 개요

`goPlus/Ops`의 콘솔 메뉴형 운영 도구(`PlatformCheck` 등)처럼 커스텀 텍스트 UI를 그리는 도구를 위한 저수준 터미널 제어 헬퍼. 키 입력 상수(특수키/화살표/기능키/Ctrl/Alt 조합), 화면 지우기, 커서 이동, 색상 등을 제공한다. (참고: `goTbmon`은 이 패키지 대신 `github.com/charmbracelet/bubbletea` 기반 TUI 프레임워크를 사용한다.)

## API

### 키 입력

- `func GetKey() (string, error)` — stdin에서 최대 6바이트를 읽어 키 시퀀스 문자열로 반환(특수키는 여러 바이트 escape sequence).
- `func SetRawMode() (func(), error)` — 터미널을 raw mode(ICANON/ECHO 해제)로 전환하고, 원상복구용 `restore func()`를 반환한다. 사용 후 `defer restore()` 형태로 호출.
- `func GETWINDOWSIZE() (int, int, error)` — `ioctl(TIOCGWINSZ)`로 터미널 크기를 (열, 행) 순서로 반환.

### 키 상수

`ESC, ENTER, SPACE, TAB, F1~F10, INS, DEL, HOME, END, PGUP, PGDN, UPARR/DNARR/RTARR/LTARR`(화살표), `CTR_A~CTR_Z`(Ctrl 조합), `ALT_A~ALT_Z`(Alt 조합) — 모두 `GetKey()`가 반환할 수 있는 원시 이스케이프 시퀀스 문자열 상수.

### 화면 제어 함수 (모두 인자 없이 stdout에 이스케이프 코드를 출력)

| 함수 | 동작 |
|---|---|
| `CLEAR_WIN()` | 화면 지우기 |
| `ECHO_OFF()` / `ECHO_ON()` | 로컬 에코 끄기/켜기 |
| `KEY_LOCK()` / `KEY_UNLOCK()` | 키보드 잠금/해제 |
| `PAD_NUM()` / `PAD_APP()` | 넘버패드 모드 |
| `TERM_RESET()` | 터미널 리셋 |
| `COL_132()` / `COL_80()` | 132/80컬럼 모드 |
| `WARAP_ON()` / `WARAP_OFF()` | 자동 줄바꿈 on/off |
| `REPEAT_ON()` / `REPEAT_OFF()` | 키 반복 on/off |
| `CURSOR_HOME()` | 커서를 (1,1)로 |
| `CURSOR_MOVE(x, y int)` | 커서를 (x,y)로 이동 |
| `CURSOR_ON()` / `CURSOR_OFF()` | 커서 표시/숨김 |
| `UNDERLINE_CURSOR()` / `BLOCK_CURSOR()` | 커서 모양 |
| `SAVE_CURSOR()` / `RESTORE_CURSOR()` | 커서 위치 저장/복원 |
| `ATTR_NONE/BOLD/DIM/UNDERLINE/BLINK/REVERSE/NORMAL/DIMRESET/UNDERRESET/BLINKRESET/REVRESET()` | 텍스트 속성 |
| `SET_TOPBOTTOM(t, b int)` | 스크롤 영역 지정 |
| `SCREEN_UP()` / `SCREEN_DOWN()` | 화면 스크롤 |
| `CLEAR_LINE_FROMCUR()` / `CLEAR_LINE_TOCUR()` / `CLEAR_LINE()` | 줄 지우기(커서 이후/이전/전체) |
| `COLOR_RED/GREEN/YELLOW/BLUE/MAGENTA/SCYAN/WHITE/RESET()` | 전경색 |

## 의존성

표준 라이브러리만 사용(`syscall`, `unsafe`).

## 주의사항

- `syscall.Syscall(SYS_IOCTL, ...)`을 직접 호출하므로 Linux 전용이다.
- `SetRawMode()`로 전환한 뒤에는 프로그램이 비정상 종료(패닉 등)하면 터미널이 raw 상태로 남을 수 있다 — 복구 함수 호출을 보장하도록 `defer`/시그널 핸들러에서 처리할 것.
- 화면제어 함수들 중 일부(`ECHO_ON`, `COL_80`, `WARAP_OFF`, `REPEAT_OFF`, `BLOCK_CURSOR`)는 이스케이프 시퀀스에 `I`(대문자 아이)를 쓰고 있어 표준 ANSI의 `l`(소문자 엘, low 계열 리셋)과 다르다 — 코드 그대로를 근거로 문서화했으며 실제 터미널 호환성은 별도 검증이 필요하다.
