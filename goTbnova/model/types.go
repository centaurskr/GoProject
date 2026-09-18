///
/// @file    model/types.go
/// @brief   goTbnova 핵심 데이터 구조체 및 타입 정의
/// @details 프로세스 생명주기 관리에 필요한 모든 타입, 상수, 구조체를 정의한다.
///
///          설계 원칙:
///            1. TOML 파싱 구조체(XXXConfig)와 런타임 구조체(PInfo, GInfo)를 분리한다.
///            2. C버전의 stcnt 음수값 이중 역할을 StopReason 타입으로 명시화한다.
///            3. C버전의 pid=-1 상태를 ProcessState 타입으로 명시화한다.
///            4. 프로세스 종료 감지는 cmd.Wait() goroutine + ExitEvent channel 로만 처리한다.
///               (주기적 /proc 폴링 없음 — cmd.Wait() 가 실시간으로 정확하게 감지)
///            5. 외부에 노출되는 상태는 ProcessSnapshot 복사본으로만 전달한다.
///
///          C버전 stcnt 음수값 대응표:
///            stcnt == -1  →  StopByOperator
///            stcnt == -2  →  StopByHoliday
///            stcnt == -3  →  StopByRestart
///            stcnt == -4  →  StopByTimeUp
///            stcnt == -5  →  StopByReload
///            비정상 종료  →  StopByAbnormal  (신규)
///            횟수 초과    →  StopByMaxCount  (신규)
///
/// @date    2025-01-04
/// @author  Cento
///

package model

import (
	"os/exec"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// 열거형 타입 정의
////////////////////////////////////////////////////////////////////////////////

/// ProcessMode 는 프로세스 운영 모드를 나타낸다.
/// TOML [[process]] 의 mode 필드에 대응한다.
type ProcessMode string

const (
	ModePermanent  ProcessMode = "P" ///< 상시 운영: 시간 제한 없이 하루 종일 운영
	ModeRestricted ProcessMode = "R" ///< 시간 제한: st ~ et 구간에만 운영
	ModeOnce       ProcessMode = "O" ///< 1회 실행: 지정 시각에 단 한번만 실행
)

/// StartMode 는 프로세스 재기동 방식을 나타낸다.
/// TOML [[process]] 의 stmode 필드에 대응한다.
type StartMode string

const (
	StartQuick StartMode = "Q" ///< 종료 즉시 재기동
	StartSlow  StartMode = "S" ///< 약간의 지연(2초) 후 재기동
)

/// RestartPolicy 는 프로세스 비정상 종료 시 재기동 정책을 나타낸다.
/// TOML [process.restart] 의 policy 필드에 대응한다.
type RestartPolicy string

const (
	PolicyAlways    RestartPolicy = "always"     ///< 종료 시 항상 재기동
	PolicyOnFailure RestartPolicy = "on-failure" ///< 비정상 종료 시만 재기동
	PolicyNever     RestartPolicy = "never"      ///< 재기동 안함, 로그 기록만
)

/// CascadeAction 은 프로세스 종료 시 종속 프로세스 처리 방식을 나타낸다.
/// TOML [process.cascade] 의 on_death 필드에 대응한다.
type CascadeAction string

const (
	CascadeNone           CascadeAction = "none"            ///< 종속 프로세스에 영향 없음
	CascadeNotify         CascadeAction = "notify"          ///< 로그 기록만
	CascadeStopDependents CascadeAction = "stop_dependents" ///< 종속 프로세스 연쇄 종료
)

/// StopReason 은 프로세스 종료 원인을 나타낸다.
/// C버전에서 stcnt 음수값으로 처리하던 것을 명시적 타입으로 대체한다.
type StopReason int

const (
	StopNone       StopReason = 0  ///< 종료 없음 (정상 운영 중)
	StopByOperator StopReason = -1 ///< 운영자 수동 종료
	StopByHoliday  StopReason = -2 ///< 토/일 비가동 정책에 의한 종료
	StopByRestart  StopReason = -3 ///< 외부 Restart 요청에 의한 종료
	StopByTimeUp   StopReason = -4 ///< 운영 시간(et) 만료에 의한 종료
	StopByReload   StopReason = -5 ///< 테이블 리로드 중 정리에 의한 종료
	StopByAbnormal StopReason = -6 ///< 비정상 종료 (재기동 대기 또는 Backoff 중)
	StopByMaxCount StopReason = -7 ///< MaxCount 도달로 StateDisabled 전환
	                                ///<   StartCount - 1 >= max_count 조건 충족 시
	                                ///<   자동 재기동 완전 중단, 운영자 start 명령 필요
)

/// ProcessState 는 프로세스의 현재 운영 상태를 나타낸다.
/// C버전의 pid 값(-1, 0, >0) 과 stcnt 조합으로 판단하던 것을 단일 타입으로 통합한다.
///
///  C버전 대응표:
///    pid==0, stcnt==0          →  StateWaitTime      (시각 미도달)
///    pid==0, stcnt==0          →  StateWaitDepend    (종속 프로세스 대기, 신규)
///    pid>0,  stcnt>0           →  StateRunning
///    pid>0,  stcnt<0           →  StateStopping
///    (신규)                    →  StateBackoff
///    pid==-1                   →  StateDisabled
///
///  상태 전이:
///
///    [최초 기동 / 리로드 후]
///    StateWaitTime ──► (st 시각 도달)
///        │
///        ├── depends_on 없음  ──────────────────► StateRunning
///        └── depends_on 있음  ──► StateWaitDepend ──► StateRunning
///
///    [정상 운영 중]
///    StateRunning ──► (et 시각 도달 / 운영자 종료 / 리로드 삭제)
///        └── StateStopping ──► StateWaitTime (또는 StateDisabled)
///
///    [비정상 종료]
///    StateRunning ──► (프로세스 사망) → StartCount++
///        ├── backoff=false  ──► StateWaitTime  (즉시 재기동 대기)
///        └── backoff=true   ──► StateBackoff   ──► StateWaitTime
///
///    [비활성화]
///    StateBackoff / StateRunning ──► (MaxCount 초과 / PolicyNever) ──► StateDisabled
///        └── 운영자 "start" 명령 → StartCount=0, Reset → StateWaitTime
///
///    [테이블 리로드]
///    StateRunning  ──► ENV 만 업데이트, 상태 유지 (규칙 1)
///    StateRunning  ──► (테이블에서 삭제) → StateStopping → 목록 제거 (규칙 2)
///    StateWaitTime ──► (테이블에서 삭제) → 즉시 목록 제거 (규칙 2)
///    (없음)        ──► 신규 추가 → StateWaitTime (규칙 3)
type ProcessState int

const (
	StateWaitTime   ProcessState = iota ///< 기동 시각(st) 미도달 대기 또는 재기동 대기
	                                    ///<   ModePermanent  : 기동 직후 즉시 StateRunning 으로 전환
	                                    ///<   ModeRestricted : st 시각까지 대기
	                                    ///<   ModeOnce       : st 시각까지 대기 (당일 1회)
	StateWaitDepend                     ///< 종속 프로세스(depends_on) 기동 대기
	                                    ///<   depends_on 목록의 모든 프로세스가 StateRunning 이 되면 해제
	                                    ///<   depends_healthy=true 이면 2초 생존 확인까지 대기
	StateRunning                        ///< 정상 실행 중
	                                    ///<   리로드 시 ENV 만 업데이트 (재기동 없음, 규칙 1)
	StateStopping                       ///< 종료 신호(SIGTERM) 발송 후 grace_period 대기 중
	                                    ///<   grace_period 초과 시 SIGKILL 발송
	StateBackoff                        ///< 비정상 종료 후 Back-off 대기 중
	                                    ///<   NextRetry 시각 도달 시 StateWaitTime 으로 전환
	StateDisabled                       ///< 비활성화
	                                    ///<   원인 1: MaxCount 초과 (StartCount >= max_count)
	                                    ///<   원인 2: PolicyNever 이고 비정상 종료
	                                    ///<   원인 3: 운영자 "stop" 명령
	                                    ///<   해제: 운영자 "start" 명령 → StartCount=0 초기화
)

////////////////////////////////////////////////////////////////////////////////
// TOML 설정 파싱 구조체
// TOML 파일에서 읽히는 구조체이며 런타임 중에는 수정되지 않는다.
////////////////////////////////////////////////////////////////////////////////

/// RestartConfig 는 프로세스 재기동 정책 설정을 담는다.
/// TOML [process.restart] 섹션에 대응한다.
type RestartConfig struct {
	Policy      RestartPolicy `toml:"policy"`       ///< 재기동 정책 (always/on-failure/never)
	Backoff     bool          `toml:"backoff"`      ///< Back-off 적용 여부
	MaxCount    int           `toml:"max_count"`    ///< 최대 재기동 횟수 (0: 무제한)
	GracePeriod int           `toml:"grace_period"` ///< SIGTERM 후 SIGKILL 까지 대기 시간(초)
}

/// CascadeConfig 는 프로세스 종료 시 종속 프로세스 연쇄 처리 설정을 담는다.
/// TOML [process.cascade] 섹션에 대응한다.
type CascadeConfig struct {
	OnDeath CascadeAction `toml:"on_death"` ///< 종료 시 종속 프로세스 처리 방식
}

/// ProcessConfig 는 TOML [[process]] 섹션 하나를 담는다.
/// 설정 파일에서 읽히는 원본 값이며 런타임 중 변경되지 않는다.
type ProcessConfig struct {

	/// 기본 식별 정보
	ID    string `toml:"id"`    ///< 프로세스 식별자 (최대 13자, 공백 불가)
	Group string `toml:"group"` ///< 소속 그룹 ID ([groups.XXX] 의 XXX)
	Desc  string `toml:"desc"`  ///< 프로세스 설명 (40자 이내 권장)

	/// 운영 모드
	Mode    ProcessMode `toml:"mode"`    ///< 운영 모드 (P/R/O)
	Weekend bool        `toml:"weekend"` ///< true: 토/일 비가동
	StMode  StartMode   `toml:"stmode"`  ///< 재기동 방식 (Q/S)

	/// 운영 시간 (ModeRestricted, ModeOnce 에서 사용)
	/// 시작 시각이 종료 시각보다 크면 자정을 넘기는 구간으로 처리한다.
	/// 예) st=23:00:00 et=01:30:00 → 당일 23시 ~ 익일 01시 30분
	St string `toml:"st"` ///< 기동 시작 시각 "HH:MM:SS"
	Et string `toml:"et"` ///< 기동 종료 시각 "HH:MM:SS"

	/// 실행 정보
	/// C버전에서 path 필드에 혼합되던 인자를 args 슬라이스로 분리한다.
	Path string            `toml:"path"` ///< 실행 파일 경로 (절대경로 또는 def_dir 기준 상대경로)
	Args []string          `toml:"args"` ///< 실행 인자 목록
	Env  map[string]string `toml:"env"`  ///< 프로세스 전용 환경변수 (전역 env 에 추가/덮어쓰기)

	/// 기동 종속관계
	/// depends_on 목록의 모든 프로세스가 StateRunning 이 된 후 기동한다.
	/// depends_healthy=true 이면 Running 진입 후 추가로 일정 시간 생존을 확인한다.
	DependsOn      []string `toml:"depends_on"`      ///< 선행 기동 필요 프로세스 ID 목록
	DependsHealthy bool     `toml:"depends_healthy"`  ///< true: 선행 프로세스 안정 확인 후 기동

	/// 세부 정책
	Restart RestartConfig `toml:"restart"` ///< 재기동 정책
	Cascade CascadeConfig `toml:"cascade"` ///< 종속 프로세스 연쇄 처리
}

/// GroupConfig 는 TOML [groups.XXX] 섹션을 담는다.
type GroupConfig struct {
	Desc string `toml:"desc"` ///< 그룹 설명
}

/// DefaultConfig 는 TOML [default] 섹션을 담는다.
type DefaultConfig struct {
	DefDir string `toml:"def_dir"` ///< 기본 실행 파일 경로 (상대 경로 기준점)
	DefCnt int    `toml:"def_cnt"` ///< 기본 최대 재기동 횟수
}

/// TableConfig 는 TOML 설정 파일 전체 구조를 담는다.
/// config.Load() 의 최상위 파싱 대상이다.
type TableConfig struct {
	Version map[string]string      `toml:"version"` ///< 버전 정보
	Env     map[string]string      `toml:"env"`     ///< 전역 환경변수
	Default DefaultConfig          `toml:"default"` ///< 기본값
	Groups  map[string]GroupConfig `toml:"groups"`  ///< 그룹 정의 맵
	Process []ProcessConfig        `toml:"process"` ///< 프로세스 목록
}

////////////////////////////////////////////////////////////////////////////////
// 런타임 구조체
// 프로세스 생명주기 관리 중 상태가 변경되는 구조체이다.
// 모든 필드는 process.Manager.Run() goroutine 에서만 접근한다.
////////////////////////////////////////////////////////////////////////////////

/// BackoffState 는 프로세스별 Back-off 재기동 대기 상태를 관리한다.
///
/// 대기 시간 수열 (초):
///   1 → 2 → 4 → 8 → 16 → 32 → 300 (이후 상한 유지)
///
/// FailCount 는 정상 기동(StateRunning 진입) 시 0 으로 초기화된다.
type BackoffState struct {
	FailCount int       ///< 연속 비정상 종료 횟수
	NextRetry time.Time ///< 다음 재기동 시도 가능 시각
}

/// PInfo 는 런타임에서 관리하는 프로세스 정보를 담는다.
///
///  구성:
///    Config      — TOML 에서 로드된 설정값 (불변)
///    런타임 필드 — 실행 중 변경되는 상태값
///    내부 필드   — manager/runner 내부에서만 사용
///
///  종료 감지 방식:
///    runner.go 에서 cmd.Wait() 를 별도 goroutine 으로 대기한다.
///    종료 시 manager 의 exitCh 채널로 ExitEvent 를 전달하고 manager 가 처리한다.
///    /proc 폴링 없이 커널 이벤트를 실시간으로 수신한다.
type PInfo struct {

	/// 설정값 (TOML 로드 후 불변)
	Config ProcessConfig ///< 원본 TOML 설정

	/// 런타임 상태
	Pid        int          ///< 현재 PID (0: 미기동)
	State      ProcessState ///< 현재 운영 상태
	StopReason StopReason   ///< 가장 최근 종료 원인
	StartCount int          ///< 비정상 종료 횟수. 초기값=1. 비정상 종료 시마다 증가.
	                         ///<   1       : 초기 상태 (정상 운영 중)
	                         ///<   2,3,... : 누적 비정상 종료 횟수
	                         ///<   "start" 명령 수신 시 1 로 재설정 (규칙 4)
	                         ///<   MaxCount 도달 시 StateDisabled 전환
	Start      time.Time    ///< 당일 기동 예정 시각 (st 파싱 결과)
	End        time.Time    ///< 당일 종료 예정 시각 (et 파싱 결과)
	LastStart  time.Time    ///< 최종 실제 기동 시각

	/// 종속관계 (설정 로드 후 manager 가 런타임에 구성)
	Dependents []*PInfo ///< 이 프로세스 종료 시 영향받는 프로세스 목록

	/// 내부 제어 (manager/runner 에서만 접근)
	Cmd     *exec.Cmd    ///< 실행 중인 프로세스 핸들 (nil: 미기동)
	Backoff BackoffState ///< Back-off 상태
}

/// GInfo 는 런타임에서 관리하는 그룹 정보를 담는다.
type GInfo struct {
	Name    string   ///< 그룹 ID (TOML groups 맵의 키)
	Desc    string   ///< 그룹 설명
	Members []*PInfo ///< 소속 프로세스 목록 (런타임 구성)
}

////////////////////////////////////////////////////////////////////////////////
// 채널 통신 / JSON 프로토콜 구조체
////////////////////////////////////////////////////////////////////////////////

/// ExitEvent 는 프로세스 종료 시 runner.go 에서 manager.go 로 전달되는 이벤트다.
/// cmd.Wait() 완료 시 생성되며 manager 의 exitCh 채널로 전달된다.
type ExitEvent struct {
	P   *PInfo ///< 종료된 프로세스 정보 포인터
	Err error  ///< 종료 원인 (nil: 정상, non-nil: 비정상)
}

/// ControlCmd 는 TCP 모니터 클라이언트에서 수신하는 JSON 제어 명령을 담는다.
///
///  Cmd 목록:
///    "status"  — 전체 또는 특정 프로세스 상태 조회 (ID 생략 시 전체)
///    "start"   — 지정 프로세스 기동
///    "stop"    — 지정 프로세스 종료
///    "restart" — 지정 프로세스 재기동
///    "reload"  — 설정 파일 리로드
///    "env"     — 환경변수 조회
///                ID 생략 시: 전역 env 반환
///                ID 지정 시: 해당 프로세스의 병합 env 반환 (전역 + 프로세스별)
type ControlCmd struct {
	Cmd string `json:"cmd"` ///< 명령어
	ID  string `json:"id"`  ///< 대상 프로세스 ID (전체 대상이면 빈 문자열)
}

/// ProcessSnapshot 은 모니터 응답용 프로세스 상태 스냅샷을 담는다.
/// PInfo 를 직접 노출하지 않고 복사본으로 전달하여 동시성 문제를 방지한다.
///
/// 구성:
///   [TOML 설정 항목] process.toml 의 [[process]] 섹션 전체 항목
///   [런타임 상태]    실행 중 변경되는 상태값
///
/// Env 필드:
///   process.toml 의 프로세스별 env 가 아닌
///   전역 env + 프로세스별 env 를 병합한 최종 적용값이다.
///   (프로세스별 env 가 전역 env 를 덮어씀)
type ProcessSnapshot struct {

	/// ── TOML 설정 항목 (process.toml [[process]] 전체) ──────────────────

	/// 기본 식별 정보
	ID    string `json:"id"`    ///< 프로세스 식별자
	Group string `json:"group"` ///< 소속 그룹 ID
	Desc  string `json:"desc"`  ///< 프로세스 설명

	/// 운영 모드
	Mode    ProcessMode `json:"mode"`    ///< P/R/O
	Weekend bool        `json:"weekend"` ///< 토/일 비가동 여부
	StMode  StartMode   `json:"stmode"`  ///< 재기동 방식 Q/S

	/// 운영 시간
	St string `json:"st"` ///< 기동 시작 시각 "HH:MM:SS"
	Et string `json:"et"` ///< 기동 종료 시각 "HH:MM:SS"

	/// 실행 정보
	Path string   `json:"path"` ///< 실행 파일 경로
	Args []string `json:"args"` ///< 실행 인자 목록

	/// 기동 종속관계
	DependsOn      []string `json:"depends_on"`      ///< 선행 기동 필요 프로세스 ID 목록
	DependsHealthy bool     `json:"depends_healthy"`  ///< 선행 프로세스 안정 확인 여부

	/// 재기동 정책
	Restart RestartConfig `json:"restart"` ///< policy/backoff/max_count/grace_period

	/// 연쇄 처리 정책
	Cascade CascadeConfig `json:"cascade"` ///< on_death

	/// 병합된 환경변수 (전역 env + 프로세스별 env, 프로세스별이 우선)
	Env map[string]string `json:"env,omitempty"` ///< 최종 적용 환경변수

	/// ── 런타임 상태 ──────────────────────────────────────────────────────

	State      ProcessState `json:"state"`                ///< 현재 운영 상태
	Pid        int          `json:"pid"`                  ///< 현재 PID (0: 미기동)
	StartCount int          `json:"start_count"`          ///< 비정상 종료 누적 횟수 (초기값 1)
	StopReason StopReason   `json:"stop_reason"`          ///< 최근 종료 원인
	LastStart  time.Time    `json:"last_start"`           ///< 최종 기동 시각
	NextRetry  *time.Time   `json:"next_retry,omitempty"` ///< Back-off 중 다음 재시도 시각
}

/// ControlResponse 는 제어 명령에 대한 JSON 응답을 담는다.
type ControlResponse struct {
	Result    string            `json:"result"`               ///< "ok" 또는 "error"
	Message   string            `json:"message,omitempty"`    ///< 오류 메시지 (에러 시)
	Processes []ProcessSnapshot `json:"processes,omitempty"`  ///< 프로세스 목록 (status 명령 시)
	Env       map[string]string `json:"env,omitempty"`        ///< 환경변수 맵 (env 명령 시)
}
