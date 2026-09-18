# goTbnova/tlog

goTbnova 전용 로거 — `centolib/netlog`(네트워크) 또는 로컬 파일 중 하나를 백엔드로 선택하는 이중 백엔드 로거.

## Import

```go
import "goTbnova/tlog"
```

## 개요

`Backend` 인터페이스(`log/close/rotateIfNeeded`)로 두 구현체를 추상화한다: `netBackend`([centolib/netlog](../../centolib/netlog/README.md) 래핑, 기본값) / `fileBackend`(로컬 텍스트 파일, 일자별 자동 회전 및 버퍼링 flush). 모든 로그는 `event=` 필드를 포함하는 구조화 포맷을 따른다.

```
goTbnova -log=net  <table>  — centolib/netlog → Vector → VictoriaLogs (기본값)
goTbnova -log=file <table>  — 로컬 파일 직접 기록 (스탠드얼론 환경)
```

로컬 파일 포맷 예:
```
2025/01/04 13:32:58.123 [INFO ] [process_start    ] id=AGENT pid=1234 count=1 mode=P
```

## API

### 초기화/종료

- `func Init(mode LogMode, table, logDir string) error` — `LogMode`(`ModeNet`/`ModeFile`)에 따라 백엔드를 초기화한다. `sync.Once`로 프로세스당 1회만 유효. `ModeNet`에서 Vector 소켓이 없으면 명확한 에러 메시지와 함께 실패(호출자는 `os.Exit(1)` 해야 함). `ModeFile`은 파일/디렉토리 생성 실패 시 에러.
- `func InitWithSocket(mode LogMode, table, logDir, socketPath string) error` — `socketPath`가 비어있지 않으면 `vectorSocketPath`를 먼저 덮어쓴 뒤 `Init()`과 동일하게 동작(`gotbnova.toml [log] vector_socket` 적용용).
- `func Close()` — 백엔드 flush 및 종료(`main()`의 `defer`로 호출).
- `func RotateIfNeeded()` — 일자 변경 시 로그 파일 교체(`fileBackend`만 동작, `netBackend`는 no-op). `process/manager.go`의 tick 루프에서 호출.

### 이벤트 전용 헬퍼 (내부적으로 `event=` 필드가 포함된 구조화 메시지를 만들어 전송)

| 함수 | 레벨 | 용도 |
|---|---|---|
| `Startup(table string, port int)` | INFO | 기동 |
| `Shutdown(reason string)` | INFO | 종료 |
| `DateChanged(weekday int, date string)` | INFO | 일자 변경 |
| `TableLoad(path string, procCnt, groupCnt int)` | INFO | 최초 테이블 로드 |
| `TableReload(path string, added, removed, modified int)` | INFO | 리로드 완료 |
| `TableReloadFail(path string, err error)` | ERROR | 리로드 실패(기존 설정 유지) |
| `ProcessStart(id string, pid, count int, mode, path string)` | INFO | 프로세스 기동 |
| `ProcessExit(id string, pid int, reason string)` | INFO | 정상 종료 |
| `ProcessAbnormal(id string, pid, failCount int, exitErr string)` | ERROR | 비정상 종료 (Grafana 알림 트리거 대상) |
| `ProcessStopReq(id string, pid, graceSec int)` | INFO | 종료 요청(SIGTERM 발송) |
| `ProcessBackoff(id string, failCount, delaySec int, nextRetry string)` | WARN | Back-off 대기 진입 |
| `ProcessDisabled(id string, maxCount int)` | ERROR | MaxCount 초과로 비활성화 |
| `ProcessWaitDepend(id, waitFor string)` | INFO | 종속 프로세스 대기 시작 |
| `ProcessCascade(sourceID, targetID, action string)` | WARN | 연쇄 종료 처리 |
| `MonitorConnect(addr string)` | INFO | 모니터 클라이언트 접속 |
| `MonitorCmd(cmd, id, result string)` | INFO | 모니터 제어 명령 수신 |

### 범용 로그 (위 이벤트 함수로 표현하기 어려운 경우만 사용)

- `func Info/Error/Warn/Debug(format string, v ...interface{})` — `Printf` 형식.

### 유틸리티

- `func StopReasonStr(reason int) string` — `model.StopReason` 정수값(0, -1~-7)을 `"none"/"operator"/"holiday"/"restart"/"time_up"/"reload"/"abnormal"/"max_count"` 문자열로 변환. 로그 출력용.

### 타입

- `LogMode string` — `ModeNet("net")`(기본) / `ModeFile("file")`.

## 의존성

`centolib/netlog`(net 모드), 표준 라이브러리(`bufio, os, time` 등, file 모드).

## 환경설정

- Vector UDS 소켓 기본 경로: `/tmp/VECTOR.UNIX.DGRAM`(패키지 내부 기본값, `InitWithSocket`으로 오버라이드).
- 파일 모드 로그 파일명: `{logDir}/goTbnova_YYYYMMDD.log`, `O_APPEND`로 열어 재기동 시에도 이어서 기록.

## 주의사항

- `Init`/`InitWithSocket`은 `sync.Once`로 보호되는 **프로세스 전역 싱글턴**이다 — 두 번째 호출은 무시된다(모드를 바꿔 재초기화할 수 없음).
- `ModeNet`에서 Vector 소켓이 없으면 **기동이 즉시 실패**한다(폴백으로 자동으로 파일/stderr 모드로 전환되지 않는다) — `stderrBackend`는 코드상 존재하지만 `Init()` 경로에서 자동 선택되지 않고, 초기화 실패는 호출자(`main()`)가 처리해야 한다.
- `fileBackend`는 ERROR/WARN을 즉시 flush하고 INFO/DEBUG는 2초 주기로만 flush한다 — 비정상 종료(프로세스 크래시 등) 직전의 INFO 로그는 최대 2초까지 유실될 수 있다.
