# centolib/netlog

Unix Domain Socket(Datagram)을 통해 Vector 에이전트 → VictoriaLogs로 비동기 전송하는 네트워크 로거. `goPlus`의 `NetLog` 빌드와 `goTbnova/tlog`(`-log=net` 기본 모드)가 사용하는 백엔드다.

## Import

```go
import Log "centolib/netlog"
```

## 개요

전사 표준 스키마(JSON)로 로그를 만들어 Unix datagram 소켓(`/tmp/VECTOR.UNIX.DGRAM`)으로 전송한다. 내부적으로 고루틴 1개 + 버퍼 채널(`BufferSize=10000`)로 논블로킹 전송하며, 채널이 가득 차면 로그를 드롭한다. C/Go/Java/Node.js 등 여러 언어의 로그를 동일 스키마로 VictoriaLogs에 통합 수집하기 위해 설계되었다. 동급 로컬 로거는 [centolib/logout](../logout/README.md)이며 API가 다르다.

## API

### `func InitNetLogout(pname string, pextra string) error`
- 설명: `sync.Once`로 1회만 초기화되는 전역 로거를 생성한다. Unix datagram 소켓(`/tmp/VECTOR.UNIX.DGRAM`)에 연결하고 백그라운드 전송 워커 고루틴을 기동한다.
- 파라미터: `pname` — 프로세스 이름, `pextra` — 프로세스 추가 식별자(로그 필드 `extra`).
- 반환값: 소켓 연결 실패 시 에러.
- 사용 예 (`goPlus/Matching/NetLog/MatMain.go`):
  ```go
  err := Log.InitNetLogout(os.Args[0], os.Args[1])
  ```

### `func Close()`
- 설명: `done` 채널을 닫아 워커에게 종료를 알리고, 남은 채널의 로그를 모두 소켓으로 flush한 뒤 워커 종료를 대기한다(`sync.WaitGroup`). 프로그램 종료 시 로그 유실을 막기 위해 `defer netlog.Close()`로 사용.

### 레벨별 로그 함수 (모두 `Printf` 형식)

- `func Infor(format string, v ...interface{})`
- `func Error(format string, v ...interface{})`
- `func Debug(format string, v ...interface{})`
- `func Warn(format string, v ...interface{})`

각 호출은 `runtime.Caller(2)`로 호출자의 파일명/함수명/라인을 자동 채워 `LogMessage`(JSON: `ts, level, pname, extra, source, function, line, message`)로 만들고, 채널에 논블로킹 전송한다(`InitNetLogout` 미호출 시 아무 동작 안 함).

## 의존성

표준 라이브러리만 사용(`net`, `encoding/json` 등). Vector 에이전트가 `/tmp/VECTOR.UNIX.DGRAM`에서 datagram을 수신하고 있어야 실제로 로그가 VictoriaLogs까지 전달된다.

## 환경설정

- 소켓 경로는 상수 `SocketPath = "/tmp/VECTOR.UNIX.DGRAM"`로 고정(패키지 자체에는 오버라이드 API가 없다 — `goTbnova/tlog`처럼 상위 레이어에서 별도 소켓 경로 설정을 원하면 자체 초기화 로직으로 감싸야 한다).
- 전송 대상 VictoriaLogs/Vector 파이프라인은 이 저장소 밖의 인프라이므로, 운영 시 Vector 에이전트가 해당 UDS를 listen 중인지 사전 확인이 필요하다.

## 주의사항

- `sync.Once`로 보호되는 프로세스 전역 싱글턴이다 — `InitNetLogout`을 여러 이름으로 재호출해도 최초 호출의 `pname`/`pextra`만 유효하다.
- 채널이 가득 차면(초당 로그량이 매우 많을 때) 로그가 조용히 드롭된다 — 손실 없는 로깅이 필요한 경로에는 적합하지 않다.
- Vector 소켓이 없거나 죽어 있으면 `InitNetLogout`이 에러를 반환한다 — 호출부에서 에러 처리(재시도/폴백) 필요.
