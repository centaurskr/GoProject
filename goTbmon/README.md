# goTbmon

`goTbnova` 프로세스 관리 데몬용 Bubble Tea 기반 터미널 인터랙티브 모니터.

## 개요

- `module goTbmon`, `go 1.24.5`. `github.com/charmbracelet/bubbletea`/`bubbles` 기반 TUI. [goTbnova/control](../goTbnova/control/README.md)의 TCP JSON 프로토콜 클라이언트다.
- 파일 구성: `main.go`(CLI 플래그+부트스트랩), `client.go`(TCP 요청/응답 클라이언트, 매 요청마다 새 연결을 열고 닫는 온디맨드 방식), `model.go`(goTbnova 프로토콜 JSON 구조체 및 상태코드 → 표시 문자열 변환), `ui.go`(Bubble Tea `Model`/`Update`/`View`, 목록·상세·그룹트리·의존성트리·환경변수·도움말 화면), `ansi.go`(커서 이동/256색 ANSI 헬퍼).

## 빌드

```bash
cd goTbmon
go build -o goTbmon
# 또는
./build.sh   # 빌드 + strip + ~/bin 으로 이동
```

## 사용법

```bash
goTbmon [-h host] [-v]
```

| 옵션 | 기본값 | 설명 |
|---|---|---|
| `-h <host>` | `localhost` | goTbnova가 떠 있는 호스트. 포트는 항상 `:30001`이 자동으로 붙는다(`host+":30001"`) |
| `-v` | — | 버전 출력 후 종료 |

### 조작키

| 키 | 동작 |
|---|---|
| `f` | 목록 재갱신 |
| `s` / `k` | 선택 프로세스 시작 / 종료 |
| `t` | goTbnova Process Table Reload |
| `h` | 프로세스 상세 정보 |
| `g` / `d` | 그룹 트리 / 의존성 트리 |
| `e` | 전역 환경변수 조회 |
| `Space`/`a`/`u` | 개별/전체 선택·해제 |
| `↑↓ PgUp PgDn Home End` | 커서 이동 |
| `/` | ID·그룹·설명 검색 |
| `?` | 도움말 |
| `q` | 종료 |

## 환경설정

접속 주소 외 별도 설정 파일/환경변수는 없다. 접속 대상은 `goTbnova`의 모니터 TCP 서버([goTbnova/control](../goTbnova/control/README.md), 기본 `127.0.0.1:30001`, `/etc/services`의 `tbnova/tcp`)이다.

## 주의사항

- 상태 코드(`State`)와 종료 원인(`StopReason`) 정수값은 [goTbnova/model](../goTbnova/model/README.md)의 `ProcessState`/`StopReason`과 **하드코딩된 정수 매핑**으로 동기화되어 있다(`model.go`의 `stateName`/`stopReasonName`) — goTbnova 쪽에서 enum 값이나 순서가 바뀌면 이 파일도 함께 수정해야 한다(공용 타입을 import하지 않고 별도로 값 복제).
- `client.go`는 요청마다 새 TCP 연결을 맺고 닫는다(상시 연결 없음) — 연결/읽기/쓰기 타임아웃은 각각 3초/5초로 고정되어 있어 goTbnova 응답이 느리면 타임아웃 에러가 뜬다.
- `main.go` 파일 상단 주석에는 `-addr` 옵션(기본값 `127.0.0.1:30001`)으로 적혀 있으나, **실제 구현된 플래그는 `-h <host>`**이며 포트는 `:30001`로 고정 부착된다 — 주석과 코드가 다르므로 실제 사용 시 `-h`를 사용할 것.
