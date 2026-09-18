///
/// @file    process/manager.go
/// @brief   프로세스 생명주기 관리 메인 루프
/// @details goTbnova 의 핵심 컴포넌트로, 모든 프로세스 상태 변경을
///          단일 goroutine(Run) 에서 처리한다.
///
///          설계 원칙:
///            1. 모든 PInfo 상태 변경은 Run() goroutine 에서만 수행한다.
///               별도 mutex 없이 동시성 안전을 보장한다.
///            2. 외부(control/server.go) 의 제어 요청은 controlCh 채널로 수신한다.
///               응답은 ControlRequest.respCh 채널로 전달한다.
///            3. 프로세스 종료 이벤트는 runner 의 cmd.Wait() goroutine 이
///               exitCh 채널로 전달한다.
///            4. 테이블 리로드는 config.Diff() 결과를 기반으로
///               변경된 프로세스만 처리한다.
///            5. Graceful Shutdown 은 실행 중인 모든 프로세스에
///               SIGTERM → grace_period → SIGKILL 순서로 처리한다.
///
///          채널 구조:
///            exitCh    chan ExitEvent      runner.Wait() → manager
///            controlCh chan ControlRequest server       → manager
///            reloadCh  chan struct{}        server       → manager (reload 신호)
///            ctx.Done()                    main         → manager (shutdown)
///
///          C버전 대응표:
///            Event_StartLoop()      →  Run() goroutine + select
///            CheckProcessHandler()  →  ticker + Scheduler.Tick()
///            CheckSignal()          →  exitCh 수신 처리
///            PublishProcessStatus() →  Snapshot() → ControlResponse
///
/// @date    2025-01-04
/// @author  Cento
///

package process

import (
	"context"
	"fmt"
	"strings"
	"syscall"
	"time"

	"goTbnova/config"
	"goTbnova/model"
	"goTbnova/tlog"
)

////////////////////////////////////////////////////////////////////////////////
// 타입 정의
////////////////////////////////////////////////////////////////////////////////

/// ControlRequest 는 control/server.go 에서 manager 로 전달되는 제어 요청이다.
/// 응답은 respCh 채널로 반환한다.
type ControlRequest struct {
	Cmd    string                    ///< 명령어 (status/start/stop/restart/reload)
	ID     string                    ///< 대상 프로세스 ID (빈 문자열: 전체)
	RespCh chan model.ControlResponse ///< 응답 채널 (server 에서 생성)
}

////////////////////////////////////////////////////////////////////////////////
// Manager 구조체
////////////////////////////////////////////////////////////////////////////////

/// Manager 는 프로세스 생명주기를 총괄한다.
/// Run() 이 유일한 상태 변경 진입점이다.
type Manager struct {
	/// 설정
	tablePath string ///< TOML 설정 파일 경로
	globalEnv map[string]string ///< 전역 환경변수 (config.Load() 반환)

	/// 런타임 상태 (Run() goroutine 에서만 접근)
	processes []*model.PInfo ///< 전체 프로세스 목록
	groups    []*model.GInfo ///< 전체 그룹 목록
	scheduler *Scheduler     ///< 스케줄러

	/// 채널
	exitCh    chan model.ExitEvent ///< 프로세스 종료 이벤트 (runner → manager)
	controlCh chan ControlRequest  ///< 제어 요청 (server → manager)
	reloadCh  chan struct{}         ///< 테이블 리로드 신호 (server → manager)
}

////////////////////////////////////////////////////////////////////////////////
// 공개 함수
////////////////////////////////////////////////////////////////////////////////

/// NewManager 는 Manager 를 생성하고 초기 설정 파일을 로드한다.
///
/// @param  tablePath  TOML 설정 파일 경로
/// @return            Manager 포인터
/// @return            로드 실패 시 에러
func NewManager(tablePath string) (*Manager, error) {
	procs, groups, globalEnv, err := config.Load(tablePath)
	if err != nil {
		return nil, fmt.Errorf("설정 로드 실패: %w", err)
	}

	m := &Manager{
		tablePath: tablePath,
		globalEnv: globalEnv,
		processes: procs,
		groups:    groups,
		scheduler: NewScheduler(),
		exitCh:    make(chan model.ExitEvent, 256),
		controlCh: make(chan ControlRequest, 32),
		reloadCh:  make(chan struct{}, 1),
	}

	/// 최초 로드 시 모든 프로세스의 StartCount 초기값을 1 로 설정
	for _, p := range procs {
		p.StartCount = 1
	}

	tlog.TableLoad(tablePath, len(procs), len(groups))
	return m, nil
}

/// Run 은 메인 이벤트 루프를 실행한다.
/// ctx 취소 시 Graceful Shutdown 을 수행하고 반환한다.
/// main() 에서 goroutine 으로 실행된다.
///
/// @param  ctx  상위 컨텍스트 (main 에서 signal 수신 시 취소)
func (m *Manager) Run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	tlog.Info("event=manager_start proc_count=%d", len(m.processes))

	for {
		select {

		/// ── 프로세스 종료 이벤트 ─────────────────────────────
		case ev := <-m.exitCh:
			m.handleExit(ev)

		/// ── 2초 tick ─────────────────────────────────────────
		case now := <-ticker.C:
			m.handleTick(now)

		/// ── 제어 명령 ─────────────────────────────────────────
		case req := <-m.controlCh:
			m.handleControl(req)

		/// ── 테이블 리로드 ─────────────────────────────────────
		case <-m.reloadCh:
			m.handleReload()

		/// ── Graceful Shutdown ─────────────────────────────────
		case <-ctx.Done():
			m.handleShutdown()
			return
		}
	}
}

/// ControlCh 는 control/server.go 가 제어 요청을 전달할 채널을 반환한다.
func (m *Manager) ControlCh() chan<- ControlRequest {
	return m.controlCh
}

/// ReloadCh 는 테이블 리로드 신호 채널을 반환한다.
func (m *Manager) ReloadCh() chan<- struct{} {
	return m.reloadCh
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수 — 이벤트 핸들러
////////////////////////////////////////////////////////////////////////////////

/// handleExit 는 프로세스 종료 이벤트를 처리한다.
/// runner.go 의 cmd.Wait() goroutine 이 exitCh 로 전달한 이벤트를 수신한다.
///
/// 처리 순서:
///   1. 리로드 중 제거된 프로세스 → 단순 정리 후 반환
///   2. 종료 원인 판단 (정상/비정상)
///   3. 비정상 종료 시 StartCount 증가 (규칙 5) + Back-off 계산
///   4. PolicyNever 이면 재기동하지 않음
///   5. Cascade 처리 (종속 프로세스 연쇄 종료)
///   6. ModeOnce 정상 종료 처리
///
/// 규칙 5: StartCount 는 비정상 종료 시에만 증가한다.
func (m *Manager) handleExit(ev model.ExitEvent) {
	p := ev.P

	/// 이미 정리된 프로세스 방어 (중복 이벤트)
	if p.Pid == 0 {
		return
	}

	/// 규칙 2/리로드: m.processes 에서 이미 제거된 프로세스
	/// 리로드 중 Removed 처리된 프로세스가 나중에 exitCh 로 도착한 경우
	if m.findByPtr(p) == nil {
		p.Pid        = 0
		p.Cmd        = nil
		p.State      = model.StateWaitTime
		p.StopReason = model.StopNone
		tlog.ProcessExit(p.Config.ID, 0, "reload_removed")
		return
	}

	isAbnormal := ev.Err != nil && p.StopReason == model.StopNone

	if isAbnormal {
		/// 비정상 종료
		/// 규칙 5: StartCount 는 비정상 종료 시에만 증가 (초기값=1)
		p.StartCount++
		tlog.ProcessAbnormal(p.Config.ID, p.Pid, p.StartCount-1, ev.Err.Error())

		p.Pid = 0
		p.Cmd = nil

		/// MaxCount 판단: StartCount - 1 = 실제 비정상 종료 횟수
		/// 예) max_count=5, 5회 비정상 종료 시 StartCount=6 → 6-1=5 >= 5 → Disabled
		maxCount := p.Config.Restart.MaxCount
		failures  := p.StartCount - 1  ///< 실제 비정상 종료 횟수 (StartCount 에서 초기값 1 제거)

		if p.Config.Restart.Policy == model.PolicyNever {
			/// PolicyNever: 1회 비정상 종료 즉시 비활성화
			p.State      = model.StateDisabled
			p.StopReason = model.StopByAbnormal
			tlog.ProcessDisabled(p.Config.ID, maxCount)
			return
		}

		if maxCount > 0 && failures >= maxCount {
			/// MaxCount 도달: StateDisabled 전환, 자동 재기동 중단
			p.State      = model.StateDisabled
			p.StopReason = model.StopByMaxCount
			tlog.ProcessDisabled(p.Config.ID, maxCount)
			/// Cascade 처리 (Disabled 이후에도 종속 프로세스 연쇄 처리)
			m.handleCascade(p)
			return
		}

		/// Back-off 대기 시간 계산 (MaxCount 판단은 위에서 완료)
		nextState    := CalcBackoff(p, time.Now())
		p.State       = nextState
		p.StopReason  = model.StopByAbnormal

		/// Cascade 처리
		m.handleCascade(p)

	} else {
		/// 정상 종료 (운영자/시간/리로드/재기동 요청 등)
		/// 규칙 5: 정상 종료 시 StartCount 증가 없음
		reason := p.StopReason
		tlog.ProcessExit(p.Config.ID, p.Pid, tlog.StopReasonStr(int(reason)))

		p.Pid = 0
		p.Cmd = nil

		switch reason {
		case model.StopByOperator:
			/// 운영자 수동 종료: StateDisabled → 수동 start 전까지 재기동 안함
			p.State = model.StateDisabled

		case model.StopByHoliday:
			p.State      = model.StateWaitTime
			p.StopReason = model.StopNone

		case model.StopByTimeUp:
			p.State      = model.StateWaitTime
			p.StopReason = model.StopNone

		case model.StopByReload:
			/// 리로드 중 ENV 외 변경으로 재기동 대상이 된 경우
			/// handleReload 에서 이미 새 Config 가 적용된 상태
			/// 설정 변경 후 count 를 1 로 재설정한다 (초기값 = 1)
			p.State      = model.StateWaitTime
			p.StopReason = model.StopNone
			p.StartCount = 1  ///< 설정 변경 후 카운트 초기값 1 로 리셋

		case model.StopByRestart:
			/// 운영자 restart 요청: 즉시 재기동 대기
			p.State      = model.StateWaitTime
			p.StopReason = model.StopNone

		default:
			p.State = model.StateWaitTime
		}

		/// ModeOnce 정상 완료 처리
		if p.Config.Mode == model.ModeOnce && reason == model.StopNone {
			RescheduleModeOnce(p)
		}
	}
}

/// handleTick 은 2초 ticker 이벤트를 처리한다.
/// Scheduler.Tick() 으로 Action 목록을 받아 순차 처리한다.
func (m *Manager) handleTick(now time.Time) {
	/// 로그 파일 일자 교체 확인 (fileBackend 시)
	tlog.RotateIfNeeded()

	actions := m.scheduler.Tick(m.processes, now)

	for _, action := range actions {
		switch action.Type {

		case ActionDateChange:
			m.handleDateChange(now)

		case ActionStart:
			m.startProcess(action.P, now)

		case ActionStop:
			action.P.StopReason = model.StopByTimeUp
			m.stopProcess(action.P)

		case ActionStopHoliday:
			action.P.StopReason = model.StopByHoliday
			m.stopProcess(action.P)
		}
	}
}

/// handleDateChange 는 일자 변경 시 필요한 처리를 수행한다.
///
/// 처리 순서:
///   1. tlog 일자 변경 이벤트 기록
///   2. AdvanceDay() 로 모든 프로세스의 Start/End 시각 갱신
///   3. ModeOnce StartCount 초기화
func (m *Manager) handleDateChange(now time.Time) {
	isWeekend := isWeekendDay(now)
	date       := now.Format("2006-01-02")

	tlog.DateChanged(int(now.Weekday()), date)
	tlog.RotateIfNeeded()

	AdvanceDay(m.processes, now, isWeekend)

	/// ModeOnce: 당일 기동 횟수 초기화 (다음 날 기동 준비)
	for _, p := range m.processes {
		if p.Config.Mode == model.ModeOnce {
			p.StartCount = 0
		}
	}
}

/// handleControl 은 TCP 모니터에서 수신한 제어 요청을 처리한다.
func (m *Manager) handleControl(req ControlRequest) {
	var resp model.ControlResponse

	switch req.Cmd {
	case "status":
		resp = m.cmdStatus(req.ID)
	case "start":
		resp = m.cmdStart(req.ID)
	case "stop":
		resp = m.cmdStop(req.ID)
	case "restart":
		resp = m.cmdRestart(req.ID)
	case "reload":
		/// 비동기 처리: reloadCh 에 신호 전달 후 즉시 응답
		select {
		case m.reloadCh <- struct{}{}:
			resp = okResp("리로드 요청 접수됨")
		default:
			resp = errResp("이미 리로드 진행 중입니다")
		}
	case "env":
		resp = m.cmdEnv(req.ID)
	default:
		resp = errResp(fmt.Sprintf("알 수 없는 명령어: %s", req.Cmd))
	}

	tlog.MonitorCmd(req.Cmd, idOrAll(req.ID), resp.Result)
	req.RespCh <- resp
}

/// handleReload 는 테이블 리로드를 처리한다.
///
/// 리로드 규칙:
///   규칙 1. 현재 실행 중인 프로세스 → local ENV 변경만 반영. 재기동 없음.
///           (path/mode/st/et 등 다른 변경은 다음 자연 재기동 시 적용)
///   규칙 2. 현재 실행 중이나 신규 테이블에서 삭제된 프로세스 → 종료 후 목록에서 제거.
///   규칙 3. 신규 추가된 프로세스 → 해당 위치에 추가. 스케줄러가 다음 tick 에서 기동.
///   공통.  변경/삭제 외 비실행 중 프로세스 → 새 설정 전체 적용.
///          리로드 후 Dependents 포인터 재구성.
func (m *Manager) handleReload() {

	/// Step 1: 새 설정 로드 및 검증 (실패 시 기존 설정 유지)
	newProcs, newGroups, newEnv, err := config.Load(m.tablePath)
	if err != nil {
		tlog.TableReloadFail(m.tablePath, err)
		return
	}

	/// 새 설정 ID 인덱스 구성
	newIdx := make(map[string]*model.PInfo, len(newProcs))
	for _, p := range newProcs {
		newIdx[p.Config.ID] = p
	}

	added, removed, envOnly, restarted := 0, 0, 0, 0

	/// Step 2: 현재 프로세스 순회 → 규칙별 처리
	for _, curr := range m.processes {
		newP, existsInNew := newIdx[curr.Config.ID]

		if !existsInNew {
			/// 규칙 2: 신규 테이블에서 삭제 → 실행 중이면 종료
			if curr.State == model.StateRunning && curr.Pid > 0 {
				curr.StopReason = model.StopByReload
				m.stopProcess(curr)
				tlog.ProcessExit(curr.Config.ID, curr.Pid, "reload_removed")
			}
			/// 비실행 중이면 목록에서 자연스럽게 제거됨 (아래 result 에서 제외)
			removed++
			continue
		}

		if curr.State == model.StateRunning {
			/// 규칙 1: 실행 중 프로세스 → ENV 만 업데이트, 재기동 없음
			curr.Config.Env = newP.Config.Env
			/// 새 목록에서도 현재 PInfo 를 그대로 사용
			newIdx[curr.Config.ID] = curr
			envOnly++
		} else {
			/// 비실행 중 프로세스 → 새 설정 전체 적용
			/// 단, 런타임 상태(StartCount, Backoff)는 보존
			newP.StartCount  = curr.StartCount
			newP.Backoff     = curr.Backoff
			newP.State       = curr.State
			newP.StopReason  = curr.StopReason
			if curr.State == model.StateDisabled {
				/// Disabled 상태는 유지 (운영자가 해제해야 함)
				newP.State = model.StateDisabled
			} else {
				newP.State = model.StateWaitTime
			}
			restarted++
		}
	}

	/// Step 3: 신규 추가 프로세스 카운트
	for _, np := range newProcs {
		if _, exists := findByIDInSlice(m.processes, np.Config.ID); !exists {
			added++
		}
	}

	/// Step 4: 새 프로세스 목록 구성 (newProcs 순서 유지)
	result := make([]*model.PInfo, 0, len(newProcs))
	for _, np := range newProcs {
		if p, ok := newIdx[np.Config.ID]; ok {
			/// 신규 추가 프로세스의 초기 StartCount 는 1
			if p.StartCount == 0 {
				p.StartCount = 1
			}
			result = append(result, p)
		}
	}

	/// Step 5: Dependents 포인터 재구성
	/// mergeReload 후 Unchanged 프로세스의 Dependents 가 구 PInfo 를 참조할 수 있으므로
	/// 전체 목록 기준으로 재연결한다.
	relinkDependents(result)

	m.processes = result
	m.groups    = newGroups
	m.globalEnv = newEnv

	tlog.TableReload(m.tablePath, added, removed, restarted)
	tlog.Info("event=reload_detail added=%d removed=%d env_only=%d config_changed=%d",
		added, removed, envOnly, restarted)
}

/// relinkDependents 는 전체 프로세스 목록 기준으로 Dependents 포인터를 재구성한다.
/// handleReload 후 반드시 호출해야 한다.
/// A.depends_on = ["B"] 이면 B.Dependents 에 A 를 추가한다.
///
/// @param  procs  재구성할 전체 프로세스 목록
func relinkDependents(procs []*model.PInfo) {
	/// 모든 Dependents 초기화
	for _, p := range procs {
		if cap(p.Dependents) > 0 {
			p.Dependents = p.Dependents[:0]
		}
	}

	/// ID → PInfo 인덱스
	idx := make(map[string]*model.PInfo, len(procs))
	for _, p := range procs {
		idx[p.Config.ID] = p
	}

	/// depends_on 기반으로 Dependents 재연결
	for _, p := range procs {
		for _, depID := range p.Config.DependsOn {
			if dep, ok := idx[depID]; ok {
				dep.Dependents = append(dep.Dependents, p)
			}
		}
	}
}

/// findByIDInSlice 는 슬라이스에서 ID 로 PInfo 를 찾는다.
func findByIDInSlice(procs []*model.PInfo, id string) (*model.PInfo, bool) {
	for _, p := range procs {
		if p.Config.ID == id {
			return p, true
		}
	}
	return nil, false
}

/// handleShutdown 은 goTbnova 종료 처리를 수행한다.
///
/// 설계 원칙:
///   goTbnova 가 종료되어도 자식 프로세스는 계속 실행된다.
///   (Setsid=true 로 자식이 독립 세션 → init(1) 하위로 재배치)
///   자식에게 SIGTERM 을 보내지 않는다.
///
/// 종료 후 자식 정리가 필요하면:
///   goTbmon 의 stop 명령 또는 clsGoTbnova 유틸리티를 사용한다.
func (m *Manager) handleShutdown() {
	/// 실행 중인 프로세스 수 로그 (자식은 종료하지 않음)
	runningCount := 0
	for _, p := range m.processes {
		if p.State == model.StateRunning && p.Pid > 0 {
			runningCount++
			tlog.Info("event=shutdown_leave_running id=%s pid=%d", p.Config.ID, p.Pid)
		}
	}

	tlog.Shutdown(fmt.Sprintf("SIGTERM/SIGINT leave_running=%d", runningCount))
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수 — 제어 명령 처리
////////////////////////////////////////////////////////////////////////////////

/// cmdStatus 는 프로세스 상태 조회 명령을 처리한다.
/// process.toml 의 모든 항목과 런타임 상태를 포함하여 반환한다.
func (m *Manager) cmdStatus(id string) model.ControlResponse {
	var snapshots []model.ProcessSnapshot

	targets := m.filterTargets(id)
	for _, p := range targets {
		snap := model.ProcessSnapshot{
			/// ── TOML 설정 항목 전체 ──────────────────────────────
			ID:             p.Config.ID,
			Group:          p.Config.Group,
			Desc:           p.Config.Desc,
			Mode:           p.Config.Mode,
			Weekend:        p.Config.Weekend,
			StMode:         p.Config.StMode,
			St:             p.Config.St,
			Et:             p.Config.Et,
			Path:           p.Config.Path,
			Args:           p.Config.Args,
			DependsOn:      p.Config.DependsOn,
			DependsHealthy: p.Config.DependsHealthy,
			Restart:        p.Config.Restart,
			Cascade:        p.Config.Cascade,
			Env:            mergeEnvMap(m.globalEnv, p.Config.Env),

			/// ── 런타임 상태 ───────────────────────────────────────
			State:      p.State,
			Pid:        p.Pid,
			StartCount: p.StartCount,
			StopReason: p.StopReason,
			LastStart:  p.LastStart,
		}
		SnapshotBackoff(p, &snap)
		snapshots = append(snapshots, snap)
	}

	return model.ControlResponse{
		Result:    "ok",
		Processes: snapshots,
	}
}

/// cmdEnv 는 환경변수 조회 명령을 처리한다.
///
/// ID 생략 시: 전역 env 반환
/// ID 지정 시: 전역 env + 프로세스별 env 병합값 반환
///             (프로세스별 env 가 전역 env 를 덮어씀)
///
/// @param  id  조회할 프로세스 ID (빈 문자열: 전역 env)
/// @return     ControlResponse (Env 필드에 환경변수 맵)
func (m *Manager) cmdEnv(id string) model.ControlResponse {
	/// 전역 env 조회
	if strings.TrimSpace(id) == "" {
		env := make(map[string]string, len(m.globalEnv))
		for k, v := range m.globalEnv {
			env[k] = v
		}
		return model.ControlResponse{
			Result:  "ok",
			Message: fmt.Sprintf("전역 환경변수 %d개", len(env)),
			Env:     env,
		}
	}

	/// 특정 프로세스 병합 env 조회
	p := m.findByID(id)
	if p == nil {
		return errResp(fmt.Sprintf("프로세스를 찾을 수 없습니다: %s", id))
	}

	env := mergeEnvMap(m.globalEnv, p.Config.Env)
	return model.ControlResponse{
		Result:  "ok",
		Message: fmt.Sprintf("[%s] 병합 환경변수 %d개 (전역 %d개 + 프로세스별 %d개)",
			id, len(env), len(m.globalEnv), len(p.Config.Env)),
		Env: env,
	}
}

/// cmdStart 는 프로세스 기동 명령을 처리한다.
func (m *Manager) cmdStart(id string) model.ControlResponse {
	p := m.findByID(id)
	if p == nil {
		return errResp(fmt.Sprintf("프로세스를 찾을 수 없습니다: %s", id))
	}
	if p.State == model.StateRunning {
		return errResp(fmt.Sprintf("[%s] 이미 실행 중입니다 (PID: %d)", id, p.Pid))
	}

	/// StateDisabled 해제 후 기동
	/// 규칙 4: client start 명령 시 수행 count 를 1 로 재설정한다.
	///         이후 비정상 종료 시 count 가 2, 3, ... 으로 증가한다.
	p.State      = model.StateWaitTime
	p.StopReason = model.StopNone
	p.StartCount = 1
	Reset(p)

	m.startProcess(p, time.Now())
	return okResp(fmt.Sprintf("[%s] 기동 요청됨", id))
}

/// cmdStop 은 프로세스 종료 명령을 처리한다.
func (m *Manager) cmdStop(id string) model.ControlResponse {
	p := m.findByID(id)
	if p == nil {
		return errResp(fmt.Sprintf("프로세스를 찾을 수 없습니다: %s", id))
	}
	if p.State != model.StateRunning {
		return errResp(fmt.Sprintf("[%s] 실행 중이 아닙니다", id))
	}

	p.StopReason = model.StopByOperator
	m.stopProcess(p)
	return okResp(fmt.Sprintf("[%s] 종료 요청됨", id))
}

/// cmdRestart 는 프로세스 재기동 명령을 처리한다.
func (m *Manager) cmdRestart(id string) model.ControlResponse {
	p := m.findByID(id)
	if p == nil {
		return errResp(fmt.Sprintf("프로세스를 찾을 수 없습니다: %s", id))
	}

	if p.State == model.StateRunning {
		p.StopReason = model.StopByRestart
		m.stopProcess(p)
	} else {
		p.State      = model.StateWaitTime
		p.StopReason = model.StopNone
		Reset(p)
		m.startProcess(p, time.Now())
	}
	return okResp(fmt.Sprintf("[%s] 재기동 요청됨", id))
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수 — 프로세스 기동/종료
////////////////////////////////////////////////////////////////////////////////

/// startProcess 는 프로세스를 실제로 기동한다.
/// exec.Command 실패 시 비정상 종료와 동일하게 StartCount 증가 + MaxCount 판단 + Back-off 적용.
///
/// @param  p    기동할 PInfo 포인터
/// @param  now  현재 시각
func (m *Manager) startProcess(p *model.PInfo, now time.Time) {
	if err := Start(p, m.globalEnv, m.exitCh); err != nil {
		/// exec 실패도 비정상 종료와 동일하게 처리
		p.StartCount++
		tlog.Error("event=start_failed id=%s fail_count=%d err=%s",
			p.Config.ID, p.StartCount-1, err.Error())

		maxCount := p.Config.Restart.MaxCount
		failures  := p.StartCount - 1

		if maxCount > 0 && failures >= maxCount {
			p.State      = model.StateDisabled
			p.StopReason = model.StopByMaxCount
			tlog.ProcessDisabled(p.Config.ID, maxCount)
			return
		}

		nextState := CalcBackoff(p, now)
		p.State    = nextState
		return
	}
	/// 정상 기동 시 Back-off FailCount 초기화 (StartCount 는 유지)
	Reset(p)
}

/// stopProcess 는 프로세스에 종료 신호를 발송한다.
/// 실제 종료 대기는 goroutine + exitCh 채널로 처리한다.
///
/// @param  p  종료할 PInfo 포인터
func (m *Manager) stopProcess(p *model.PInfo) {
	if p.Pid <= 0 || p.Cmd == nil {
		return
	}

	graceSec := p.Config.Restart.GracePeriod
	if graceSec <= 0 {
		graceSec = 10
	}

	tlog.ProcessStopReq(p.Config.ID, p.Pid, graceSec)
	p.State = model.StateStopping

	/// SIGTERM 발송 후 grace_period 타이머 goroutine 기동
	/// goroutine 은 timer 만료 시 SIGKILL 을 발송하고 종료한다.
	/// 실제 종료는 runner 의 cmd.Wait() → exitCh 로 전달된다.
	proc := p.Cmd.Process
	go func() {
		_ = proc.Signal(syscallSIGTERM())
		time.Sleep(time.Duration(graceSec) * time.Second)
		/// grace_period 후에도 살아있으면 SIGKILL
		/// (이미 종료된 경우 Kill() 은 에러를 반환하지만 무시)
		_ = proc.Kill()
	}()
}

/// handleCascade 는 프로세스 비정상 종료 시 종속 프로세스를 처리한다.
///
/// @param  p  종료된 프로세스 PInfo 포인터
func (m *Manager) handleCascade(p *model.PInfo) {
	if p.Config.Cascade.OnDeath == model.CascadeNone {
		return
	}

	for _, dep := range p.Dependents {
		tlog.ProcessCascade(p.Config.ID, dep.Config.ID,
			string(p.Config.Cascade.OnDeath))

		if p.Config.Cascade.OnDeath == model.CascadeStopDependents {
			if dep.State == model.StateRunning {
				dep.StopReason = model.StopByReload
				m.stopProcess(dep)
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수 — 유틸리티
////////////////////////////////////////////////////////////////////////////////

/// findByID 는 ID 로 PInfo 를 찾아 반환한다. 없으면 nil.
func (m *Manager) findByID(id string) *model.PInfo {
	for _, p := range m.processes {
		if p.Config.ID == id {
			return p
		}
	}
	return nil
}

/// findByPtr 는 포인터 동일성으로 PInfo 를 찾는다.
/// handleExit 에서 리로드 중 제거된 프로세스를 감지할 때 사용한다.
/// 없으면 nil 반환.
func (m *Manager) findByPtr(target *model.PInfo) *model.PInfo {
	for _, p := range m.processes {
		if p == target {
			return p
		}
	}
	return nil
}

/// filterTargets 는 id 가 빈 문자열이면 전체, 아니면 해당 ID 만 반환한다.
func (m *Manager) filterTargets(id string) []*model.PInfo {
	if id == "" {
		return m.processes
	}
	if p := m.findByID(id); p != nil {
		return []*model.PInfo{p}
	}
	return nil
}

/// extractConfigs 는 PInfo 슬라이스에서 ProcessConfig 슬라이스를 추출한다.
/// 현재는 handleReload 에서 직접 구성하므로 미사용. 호환성을 위해 유지.
func extractConfigs(procs []*model.PInfo) []model.ProcessConfig {
	cfgs := make([]model.ProcessConfig, len(procs))
	for i, p := range procs {
		cfgs[i] = p.Config
	}
	return cfgs
}

/// syscallSIGTERM 은 syscall.SIGTERM 을 반환한다.
/// import cycle 방지를 위해 래퍼 함수로 분리한다.
func syscallSIGTERM() syscall.Signal {
	return syscall.SIGTERM
}

/// okResp 는 성공 응답을 생성한다.
func okResp(msg string) model.ControlResponse {
	return model.ControlResponse{Result: "ok", Message: msg}
}

/// errResp 는 에러 응답을 생성한다.
func errResp(msg string) model.ControlResponse {
	return model.ControlResponse{Result: "error", Message: msg}
}

/// idOrAll 은 id 가 빈 문자열이면 "all" 을 반환한다.
func idOrAll(id string) string {
	if strings.TrimSpace(id) == "" {
		return "all"
	}
	return id
}

/// mergeEnvMap 은 전역 env 와 프로세스별 env 를 병합한 맵을 반환한다.
/// 프로세스별 env 가 전역 env 를 덮어쓴다.
/// runner.mergeEnv 와 동일한 우선순위이나 []string 대신 map 을 반환한다.
///
/// @param  global  전역 환경변수 맵
/// @param  local   프로세스별 환경변수 맵 (nil 허용)
/// @return         병합된 환경변수 맵 (복사본)
func mergeEnvMap(global, local map[string]string) map[string]string {
	merged := make(map[string]string, len(global)+len(local))
	for k, v := range global {
		merged[k] = v
	}
	for k, v := range local {
		merged[k] = v ///< 프로세스별이 전역을 덮어씀
	}
	return merged
}
