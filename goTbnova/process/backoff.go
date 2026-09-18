///
/// @file    process/backoff.go
/// @brief   프로세스 재기동 Back-off 상태 관리
/// @details 프로세스 비정상 종료 시 재기동 간격을 점진적으로 늘려
///          시스템 부하와 로그 폭주를 방지한다.
///
///          Back-off 대기 시간 수열 (초):
///            실패 1회 →   1초
///            실패 2회 →   2초
///            실패 3회 →   4초
///            실패 4회 →   8초
///            실패 5회 →  16초
///            실패 6회 →  32초
///            실패 7회 → 300초 (이후 상한 유지)
///
///          상태 전이:
///
///            [비정상 종료 발생]
///            Calc(p) 호출
///                │
///                ├── backoff=false → NextRetry = 즉시 (now)
///                │                   StateBackoff 로 전환 안함
///                │
///                └── backoff=true  → FailCount++
///                                    NextRetry = now + backoffTable[FailCount]
///                                    StateBackoff 로 전환
///
///            [스케줄러 2초 tick]
///            IsReady(p) 호출
///                │
///                ├── NextRetry <= now → true  (재기동 가능)
///                └── NextRetry >  now → false (아직 대기 중)
///
///            [정상 기동 후 안정적으로 실행 중]
///            Reset(p) 호출
///                └── FailCount = 0, NextRetry = zero
///
///          MaxCount 초과 처리:
///            Calc() 내부에서 FailCount >= MaxCount 이면
///            StateDisabled 를 반환하여 manager 가 비활성화 처리한다.
///            MaxCount == 0 이면 무제한으로 재기동을 시도한다.
///
/// @date    2025-01-04
/// @author  Cento
///

package process

import (
	"time"

	"goTbnova/model"
	"goTbnova/tlog"
)

////////////////////////////////////////////////////////////////////////////////
// 상수 정의
////////////////////////////////////////////////////////////////////////////////

/// backoffTable 은 실패 횟수별 대기 시간(초) 수열이다.
/// 인덱스는 0 부터 시작하며 FailCount-1 로 참조한다.
var backoffTable = []time.Duration{
	1   * time.Second, ///< 실패 1회
	2   * time.Second, ///< 실패 2회
	4   * time.Second, ///< 실패 3회
	8   * time.Second, ///< 실패 4회
	16  * time.Second, ///< 실패 5회
	32  * time.Second, ///< 실패 6회
	300 * time.Second, ///< 실패 7회 이상 (상한)
}

/// slowStartDelay 는 StartMode=S (Slow) 일 때 추가 대기 시간이다.
const slowStartDelay = 2 * time.Second

////////////////////////////////////////////////////////////////////////////////
// 공개 함수
////////////////////////////////////////////////////////////////////////////////

/// CalcBackoff 는 비정상 종료 후 Back-off 대기 시간을 계산한다.
///
/// 역할:
///   Back-off 대기 시간(NextRetry) 계산 전용.
///   MaxCount 판단은 handleExit 에서 StartCount 기준으로 수행한다.
///
/// 처리 순서:
///   1. FailCount 증가 (대기 시간 인덱스용)
///   2. backoff=false 이면 즉시 재기동 (NextRetry = now)
///   3. backoff=true  이면 backoffTable[FailCount] 로 NextRetry 설정
///
/// @param  p    갱신할 PInfo 포인터
/// @param  now  현재 시각
/// @return      StateBackoff (항상. MaxCount 판단은 호출자 책임)
func CalcBackoff(p *model.PInfo, now time.Time) model.ProcessState {

	p.Backoff.FailCount++

	/// backoff=false: 즉시 재기동 (NextRetry = now)
	if !p.Config.Restart.Backoff {
		p.Backoff.NextRetry = now
		return model.StateBackoff
	}

	/// backoff=true: 대기 시간 계산
	delay := calcDelay(p.Backoff.FailCount)

	/// StartMode=S (Slow) 이면 추가 2초 적용
	if p.Config.StMode == model.StartSlow {
		delay += slowStartDelay
	}

	p.Backoff.NextRetry = now.Add(delay)

	tlog.ProcessBackoff(
		p.Config.ID,
		p.StartCount,        ///< 로그에는 StartCount (비정상 종료 누적 횟수) 사용
		int(delay.Seconds()),
		p.Backoff.NextRetry.Format("15:04:05"),
	)

	return model.StateBackoff
}

/// IsReady 는 Back-off 대기가 끝나 재기동 가능 여부를 반환한다.
/// 스케줄러의 2초 tick 에서 StateBackoff 인 프로세스에 대해 호출한다.
///
/// @param  p    확인할 PInfo 포인터
/// @param  now  현재 시각
/// @return      true: 재기동 가능, false: 아직 대기 중
func IsReady(p *model.PInfo, now time.Time) bool {
	return !now.Before(p.Backoff.NextRetry)
}

/// Reset 은 Back-off 상태를 초기화한다.
/// 프로세스가 정상적으로 기동하여 StateRunning 이 되면 호출한다.
///
/// @param  p  초기화할 PInfo 포인터
func Reset(p *model.PInfo) {
	p.Backoff.FailCount = 0
	p.Backoff.NextRetry = time.Time{}
}

/// DependsSatisfied 는 프로세스의 모든 선행 종속 프로세스가
/// StateRunning 상태인지 확인한다.
///
/// depends_healthy=true 이면 Running 진입 후 추가로
/// 최소 1 tick(2초) 이상 생존했는지도 확인한다.
///
/// @param  p         확인할 PInfo 포인터
/// @param  allProcs  전체 프로세스 목록 (ID 탐색용)
/// @param  now       현재 시각
/// @return           true: 모든 선행 조건 충족, false: 대기 필요
func DependsSatisfied(p *model.PInfo, allProcs []*model.PInfo, now time.Time) bool {
	if len(p.Config.DependsOn) == 0 {
		return true
	}

	/// ID → PInfo 빠른 탐색
	idx := buildProcIndex(allProcs)

	for _, depID := range p.Config.DependsOn {
		dep, ok := idx[depID]
		if !ok {
			/// 존재하지 않는 ID (validator 에서 잡혔어야 하지만 방어 처리)
			tlog.Error("event=depends_error id=%s dep=%s reason=ID없음",
				p.Config.ID, depID)
			return false
		}

		/// 선행 프로세스가 Running 이어야 함
		if dep.State != model.StateRunning {
			return false
		}

		/// depends_healthy=true: 최소 1 tick 이상 안정적으로 실행 중이어야 함
		/// LastStart 로부터 2초(1 tick) 이상 경과했는지 확인
		if p.Config.DependsHealthy {
			if now.Sub(dep.LastStart) < 2*time.Second {
				return false
			}
		}
	}
	return true
}

/// SnapshotBackoff 는 PInfo 의 Back-off 상태를 ProcessSnapshot 에 반영한다.
/// manager 가 모니터 응답을 생성할 때 호출한다.
///
/// @param  p         원본 PInfo
/// @param  snapshot  갱신할 ProcessSnapshot 포인터
func SnapshotBackoff(p *model.PInfo, snapshot *model.ProcessSnapshot) {
	if p.State == model.StateBackoff && !p.Backoff.NextRetry.IsZero() {
		t := p.Backoff.NextRetry
		snapshot.NextRetry = &t
	}
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수
////////////////////////////////////////////////////////////////////////////////

/// calcDelay 는 실패 횟수에 따른 Back-off 대기 시간을 반환한다.
///
/// @param  failCount  연속 실패 횟수 (1 부터 시작)
/// @return            대기 시간 (time.Duration)
func calcDelay(failCount int) time.Duration {
	idx := failCount - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(backoffTable) {
		idx = len(backoffTable) - 1
	}
	return backoffTable[idx]
}

/// buildProcIndex 는 프로세스 ID → PInfo 포인터 맵을 구성한다.
/// DependsSatisfied 내부에서만 사용한다.
///
/// @param  procs  전체 프로세스 목록
/// @return        ID → *PInfo 맵
func buildProcIndex(procs []*model.PInfo) map[string]*model.PInfo {
	idx := make(map[string]*model.PInfo, len(procs))
	for _, p := range procs {
		idx[p.Config.ID] = p
	}
	return idx
}
