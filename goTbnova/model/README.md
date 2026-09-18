# goTbnova/model

goTbnova 전역 데이터 구조체, enum 타입, 상태 머신 정의. 순수 타입 정의 패키지(로직 없음).

## Import

```go
import "goTbnova/model"
```

## 개요

TOML 파싱 구조체(`XXXConfig`, 불변)와 런타임 구조체(`PInfo`/`GInfo`, 상태 변경됨)를 명확히 분리한다. C버전의 `stcnt` 음수값 이중 역할(종료 원인 + 카운터)을 `StopReason` 타입으로, `pid` 값(-1/0/>0) 조합으로 판단하던 상태를 `ProcessState` enum으로 명시화했다. 모든 `PInfo` 필드는 [process.Manager.Run()](../process/README.md) goroutine에서만 접근한다는 것이 설계 전제다.

## API

### 열거형

- `ProcessMode` (`toml:"mode"`): `ModePermanent("P")` 상시운영 / `ModeRestricted("R")` 시간제한(`st`~`et`) / `ModeOnce("O")` 1회 실행.
- `StartMode` (`toml:"stmode"`): `StartQuick("Q")` 즉시 재기동 / `StartSlow("S")` 2초 지연 후 재기동.
- `RestartPolicy` (`toml:"policy"`): `PolicyAlways` / `PolicyOnFailure` / `PolicyNever`.
- `CascadeAction` (`toml:"on_death"`): `CascadeNone` / `CascadeNotify`(로그만) / `CascadeStopDependents`(종속 프로세스 연쇄 종료).
- `StopReason`(int): `StopNone(0)`, `StopByOperator(-1)`, `StopByHoliday(-2)`, `StopByRestart(-3)`, `StopByTimeUp(-4)`, `StopByReload(-5)`, `StopByAbnormal(-6)`, `StopByMaxCount(-7)`. C버전 `stcnt` 음수값과 1:1 대응(단 `-6`/`-7`은 Go 버전 신규 추가).
- `ProcessState`(int, iota): `StateWaitTime`(기동시각 대기) → `StateWaitDepend`(종속 대기) → `StateRunning`(실행중) → `StateStopping`(SIGTERM 발송, grace 대기) / `StateBackoff`(비정상 종료 후 대기) → `StateDisabled`(비활성화, 운영자 개입 필요). 전이 규칙 전체는 소스 상단 주석의 다이어그램 참고.

### TOML 파싱 구조체 (`config.Load()`가 채움, 불변)

- `RestartConfig{Policy, Backoff bool, MaxCount int, GracePeriod int}` — `[process.restart]`.
- `CascadeConfig{OnDeath}` — `[process.cascade]`.
- `ProcessConfig{ID, Group, Desc, Mode, Weekend, StMode, St, Et, Path, Args []string, Env map[string]string, DependsOn []string, DependsHealthy bool, Restart RestartConfig, Cascade CascadeConfig}` — `[[process]]` 섹션 1건.
- `GroupConfig{Desc}` — `[groups.XXX]`.
- `DefaultConfig{DefDir, DefCnt int}` — `[default]`.
- `TableConfig{Version, Env, Default, Groups map[string]GroupConfig, Process []ProcessConfig}` — `process.toml` 전체.

### 런타임 구조체 (`process` 패키지가 상태를 변경, `Run()` goroutine 전용)

- `BackoffState{FailCount int, NextRetry time.Time}` — Back-off 대기 시퀀스(1→2→4→8→16→32→300초) 상태.
- `PInfo{Config ProcessConfig, Pid, State, StopReason, StartCount int, Start/End/LastStart time.Time, Dependents []*PInfo, Cmd *exec.Cmd, Backoff BackoffState}` — 프로세스 1개의 전체 런타임 상태. `StartCount`는 초기값 1이며 **비정상 종료 시에만 증가**한다(정상 종료·재기동 요청으로는 증가하지 않음).
- `GInfo{Name, Desc, Members []*PInfo}` — 그룹(`[groups.XXX]`) 1개.

### 채널/프로토콜 구조체

- `ExitEvent{P *PInfo, Err error}` — [process/runner.go](../process/README.md)의 `cmd.Wait()` goroutine이 manager의 `exitCh`로 전달.
- `ControlCmd{Cmd, ID string}` — TCP 모니터로부터 수신하는 JSON 명령(`status/start/stop/restart/reload/env`).
- `ProcessSnapshot` — `status` 응답용 읽기 전용 스냅샷. `PInfo`를 직접 노출하지 않고 복사본으로 전달해 동시성 문제를 방지한다. `Env` 필드는 전역+프로세스별 env를 병합한 최종 적용값.
- `ControlResponse{Result("ok"|"error"), Message, Processes []ProcessSnapshot, Env map[string]string}` — 모니터 JSON 응답.

## 의존성

`os/exec`, `time` (표준 라이브러리만).

## 주의사항

- 이 패키지는 상태를 스스로 변경하지 않는다 — 모든 필드 변경은 [process](../process/README.md) 패키지(주로 `Manager.Run()` 단일 goroutine)의 책임이다. `model` 밖에서 `PInfo`를 직접 조작하면 동시성 안전성이 깨진다.
- `StartCount`는 "누적 기동 횟수"가 아니라 "**비정상 종료 횟수 + 1**"이다 — 로그/모니터링 해석 시 혼동 주의.
