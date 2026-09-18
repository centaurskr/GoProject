# goClsTbnova

`goTbnova` 종료 후 독립 실행 중인(고아가 된) 관리 대상 프로세스를 탐지·종료하는 정리 유틸리티. 더 상세한 설명/예시는 [goClsTbnova.md](goClsTbnova.md)에 있으며, 이 문서는 표준 요약이다.

## 개요

- `module goClsTbnova`, `go 1.24.5` — 독립적인 `go.mod`를 가진 단일 파일(`goclstbnova.go`) 실행 프로그램이다.
- `goTbnova`는 자식 프로세스를 `Setsid=true`로 독립 세션 기동한다. `goTbnova`가 종료되면 자식들은 `ppid=1(init)` 하위로 재배치되어 계속 실행된다 — 이 유틸리티는 그 프로세스들을 `process.toml` 기준으로 탐지해 정리한다.
- 레거시 C버전 `clschaos`를 대체한다(TOML 설정, `/proc` 기반 탐지, dry-run 지원 등이 신규/개선점).

## 빌드

```bash
cd goClsTbnova
go build -o goClsTbnova
# 또는
./build.sh   # 빌드 + strip + ~/bin 으로 이동
```

## 사용법

```
goClsTbnova [options] <table_path>
```

`table_path`는 `process.toml` 경로(필수).

| 옵션 | 기본값 | 설명 |
|---|---|---|
| `-u <user>` | 현재 유저 | 대상 유저명 |
| `-n` | false | dry-run: 목록만 출력, 종료 안함 |
| `-f` | false | 확인 없이 즉시 종료 |
| `-s` | false | SIGKILL 대신 SIGTERM 발송 |
| `-v` | false | 상세 출력 (cmdline 전체 표시) |

```bash
goClsTbnova -n /home/cento/conf/process.toml           # dry-run으로 먼저 확인
goClsTbnova    /home/cento/conf/process.toml           # 목록 확인 후 y/N 프롬프트
goClsTbnova -f /home/cento/conf/process.toml           # 확인 없이 SIGKILL
goClsTbnova -u cento -s /home/cento/conf/process.toml  # 특정 유저, SIGTERM
```

종료 코드: `0` 정상(탐지 없음 포함), `1` 오류(인자/파일읽기/시그널 발송 실패).

## 탐지 기준 (모두 AND)

1. `/proc/{pid}/status`의 `PPid == 1`
2. `/proc/{pid}/status`의 `Uid`(real uid) == 대상 유저 UID
3. `/proc/{pid}/exe` 절대경로가 `process.toml`의 `path`와 일치 (미일치 시 `basename` 보조 비교)

## 환경설정

별도 환경변수는 사용하지 않는다. `table_path` 인자로 넘기는 `process.toml`이 유일한 입력이며, `[default].def_dir` 기준으로 상대경로를 절대경로로 변환해 대상 실행파일 경로를 판단한다(goTbnova의 [config](../goTbnova/config/README.md)와 동일한 TOML을 최소 필드(`id`, `path`, `default.def_dir`)만 파싱).

## 주의사항

- **`goTbnova` 실행 중에는 사용하지 말 것.** goTbnova가 실행 중이면 자식의 `ppid`는 goTbnova의 PID이므로 원칙적으로 탐지 대상에서 제외되지만, goTbnova가 관리 중인 프로세스를 종료시키면 goTbnova가 즉시 재기동을 시도해 충돌할 수 있다. 반드시 goTbnova 종료 후 사용한다.
- 기본 시그널은 **SIGKILL**이다(`-s` 지정 시에만 SIGTERM으로 graceful 종료).
- [goClsTbnova.md](goClsTbnova.md)에는 "go.mod는 상위 goTbnova 프로젝트의 것을 공유한다"고 적혀 있으나, 실제로는 **`goClsTbnova`가 자체 `go.mod`(`module goClsTbnova`)를 갖는 독립 모듈**이다(문서와 실제 코드가 다른 부분이므로 빌드 시 혼동 주의).
- 동일 실행 파일을 여러 `[[process]]` 항목이 공유하는 경우(예: 인덱스가 붙은 동일 바이너리) ID 목록을 배열로 관리하여 출력 시 콤마로 함께 표시한다.
