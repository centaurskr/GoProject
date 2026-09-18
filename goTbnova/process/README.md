# goTbnova/process

프로세스 생명주기 관리 핵심 패키지 — 기동/종료/재기동/스케줄링/Back-off/제어명령 처리를 모두 담당한다.

## Import

```go
import "goTbnova/process"
```

## 개요

파일 구성: `manager.go`(`Manager` — 모든 상태 변경의 단일 진입점), `runner.go`(`exec.Cmd` 기동/종료 래퍼), `scheduler.go`(2초 tick마다 시각/요일 기반 Action 판단, 순수 함수), `backoff.go`(재기동 대기시간 계산, 종속관계 충족 확인).

### 설계 원칙

모든 `model.PInfo` 상태 변경은 `Manager.Run()` **단일 goroutine**에서만 일어난다 — 별도 mutex 없이 동시성 안전성을 확보하는 핵심 전제다. 외부 제어 요청(TCP)은 `controlCh` 채널로, 프로세스 종료 감지는 `runner.go`의 `cmd.Wait()` goroutine이 `exitCh` 채널로 전달한다. `/proc` 폴링은 전혀 사용하지 않는다.

```
                    ┌──────────────────────────────┐
control/Server ───▶ │ controlCh  chan ControlRequest│
                    │                                │
runner (cmd.Wait) ─▶│ exitCh     chan model.ExitEvent│──▶ Manager.Run() (단일 goroutine)
                    │ reloadCh   chan struct{}      │      └─ scheduler.Tick() 2초마다
main (signal) ─────▶│ ctx.Done()                     │
                    └──────────────────────────────┘
```

## API

### `Manager` (manager.go)

#### `func NewManager(tablePath string) (*Manager, error)`
- 설명: `config.Load(tablePath)`로 초기 설정을 로드하고 채널(`exitCh` 버퍼 256, `controlCh` 버퍼 32, `reloadCh` 버퍼 1)을 생성한다. 모든 프로세스의 초기 `StartCount`를 1로 설정.

#### `func (m *Manager) Run(ctx context.Context)`
- 설명: **메인 이벤트 루프.** `select`로 `exitCh`(종료 이벤트) / 2초 `ticker`(스케줄링) / `controlCh`(제어명령) / `reloadCh`(리로드) / `ctx.Done()`(Graceful Shutdown)을 처리한다. `ctx` 취소 시 `handleShutdown()` 실행 후 반환. `main()`에서 `go mgr.Run(ctx)`로 기동.

#### `func (m *Manager) ControlCh() chan<- ControlRequest` / `func (m *Manager) ReloadCh() chan<- struct{}`
- 설명: 외부(`control.Server`)가 요청을 보낼 채널의 송신 전용 뷰를 반환.

#### `ControlRequest{Cmd, ID string, RespCh chan model.ControlResponse}`
- 설명: `control/server.go`가 manager로 보내는 요청 단위. 응답은 요청자가 만든 `RespCh`로 정확히 1회 전달된다.

### 리로드 규칙 (`handleReload`, 상세 규칙은 소스 주석)

1. 실행 중 프로세스: ENV만 갱신, 재기동 없음(다른 필드 변경은 다음 자연 재기동 시 적용).
2. 실행 중이나 새 테이블에서 삭제된 프로세스: 종료(`StopByReload`) 후 목록에서 제거.
3. 신규 추가: 목록에 추가, 다음 tick에서 스케줄러가 기동 판단.
4. 그 외(비실행 중) 변경분: 새 설정 전체 적용, 단 `StartCount`/`Backoff`/`State`(Disabled면 유지)는 보존.

### 기동/종료 (`Start`/`Stop`, runner.go)

#### `func Start(p *model.PInfo, globalEnv map[string]string, exitCh chan<- model.ExitEvent) error`
- 설명: 실행 파일 존재 확인 → env 병합(`mergeEnv`, 프로세스별이 전역을 덮어씀) → `exec.Cmd` 구성(stdin/stdout/stderr 전부 `/dev/null`, `SysProcAttr{Setsid: true}`로 독립 세션) → `cmd.Start()` → `PInfo` 상태 갱신(`Pid, State=StateRunning, LastStart`) → 종료 감지용 `cmd.Wait()` goroutine 기동(완료 시 `exitCh`로 `ExitEvent` 전달).
- 주의: **`StartCount`는 여기서 증가하지 않는다** — 비정상 종료 시에만 `manager.handleExit`이 증가시킨다.

#### `func Stop(p *model.PInfo, reason model.StopReason, exitCh <-chan model.ExitEvent)`
- 설명: SIGTERM 발송 후 `grace_period`(기본 10초) 동안 `exitCh`에서 이 프로세스의 종료 이벤트를 기다리고, 타임아웃 시 SIGKILL. **주의**: 이 함수는 `manager.go`의 `stopProcess()`(별도 경량 구현, SIGTERM+goroutine 타이머 방식)와는 별개의 구현이다 — 현재 `Manager`가 실제로 쓰는 것은 `manager.stopProcess()`이며, `runner.Stop()`은 `exitCh`를 직접 컨슘하려 하므로 `Manager.Run()`의 단일 소비자 원칙과 맞지 않아 **manager 루프 밖에서 직접 호출하면 안 된다**(소스 주석에도 명시).

### 스케줄러 (scheduler.go)

#### `func (s *Scheduler) Tick(procs []*model.PInfo, now time.Time) []Action`
- 설명: 2초마다 전체 프로세스를 순회해 상태별로 취해야 할 `Action`(`ActionStart/ActionStop/ActionStopHoliday/ActionDateChange`)을 판단하는 **순수 함수** — 상태를 직접 바꾸지 않고 `Action` 목록만 반환, 실제 처리는 `manager.handleTick`이 담당(단, 종속 대기 전환(`StateWaitDepend`)처럼 일부 상태 전이는 판단 과정에서 바로 반영됨).
- 상태별 판단: `StateRunning`(시간 만료/주말 정책 → 종료), `StateWaitTime`(시각 도달+종속조건 → 기동), `StateWaitDepend`(종속조건 충족 확인), `StateBackoff`(대기 완료 확인).

#### `func AdvanceDay(procs []*model.PInfo, now time.Time, isWeekend bool)`
- 설명: 일자 변경 시(`ActionDateChange`) 모든 비-Permanent 프로세스의 `Start`/`End` 시각을 새 날짜 기준으로 재계산.

#### `func RescheduleModeOnce(p *model.PInfo)`
- 설명: `ModeOnce` 프로세스가 당일 실행을 정상 완료하면 `StateWaitTime`으로 전환해 다음 날 재기동을 준비.

### Back-off (backoff.go)

#### `func CalcBackoff(p *model.PInfo, now time.Time) model.ProcessState`
- 설명: 비정상 종료 시 호출. `FailCount`를 증가시키고, `restart.backoff=false`면 즉시 재기동(`NextRetry=now`), `true`면 `backoffTable=[1,2,4,8,16,32,300]`초(`StartMode=S`면 +2초) 기준 `NextRetry`를 설정. 항상 `StateBackoff`를 반환하며(MaxCount 판단은 호출자인 `handleExit`의 책임), `p.StartCount` 기준으로 로그를 남긴다.

#### `func IsReady(p *model.PInfo, now time.Time) bool` / `func Reset(p *model.PInfo)`
- 설명: Back-off 대기 완료 여부 확인 / 정상 기동 시 `FailCount`·`NextRetry` 초기화(`StartCount`는 유지).

#### `func DependsSatisfied(p *model.PInfo, allProcs []*model.PInfo, now time.Time) bool`
- 설명: `depends_on` 전체가 `StateRunning`인지 확인. `depends_healthy=true`이면 추가로 `LastStart`로부터 2초(1 tick) 이상 경과했는지도 확인.

## 의존성

`goTbnova/model`, `goTbnova/config`, `goTbnova/tlog`.

## 주의사항

- `runner.Stop()`과 `manager.stopProcess()`는 별개의 종료 구현이다 — **manager가 실제로 사용하는 것은 `stopProcess()`**이며, `runner.Stop()`을 직접 호출하면 `exitCh`를 두 goroutine이 동시에 소비하게 되어 이벤트 유실 위험이 있다.
- `StartCount`는 "비정상 종료 시에만 증가"(규칙 5) — 정상 종료·재기동 요청·리로드로 인한 재시작에서는 증가하지 않는다. `MaxCount` 판단은 `StartCount - 1`(실제 비정상 종료 횟수)을 기준으로 한다.
- `handleShutdown()`은 실행 중인 자식 프로세스에 **아무 신호도 보내지 않는다** — goTbnova 프로세스만 종료되고 자식은 그대로 살아있다(설계 의도, [goTbnova README](../README.md) 참고).
- `PolicyNever`는 1회 비정상 종료만으로 즉시 `StateDisabled`로 전환한다(재시도 없음).
