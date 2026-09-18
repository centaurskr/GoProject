# centolib/logout

`go.uber.org/zap` 기반 로컬(콘솔/파일) 로거. `goPlus`의 `LocalLog` 빌드가 사용하는 로깅 백엔드다.

## Import

```go
import Log "centolib/logout"
```

## 개요

콘솔 출력, 파일 출력(자동 회전), 또는 둘 다를 지원하는 레벨 로거. 콘솔 모드에서는 레벨별 색상이 적용된다(INFOR=파랑, DEBUG=마젠타, WARN=노랑, ERROR=빨강 계열). 파일 출력은 `gopkg.in/natefinch/lumberjack.v2`로 100MB/5개 백업/30일 보관/gzip 압축 회전한다. 네트워크로 전송하는 동급 패키지는 [centolib/netlog](../netlog/README.md)이며, 둘은 API가 다르다.

## API

### `func InitLogger(ext string, mode int)`
- 설명: 로거를 초기화한다. **반드시 다른 로그 함수보다 먼저 1회 호출**해야 한다.
- 파라미터:
  - `ext` — 로그파일 이름에 붙일 구분 문자열. 같은 이름의 프로세스를 여러 개 띄울 때 구분 용도. `""`이면 `{프로세스명}.log`, 아니면 `{프로세스명}_{ext}.log`.
  - `mode` — 출력 대상: `CONSOLE_MODE`(stdout만), `FILE_MODE`(파일만, 운영모드), `ALL_MODE`(둘 다).
- 환경설정: 파일 경로는 환경변수 `LOG_HOME` 기준(`$LOG_HOME/{name}.log`), 미설정 시 `/tmp/Log/{name}.log`. 초기화 시 `os.Args` 전체를 INFO로 로깅한다.
- 사용 예 (`goPlus/Matching/LocalLog/MatMain.go`):
  ```go
  Log.InitLogger(os.Args[1], Log.FILE_MODE)   // 또는 Log.CONSOLE_MODE / Log.ALL_MODE
  ```

### 레벨별 로그 함수 (모두 `Printf` 형식)

- `func Debug(format string, a ...interface{})`
- `func Infor(format string, a ...interface{})`
- `func Warn(format string, a ...interface{})`
- `func Error(format string, a ...interface{})`

`InitLogger` 호출 전에는 내부 `log`가 `nil`이라 아무 동작도 하지 않는다(패닉하지 않음).

### `func Close()`
- 설명: `zap.Logger.Sync()` — 버퍼링된 로그를 flush. 프로그램 종료 전 호출 권장.

### 로그 레벨 동적 변경

- `func SetFileLogLevel(loglv zapcore.Level)` — 파일 로그 레벨만 변경(변경 중 콘솔 레벨은 일시적으로 Debug로 내렸다가 원복).
- `func SetConsoleLogLevel(loglv zapcore.Level)` — 콘솔 로그 레벨만 변경(동일 방식).
- `func GetFileLogLevel() string` / `func GetConsoleLogLevel() string` — 현재 레벨 문자열 조회(`InitLogger` 미호출 시 `"NULL"`).

레벨 상수: `DEBUG`, `INFOR`, `WARN`, `ERROR` (= `zapcore.DebugLevel` 등의 별칭).
모드 상수: `CONSOLE_MODE = 0x01`, `FILE_MODE = 0x10`, `ALL_MODE = CONSOLE_MODE|FILE_MODE`.

## 의존성

`go.uber.org/zap`, `gopkg.in/natefinch/lumberjack.v2`.

## 주의사항

- 프로세스 전역 싱글턴(패키지 레벨 `log` 변수)이다 — 프로세스당 1회만 `InitLogger`를 호출해야 하며, 동시성 안전성(mutex)은 zap 내부 구현에 맡긴다.
- 기본 로그 레벨은 Debug이다. 운영 환경에서는 `SetFileLogLevel`/`SetConsoleLogLevel`로 낮춰야 로그량을 줄일 수 있다.
- `LOG_HOME` 미설정 시 `/tmp/Log`에 쓰므로, 운영 배포 시 반드시 `LOG_HOME`을 설정할 것(디스크 정리 정책 대상에서 벗어날 수 있음).
