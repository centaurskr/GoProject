///
/// @file    process/scheduler.go
/// @brief   프로세스 기동/종료 시각 및 요일 판단
/// @details 매 2초 tick 마다 manager 에서 호출되어
///          각 프로세스의 현재 상태에 따라 취해야 할 Action 을 반환한다.
///
///          설계 원칙:
///            1. 순수 함수(pure function) 로 설계한다.
///               상태 변경 없이 Action 만 반환하며 실제 처리는 manager 가 담당한다.
///            2. 일자 변경 감지는 Scheduler 구조체가 담당한다.
///            3. 자정을 넘기는 운영 구간(st > et) 을 정확히 처리한다.
///               예) st=23:00:00 et=01:30:00 → 당일 23시 ~ 익일 01시 30분
///            4. 주말(토/일) 비가동 정책은 일자 변경 시점에 일괄 적용한다.
///
///          Action 종류:
///            ActionNone         아무것도 하지 않음
///            ActionStart        기동 시작
///            ActionStop         종료 (시간 만료)
///            ActionStopHoliday  종료 (주말 정책)
///            ActionDateChange   일자 변경 처리 요청
///
///          C버전 CheckProcessHandler() 대응표:
///            일자 변경 감지    →  Tick() 내 lastDay 비교
///            주말 종료         →  ActionStopHoliday
///            시간 만료 종료    →  ActionStop (StopByTimeUp)
///            기동 판단         →  ActionStart
///
/// @date    2026-04-27
/// @author  Cento
///

package process

import (
	"fmt"
	"time"

	"goTbnova/model"
)

////////////////////////////////////////////////////////////////////////////////
// 타입 정의
////////////////////////////////////////////////////////////////////////////////

/// ActionType 은 스케줄러가 반환하는 처리 요청 종류이다.
type ActionType int

const (
	ActionNone        ActionType = iota ///< 현재 상태 유지
	ActionStart                         ///< 프로세스 기동 요청
	ActionStop                          ///< 시간 만료 종료 요청
	ActionStopHoliday                   ///< 주말 정책 종료 요청
	ActionDateChange                    ///< 일자 변경 처리 요청
)

/// Action 은 스케줄러가 특정 프로세스에 대해 요청하는 처리를 담는다.
type Action struct {
	Type ActionType   ///< 처리 종류
	P    *model.PInfo ///< 대상 프로세스 (ActionDateChange 는 nil)
}

////////////////////////////////////////////////////////////////////////////////
// Scheduler 구조체
////////////////////////////////////////////////////////////////////////////////

/// Scheduler 는 프로세스 기동/종료 판단을 담당한다.
/// manager 가 단독으로 소유하며 Run() goroutine 에서만 접근한다.
type Scheduler struct {
	lastDay int ///< 마지막으로 처리한 일(day) 숫자 (일자 변경 감지용)
}

/// NewScheduler 는 Scheduler 를 생성한다.
///
/// @return  Scheduler 포인터
func NewScheduler() *Scheduler {
	return &Scheduler{lastDay: time.Now().Day()}
}

////////////////////////////////////////////////////////////////////////////////
// 공개 함수
////////////////////////////////////////////////////////////////////////////////

/// Tick 은 매 2초 tick 마다 manager 에서 호출된다.
/// 모든 프로세스를 순회하여 취해야 할 Action 슬라이스를 반환한다.
///
/// 처리 순서:
///   1. 일자 변경 감지 → ActionDateChange 를 첫 번째로 추가
///   2. 각 프로세스 상태별 판단
///      - StateRunning   : 종료 시각 초과 여부 확인
///      - StateWaitTime  : 기동 시각 도달 + 종속 조건 확인
///      - StateWaitDepend: 종속 조건 충족 여부 확인
///      - StateBackoff   : Back-off 대기 완료 여부 확인
///      - StateStopping  : 대기 (runner.Stop() 이 처리 중)
///      - StateDisabled  : 무시
///
/// @param  procs  전체 프로세스 목록
/// @param  now    현재 시각
/// @return        처리 요청 Action 슬라이스
func (s *Scheduler) Tick(procs []*model.PInfo, now time.Time) []Action {
	actions := make([]Action, 0, len(procs))

	/// 일자 변경 감지
	if s.lastDay != now.Day() {
		actions = append(actions, Action{Type: ActionDateChange})
		s.lastDay = now.Day()
	}

	isWeekend := isWeekendDay(now)

	for _, p := range procs {
		action := s.evaluate(p, now, isWeekend, procs)
		if action.Type != ActionNone {
			actions = append(actions, action)
		}
	}
	return actions
}

/// AdvanceDay 는 일자 변경 시 각 프로세스의 Start/End 시각을 다음 날로 갱신한다.
/// manager 가 ActionDateChange 수신 후 호출한다.
///
/// 처리 대상:
///   - ModePermanent  : 갱신 불필요 (시각 제한 없음)
///   - ModeRestricted : Start, End 를 +24시간
///   - ModeOnce       : Start, End 를 +24시간 (다음 날 실행 준비)
///
/// @param  procs    전체 프로세스 목록
/// @param  now      새 날짜 기준 시각
/// @param  isWeekend 새 날짜가 주말인지 여부
func AdvanceDay(procs []*model.PInfo, now time.Time, isWeekend bool) {
	for _, p := range procs {
		if p.Config.Mode == model.ModePermanent {
			continue
		}

		/// 기동/종료 시각을 새 날짜 기준으로 재계산
		st, et, err := parseTodayTimeRange(p.Config.St, p.Config.Et, now)
		if err != nil {
			continue
		}
		p.Start = st
		p.End   = et

		/// 실행 중인 프로세스의 주말 정책 처리는
		/// Tick() → ActionStopHoliday 로 반환한다
		/// 대기 중인 프로세스는 횟수와 상태 초기화
		if p.State == model.StateWaitTime || p.State == model.StateWaitDepend {
			p.StartCount  = 0
			p.StopReason  = model.StopNone
		}
	}
}

/// RescheduleModeOnce 는 ModeOnce 프로세스가 당일 실행 완료 후
/// 다음 날 실행을 위해 StateDisabled 로 전환한다.
/// manager 가 ModeOnce 프로세스 종료 이벤트 수신 후 호출한다.
///
/// @param  p  처리할 PInfo 포인터
func RescheduleModeOnce(p *model.PInfo) {
	/// ModeOnce 는 당일 1회 실행 후 다음 날까지 대기
	/// StateDisabled 가 아닌 StateWaitTime 으로 전환하여
	/// AdvanceDay() 에서 다음 날 기동을 준비한다
	p.State      = model.StateWaitTime
	p.StopReason = model.StopByOperator ///< 정상 완료
	p.Pid        = 0
	p.Cmd        = nil
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수 — 프로세스 상태별 판단
////////////////////////////////////////////////////////////////////////////////

/// evaluate 는 프로세스 하나에 대해 Action 을 판단한다.
///
/// @param  p          판단할 PInfo 포인터
/// @param  now        현재 시각
/// @param  isWeekend  현재 날짜가 주말인지 여부
/// @param  allProcs   전체 프로세스 목록 (종속 조건 확인용)
/// @return            Action
func (s *Scheduler) evaluate(
	p         *model.PInfo,
	now       time.Time,
	isWeekend bool,
	allProcs  []*model.PInfo,
) Action {

	none := Action{Type: ActionNone}

	switch p.State {

	/// ── 실행 중 ────────────────────────────────────────────
	case model.StateRunning:
		return s.evaluateRunning(p, now, isWeekend)

	/// ── 기동 시각 대기 ──────────────────────────────────────
	case model.StateWaitTime:
		return s.evaluateWaitTime(p, now, isWeekend, allProcs)

	/// ── 종속 프로세스 대기 ──────────────────────────────────
	case model.StateWaitDepend:
		return s.evaluateWaitDepend(p, now, isWeekend, allProcs)

	/// ── Back-off 대기 ───────────────────────────────────────
	case model.StateBackoff:
		return s.evaluateBackoff(p, now, isWeekend, allProcs)

	/// ── 종료 중 / 비활성화 ──────────────────────────────────
	case model.StateStopping, model.StateDisabled:
		return none

	default:
		return none
	}
}

/// evaluateRunning 은 StateRunning 프로세스의 종료 조건을 판단한다.
///
/// 종료 조건:
///   1. ModeRestricted 이고 현재 시각이 운영 구간 외
///   2. 주말 비가동 정책(weekend=true) 이고 현재 날짜가 주말
func (s *Scheduler) evaluateRunning(
	p         *model.PInfo,
	now       time.Time,
	isWeekend bool,
) Action {

	/// 주말 비가동 정책 확인
	if p.Config.Weekend && isWeekend {
		return Action{Type: ActionStopHoliday, P: p}
	}

	/// ModePermanent 는 시간 제한 없음
	if p.Config.Mode == model.ModePermanent {
		return Action{Type: ActionNone}
	}

	/// ModeRestricted: 운영 구간 외이면 종료
	if !inTimeRange(now, p.Start, p.End) {
		return Action{Type: ActionStop, P: p}
	}

	return Action{Type: ActionNone}
}

/// evaluateWaitTime 은 StateWaitTime 프로세스의 기동 조건을 판단한다.
///
/// 기동 조건:
///   - 주말 비가동이면 기동하지 않음
///   - ModePermanent: 즉시 기동 (시각 조건 없음)
///   - ModeRestricted: 운영 구간 내이고 종속 조건 충족
///   - ModeOnce: 기동 시각 도달하고 당일 미기동이고 종속 조건 충족
func (s *Scheduler) evaluateWaitTime(
	p         *model.PInfo,
	now       time.Time,
	isWeekend bool,
	allProcs  []*model.PInfo,
) Action {

	/// 주말 비가동 정책
	if p.Config.Weekend && isWeekend {
		return Action{Type: ActionNone}
	}

	/// 시각 조건 확인
	switch p.Config.Mode {
	case model.ModePermanent:
		/// 시각 제한 없음 — 종속 조건만 확인

	case model.ModeRestricted:
		/// 운영 구간 내에 있어야 함
		if !inTimeRange(now, p.Start, p.End) {
			return Action{Type: ActionNone}
		}

	case model.ModeOnce:
		/// 기동 시각 도달 + 당일 미기동(StartCount == 0)
		if now.Before(p.Start) || p.StartCount > 0 {
			return Action{Type: ActionNone}
		}
	}

	/// 종속 조건 확인
	if !DependsSatisfied(p, allProcs, now) {
		/// StateWaitDepend 로 전환 요청
		p.State = model.StateWaitDepend
		return Action{Type: ActionNone}
	}

	return Action{Type: ActionStart, P: p}
}

/// evaluateWaitDepend 는 StateWaitDepend 프로세스의 종속 조건 충족을 확인한다.
///
/// 조건 충족 시 StateWaitTime 으로 복귀하여 다음 tick 에서 기동 판단을 받는다.
func (s *Scheduler) evaluateWaitDepend(
	p         *model.PInfo,
	now       time.Time,
	isWeekend bool,
	allProcs  []*model.PInfo,
) Action {

	/// 주말 비가동
	if p.Config.Weekend && isWeekend {
		p.State = model.StateWaitTime
		return Action{Type: ActionNone}
	}

	/// 종속 조건 충족 확인
	if DependsSatisfied(p, allProcs, now) {
		/// StateWaitTime 으로 복귀 → 다음 tick 에서 기동 판단
		p.State = model.StateWaitTime
		return Action{Type: ActionNone}
	}

	return Action{Type: ActionNone}
}

/// evaluateBackoff 는 StateBackoff 프로세스의 재기동 가능 여부를 확인한다.
func (s *Scheduler) evaluateBackoff(
	p         *model.PInfo,
	now       time.Time,
	isWeekend bool,
	allProcs  []*model.PInfo,
) Action {

	/// 주말 비가동
	if p.Config.Weekend && isWeekend {
		return Action{Type: ActionNone}
	}

	/// Back-off 대기 완료 확인
	if !IsReady(p, now) {
		return Action{Type: ActionNone}
	}

	/// 시각 조건 재확인 (Back-off 대기 중 운영 시각을 벗어났을 수 있음)
	if p.Config.Mode == model.ModeRestricted && !inTimeRange(now, p.Start, p.End) {
		p.State = model.StateWaitTime
		return Action{Type: ActionNone}
	}

	/// 종속 조건 확인
	if !DependsSatisfied(p, allProcs, now) {
		p.State = model.StateWaitDepend
		return Action{Type: ActionNone}
	}

	return Action{Type: ActionStart, P: p}
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수 — 시각/요일 유틸리티
////////////////////////////////////////////////////////////////////////////////

/// inTimeRange 는 now 가 start ~ end 운영 구간 내인지 확인한다.
/// start > end 인 경우(자정 초과 구간) 를 정확히 처리한다.
///
/// 예) start=23:00, end=01:30
///   - 23:30 → true  (당일 23시 이후)
///   - 01:00 → true  (익일 01시 이전)
///   - 02:00 → false (구간 밖)
///
/// @param  now    현재 시각
/// @param  start  운영 시작 시각
/// @param  end    운영 종료 시각
/// @return        운영 구간 내 여부
func inTimeRange(now, start, end time.Time) bool {
	if start.Before(end) || start.Equal(end) {
		/// 일반 구간: start <= now < end
		return !now.Before(start) && now.Before(end)
	}
	/// 자정 초과 구간: now >= start 또는 now < end
	return !now.Before(start) || now.Before(end)
}

/// isWeekendDay 는 주어진 시각이 토요일 또는 일요일인지 반환한다.
///
/// @param  t  확인할 시각
/// @return    토/일 여부
func isWeekendDay(t time.Time) bool {
	wd := t.Weekday()
	return wd == time.Saturday || wd == time.Sunday
}

/// parseTodayTimeRange 는 HH:MM:SS 문자열을 오늘 날짜 기준 time.Time 으로 변환한다.
/// 자정 초과 구간(st > et) 이면 end 를 익일로 처리한다.
/// AdvanceDay() 에서 날짜 갱신 시 사용한다.
///
/// @param  st   시작 시각 문자열 "HH:MM:SS"
/// @param  et   종료 시각 문자열 "HH:MM:SS"
/// @param  base 기준 날짜 시각
/// @return      start, end time.Time
/// @return      에러
func parseTodayTimeRange(st, et string, base time.Time) (time.Time, time.Time, error) {
	var shh, smm, sss int
	var ehh, emm, ess int

	if _, err := fmt.Sscanf(st, "%02d:%02d:%02d", &shh, &smm, &sss); err != nil {
		return time.Time{}, time.Time{}, err
	}
	if _, err := fmt.Sscanf(et, "%02d:%02d:%02d", &ehh, &emm, &ess); err != nil {
		return time.Time{}, time.Time{}, err
	}

	loc   := base.Location()
	start := time.Date(base.Year(), base.Month(), base.Day(), shh, smm, sss, 0, loc)
	end   := time.Date(base.Year(), base.Month(), base.Day(), ehh, emm, ess, 0, loc)

	/// 자정 초과 구간 처리
	if start.After(end) {
		end = end.Add(24 * time.Hour)
	}
	return start, end, nil
}
