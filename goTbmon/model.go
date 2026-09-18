// model.go — goTbnova TCP 프로토콜 데이터 모델 및 상태 변환 유틸리티
//
// goTbnova 제어 소켓과 주고받는 JSON 구조체를 정의한다.
//
// 요청 흐름:
//   ControlCmd  →  JSON 직렬화  →  TCP 전송
//   TCP 수신    →  JSON 역직렬화  →  ControlResponse
//
// 주요 타입:
//   ControlCmd       — 클라이언트 → 데몬 요청
//   ControlResponse  — 데몬 → 클라이언트 응답 (ProcessSnapshot 목록 포함)
//   ProcessSnapshot  — 개별 프로세스의 완전한 스냅샷
package main

import "time"

// goTbnova TCP 프로토콜 구조체

// ControlCmd: 데몬에 보내는 제어 명령. cmd 필드가 동작을 결정한다.
// id가 비어 있으면 전체(all) 대상으로 해석하는 명령도 있다.
type ControlCmd struct {
	Cmd string `json:"cmd"`
	ID  string `json:"id,omitempty"`
}

// RestartConfig: 프로세스 재시작 정책 설정.
type RestartConfig struct {
	Policy      string `json:"policy"`
	Backoff     bool   `json:"backoff"`
	MaxCount    int    `json:"max_count"`
	GracePeriod int    `json:"grace_period"`
}

// CascadeConfig: 프로세스 종료 시 연쇄 동작 설정.
type CascadeConfig struct {
	OnDeath string `json:"on_death"` // stop_dependents / notify / ""
}

// ProcessSnapshot: 단일 프로세스의 전체 상태 스냅샷.
// goTbnova가 status 명령 응답으로 반환하는 JSON 객체와 1:1 대응한다.
type ProcessSnapshot struct {
	ID             string            `json:"id"`
	Group          string            `json:"group"`
	Desc           string            `json:"desc"`
	Mode           string            `json:"mode"`            // P=Process 등
	Weekend        bool              `json:"weekend"`         // 주말 실행 여부
	StMode         string            `json:"stmode"`          // 스케줄 모드
	St             string            `json:"st"`              // 시작 시각 "HH:MM:SS"
	Et             string            `json:"et"`              // 종료 시각 "HH:MM:SS"
	Path           string            `json:"path"`            // 실행 파일 경로
	Args           []string          `json:"args"`            // 실행 인수
	DependsOn      []string          `json:"depends_on,omitempty"`
	DependsHealthy bool              `json:"depends_healthy"` // 선행 프로세스 전원 실행 중이어야 하는지
	Restart        RestartConfig     `json:"restart"`
	Cascade        CascadeConfig     `json:"cascade"`
	Env            map[string]string `json:"env,omitempty"`   // 프로세스 전용 환경변수
	State          int               `json:"state"`
	Pid            int               `json:"pid"`
	StartCount     int               `json:"start_count"`
	StopReason     int               `json:"stop_reason"`
	LastStart      time.Time         `json:"last_start"`
	NextRetry      *time.Time        `json:"next_retry"`
}

// ControlResponse: 데몬의 응답 봉투.
// Result == "ok" 이면 정상, "error" 이면 Message에 오류 설명이 담긴다.
type ControlResponse struct {
	Result    string            `json:"result"`
	Message   string            `json:"message"`
	Processes []ProcessSnapshot `json:"processes"`
	Env       map[string]string `json:"env,omitempty"`
}

// stateName: 목록·트리 화면의 STATE 열에 쓸 짧은 문자열을 반환한다.
func stateName(s int) string {
	switch s {
	case 0:
		return "WT" // WAIT_TIME
	case 1:
		return "WD" // WAIT_DEPEND
	case 2:
		return "RUN" // RUNNING
	case 3:
		return "STOP" // STOPPING
	case 4:
		return "BK" // BACKOFF
	case 5:
		return "DIS" // DISABLED
	default:
		return "?"
	}
}

// stateNameFull: 상세 화면(h)의 State 필드에 쓸 전체 이름을 반환한다.
func stateNameFull(s int) string {
	switch s {
	case 0:
		return "WAIT_TIME"
	case 1:
		return "WAIT_DEPEND"
	case 2:
		return "RUNNING"
	case 3:
		return "STOPPING"
	case 4:
		return "BACKOFF"
	case 5:
		return "DISABLED"
	default:
		return "UNKNOWN"
	}
}

// stopReasonName: 상세 화면의 Stop Reason 필드에 쓸 전체 이름을 반환한다.
func stopReasonName(r int) string {
	switch r {
	case 0:
		return "-"
	case -1:
		return "운영자"
	case -2:
		return "휴일"
	case -3:
		return "재시작"
	case -4:
		return "시간종료"
	case -5:
		return "Reload"
	case -6:
		return "비정상"
	case -7:
		return "횟수초과"
	default:
		return "알수없음"
	}
}
