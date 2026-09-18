///
/// @file    control/server.go
/// @brief   TCP 모니터 서버 — 기존 C버전 monitor client 호환
/// @details tbnova/tcp 포트로 TCP 연결을 수락하고
///          JSON 형식의 제어 명령을 수신하여 manager 로 전달한다.
///
///          설계 원칙:
///            1. /etc/services 의 tbnova/tcp 포트를 C버전과 공유한다.
///               → C버전과 동시 실행 시 bind 실패로 중복 기동을 방지한다.
///            2. 127.0.0.1 에만 바인딩한다 (로컬 전용, 외부 노출 없음).
///            3. 클라이언트 연결당 goroutine 1개로 처리한다.
///            4. 요청/응답 모두 JSON 이다.
///            5. 연결당 ReadTimeout 을 적용하여 idle 연결 자원 낭비를 방지한다.
///            6. 명령 처리는 manager.controlCh 채널로 위임한다.
///               → 상태 변경은 항상 manager.Run() goroutine 에서만 수행.
///
///          프로토콜:
///            Request  (client → server): model.ControlCmd  JSON 1행
///            Response (server → client): model.ControlResponse JSON 1행
///
///          보안:
///            - 127.0.0.1 바인딩으로 외부 접근 차단
///            - ReadTimeout 으로 연결 고갈 방지
///            - 요청 크기 제한 (4KB) 으로 메모리 고갈 방지
///
/// @date    2026-04-27
/// @author  Cento
///

package control

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"goTbnova/model"
	"goTbnova/process"
	"goTbnova/tlog"
)

////////////////////////////////////////////////////////////////////////////////
// 상수 정의
////////////////////////////////////////////////////////////////////////////////

const (
	serviceName    = "tbnova"           ///< /etc/services 서비스명
	defaultBind    = "127.0.0.1"        ///< 기본 바인딩 주소 (로컬 전용)
	readTimeout    = 30 * time.Second   ///< 클라이언트 읽기 타임아웃
	maxRequestSize = 4 * 1024           ///< 최대 요청 크기 (4KB)
	respTimeout    = 10 * time.Second   ///< manager 응답 대기 타임아웃
)

////////////////////////////////////////////////////////////////////////////////
// Server 구조체
////////////////////////////////////////////////////////////////////////////////

/// Server 는 TCP 모니터 서버를 담당한다.
type Server struct {
	listener  net.Listener                 ///< TCP 리스너
	controlCh chan<- process.ControlRequest ///< manager 제어 요청 채널
	port      int                          ///< 바인딩 포트
	bindHost  string                       ///< 바인딩 주소
}

////////////////////////////////////////////////////////////////////////////////
// 공개 함수
////////////////////////////////////////////////////////////////////////////////

/// NewServer 는 Server 를 생성하고 TCP 포트를 바인딩한다.
/// /etc/services 에서 tbnova/tcp 포트를 조회한다.
/// 포트 바인딩 실패 시 C버전 또는 다른 goTbnova 가 이미 실행 중임을 의미한다.
///
/// @param  controlCh  manager 의 제어 요청 채널
/// @param  bindHost   바인딩 IP 주소
///                    ""           → defaultBind(127.0.0.1) 사용
///                    "0.0.0.0"    → 모든 인터페이스 허용
///                    "192.168.x.x" → 특정 인터페이스만 허용
/// @return            Server 포인터
/// @return            포트 조회/바인딩 실패 시 에러
func NewServer(controlCh chan<- process.ControlRequest, bindHost string) (*Server, error) {

	/// 바인딩 주소 결정
	if bindHost == "" {
		bindHost = defaultBind
	}

	/// /etc/services 에서 tbnova/tcp 포트 조회
	port, err := lookupPort()
	if err != nil {
		return nil, err
	}

	addr := fmt.Sprintf("%s:%d", bindHost, port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf(
			"TCP 포트 바인딩 실패 [%s] — tb_nova(C) 또는 goTbnova 가 이미 실행 중입니다: %w",
			addr, err,
		)
	}

	tlog.Info("event=monitor_listen addr=%s", addr)

	return &Server{
		listener:  ln,
		controlCh: controlCh,
		port:      port,
		bindHost:  bindHost,
	}, nil
}

/// Serve 는 클라이언트 연결을 수락하고 처리한다.
/// ctx 취소 시 리스너를 닫고 반환한다.
/// main() 에서 goroutine 으로 실행된다.
///
/// @param  ctx  상위 컨텍스트 (main 에서 signal 수신 시 취소)
func (s *Server) Serve(ctx context.Context) {
	/// ctx 취소 시 리스너를 닫아 Accept() 를 해제한다
	go func() {
		<-ctx.Done()
		s.listener.Close()
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				/// 정상 종료 (ctx 취소로 인한 Accept 해제)
				return
			default:
				tlog.Error("event=monitor_accept_error err=%s", err.Error())
				continue
			}
		}

		tlog.MonitorConnect(conn.RemoteAddr().String())

		/// 클라이언트당 goroutine 1개
		go s.handleConn(conn)
	}
}

/// Port 는 바인딩된 포트 번호를 반환한다.
/// main() 에서 tlog.Startup() 에 전달하기 위해 사용한다.
///
/// @return  포트 번호
func (s *Server) Port() int {
	return s.port
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수 — 연결 처리
////////////////////////////////////////////////////////////////////////////////

/// handleConn 은 클라이언트 연결 하나를 처리한다.
/// 연결 종료 시 goroutine 이 자동으로 종료된다.
///
/// 처리 순서:
///   1. ReadTimeout 설정
///   2. JSON 요청 수신 (최대 4KB)
///   3. 요청 파싱 및 검증
///   4. manager.controlCh 로 전달
///   5. 응답 수신 후 JSON 으로 전송
///
/// @param  conn  수락된 TCP 연결
func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	addr := conn.RemoteAddr().String()

	/// ReadTimeout 설정 — idle 연결 자원 낭비 방지
	if err := conn.SetDeadline(time.Now().Add(readTimeout)); err != nil {
		tlog.Warn("event=monitor_deadline_err addr=%s err=%s", addr, err.Error())
		return
	}

	/// 요청 수신 (최대 4KB, 줄바꿈 구분)
	reader := bufio.NewReaderSize(conn, maxRequestSize)
	raw, err := reader.ReadString('\n')
	if err != nil {
		if err != io.EOF {
			tlog.Warn("event=monitor_read_err addr=%s err=%s", addr, err.Error())
		}
		return
	}

	raw = strings.TrimSpace(raw)
	if len(raw) == 0 {
		s.writeResp(conn, errResp("빈 요청"))
		return
	}

	/// JSON 파싱
	var cmd model.ControlCmd
	if err := json.Unmarshal([]byte(raw), &cmd); err != nil {
		tlog.Warn("event=monitor_parse_err addr=%s err=%s raw=%s",
			addr, err.Error(), truncate(raw, 80))
		s.writeResp(conn, errResp(fmt.Sprintf("JSON 파싱 오류: %v", err)))
		return
	}

	/// 명령어 검증
	if err := validateCmd(&cmd); err != nil {
		s.writeResp(conn, errResp(err.Error()))
		return
	}

	/// manager 로 요청 전달 및 응답 수신
	respCh := make(chan model.ControlResponse, 1)
	req    := process.ControlRequest{
		Cmd:    cmd.Cmd,
		ID:     cmd.ID,
		RespCh: respCh,
	}

	/// controlCh 전송 타임아웃
	select {
	case s.controlCh <- req:
	case <-time.After(respTimeout):
		s.writeResp(conn, errResp("manager 응답 타임아웃 (요청 전달 실패)"))
		return
	}

	/// 응답 수신 타임아웃
	select {
	case resp := <-respCh:
		/// 응답 전송 전 Deadline 갱신
		_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
		s.writeResp(conn, resp)
	case <-time.After(respTimeout):
		s.writeResp(conn, errResp("manager 응답 타임아웃 (처리 지연)"))
	}
}

/// writeResp 는 ControlResponse 를 JSON 으로 직렬화하여 클라이언트에 전송한다.
/// 응답 끝에 줄바꿈(\n) 을 추가한다.
///
/// @param  conn  전송 대상 TCP 연결
/// @param  resp  전송할 응답
func (s *Server) writeResp(conn net.Conn, resp model.ControlResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		tlog.Error("event=monitor_marshal_err err=%s", err.Error())
		return
	}
	data = append(data, '\n')

	if _, err := conn.Write(data); err != nil {
		tlog.Warn("event=monitor_write_err err=%s", err.Error())
	}
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수 — 유틸리티
////////////////////////////////////////////////////////////////////////////////

/// lookupPort 는 /etc/services 에서 tbnova/tcp 포트를 조회한다.
///
/// @return  포트 번호
/// @return  조회 실패 시 에러
func lookupPort() (int, error) {
	port, err := net.LookupPort("tcp", serviceName)
	if err != nil {
		return 0, fmt.Errorf(
			"/etc/services 에 %s/tcp 가 없습니다: %w",
			serviceName, err,
		)
	}
	return port, nil
}

/// validateCmd 는 수신된 제어 명령을 검증한다.
///
/// 허용 명령어: status / start / stop / restart / reload / env
///   start/stop/restart : ID 필수
///   status / reload / env : ID 생략 허용 (생략 시 전체 대상)
///
/// @param  cmd  검증할 ControlCmd
/// @return      검증 실패 시 에러
func validateCmd(cmd *model.ControlCmd) error {
	switch cmd.Cmd {
	case "status":
		/// ID 없으면 전체 조회 — 허용
	case "start", "stop", "restart":
		if strings.TrimSpace(cmd.ID) == "" {
			return fmt.Errorf("[%s] 명령은 id 가 필수입니다", cmd.Cmd)
		}
	case "reload":
		/// ID 불필요
	case "env":
		/// ID 생략 시 전역 env, ID 지정 시 해당 프로세스 병합 env — 허용
	default:
		return fmt.Errorf("허용되지 않은 명령어: %q (status/start/stop/restart/reload)",
			cmd.Cmd)
	}
	return nil
}

/// errResp 는 에러 응답을 생성한다.
///
/// @param  msg  에러 메시지
/// @return      ControlResponse
func errResp(msg string) model.ControlResponse {
	return model.ControlResponse{Result: "error", Message: msg}
}

/// truncate 는 문자열을 최대 n 바이트로 자른다.
/// 로그 출력 시 긴 문자열을 안전하게 축약한다.
///
/// @param  s  원본 문자열
/// @param  n  최대 바이트 수
/// @return    축약된 문자열
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
