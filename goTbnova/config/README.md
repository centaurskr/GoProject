# goTbnova/config

`process.toml`/`gotbnova.toml` 로드, 검증, 리로드 시 변경분(diff) 계산을 담당하는 패키지.

## Import

```go
import "goTbnova/config"
```

## 개요

파일 구성: `table.go`(`process.toml` 로드 및 런타임 구조체 변환), `validator.go`(검증 + `-t`/`-d` 옵션용 CLI 출력), `envfile.go`(`gotbnova.toml` 실행 환경 로드), `diff.go`(리로드 시 추가/삭제/변경 프로세스 판별).

## API

### 설정 로드 (table.go)

#### `func Load(path string) (processes []*model.PInfo, groups []*model.GInfo, globalEnv map[string]string, err error)`
- 설명: **최상위 진입점.** TOML 파싱 → `Validate()` → 전역 env `$VAR` 치환 → 기본값 보완 → `GInfo`/`PInfo` 생성 → `depends_on` 연결까지 전체 로드 파이프라인을 수행한다. `process.NewManager()`와 리로드 시 호출된다.
- 환경변수 치환 규칙: `$VAR` 형식만 지원(`${VAR}` 미지원), 탐색 순서는 전역 `[env]` → OS 환경변수 → 실패 시 빈 문자열(쉘과 동일). 치환은 `Load()` 시점에 1회만 수행.

#### `func BackoffDuration(failCount int) time.Duration`
- 설명: 실패 횟수(1부터)에 대응하는 Back-off 대기시간을 `backoffTable = [1,2,4,8,16,32,300]`(초)에서 조회. `failCount`가 테이블을 초과하면 300초 상한 유지.

### 기본값

- `defaultGracePeriod = 10`(초), `defaultMaxCount = 5`, `defaultInterval = 10`(현재 미사용, 확장 예비).
- `applyDefaults`가 `[default].def_cnt` 미설정 시 5, 프로세스별 `restart.policy` 미설정 시 `on-failure`, `restart.grace_period` 미설정 시 10초, `cascade.on_death` 미설정 시 `none`, `stmode` 미설정 시 `Q`로 채운다.

### 검증 (validator.go)

#### `func Validate(cfg *model.TableConfig) error`
- 설명: `Load()` 내부에서 호출되는 검증 진입점. `[process]` 섹션 비어있음/ID 중복/개별 필드 유효성(`validateProcess`)/`depends_on` 순환 참조(`checkCircularDepends`, DFS)를 첫 오류에서 즉시 중단하며 검사.

#### `func TestConfig(path string)`
- 설명: `goTbnova -t <path>` 구현체. TOML 문법 → 프로세스별 전체 검증(`validateProcessFull`, 오류를 모두 수집)까지 수행하고 라인번호(`buildLineIndex`로 사전 스캔)와 함께 사람이 읽는 리포트를 표준출력/stderr에 출력한다. **내부에서 `os.Exit(0/1)` 호출** — 성공 0, 실패 시 오류 개수와 함께 1.

#### `func DumpConfig(path string)`
- 설명: `goTbnova -d <path>` 구현체. `$VAR` 치환 및 기본값 적용까지 반영한 최종 설정값을 프로세스별로 사람이 읽는 형태로 표준출력에 덤프한다. **내부에서 `os.Exit(0/1)` 호출.**

검증 항목(요약): 필수 필드(ID/path/mode), ID 13자 이내 및 중복 금지, `mode∈{P,R,O}`, R/O 모드는 `st/et`가 `HH:MM:SS` 형식, `stmode∈{Q,S}`, `restart.policy∈{always,on-failure,never}`, `grace_period`/`max_count` ≥ 0, `cascade.on_death∈{none,notify,stop_dependents}`, 실행 파일 존재 및 실행 권한(`TestConfig`만 상세 확인, `Validate`는 `os.Stat` 존재 여부만), `depends_on` 순환 참조 없음.

### 환경 파일 로드 (envfile.go)

#### `func LoadEnvFile() (EnvFile, string, error)`
- 설명: `$ENV_HOME/gotbnova.toml` → 없으면 실행 파일 디렉토리의 `gotbnova.toml`을 탐색해 로드. **파일이 없어도 에러가 아니며** 기본값(`EnvFile{Server.Bind:"127.0.0.1", Log.Type:"net", Log.VectorSocket:"/tmp/VECTOR.UNIX.DGRAM"}`)으로 채워 반환한다. 반환값의 두 번째(`string`)는 실제 로드된 경로(없으면 `""`).
- `EnvFile{Server EnvServer{Bind}, Log EnvLog{Type, VectorSocket, Dir}, Work EnvWork{Dir}}` — `gotbnova.toml`의 `[server]/[log]/[work]` 섹션.

#### `func DumpEnvFile(ef EnvFile, path string)`
- 설명: `goTbnova -te` 구현체가 사용하는 출력 헬퍼. `EnvFile` 내용을 사람이 읽는 형태로 표준출력에 덤프(종료 처리는 호출자인 `main.testEnvFile()`이 담당).

### 리로드 diff (diff.go)

#### `func Diff(current []*model.PInfo, next *model.TableConfig) DiffResult`
- 설명: ID 기준으로 현재 실행 중 설정과 새로 로드한 설정을 비교해 `Added/Removed/Modified/Unchanged`로 분류한다.
- `DiffResult{Added, Removed, Modified, Unchanged []model.ProcessConfig}`.
- 변경 판정(`isModified`) 대상 필드: `path, args, mode, st, et, weekend, restart 정책 전체, cascade.on_death, depends_on`. **제외**: `desc`(운영 무관), `env`(다음 기동 시 적용되므로 diff 트리거 아님).
- 참고: 이 `Diff` 함수는 현재 [process/manager.go](../process/README.md)의 `handleReload`가 직접 자체 로직으로 재구현하여 사용 중이며(맵 기반 비교), 이 함수 자체는 동일한 목적의 별도 유틸리티로 존재한다.

## 의존성

`github.com/BurntSushi/toml`, `goTbnova/model`.

## 주의사항

- `TestConfig`/`DumpConfig`/`DumpEnvFile`을 호출하는 `testEnvFile()`은 **내부에서 `os.Exit()`을 호출**한다 — 일반적인 에러 반환 함수가 아니므로 `defer`로 정리해야 할 리소스가 있는 컨텍스트에서 호출하면 안 된다(현재 `main()`에서 다른 초기화 이전에만 호출됨).
- `$VAR` 치환은 `${VAR}` 형식을 지원하지 않는다 — `process.toml`/`gotbnova.toml`에 `${VAR}`를 쓰면 리터럴 문자열로 남는다.
- `Load()`의 환경변수 순환 참조는 탐지하지 않는다(`expandGlobalEnv`) — 전역 env끼리 서로를 순환 참조하면 무한 확장 대신 그냥 한 번만 치환되고 끝난다(원본 맵을 먼저 복사해두고 그 위에서 치환하므로).
