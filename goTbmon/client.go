// client.go — goTbnova TCP 제어 소켓 클라이언트
//
// 온디맨드(on-demand) 방식으로 동작한다: 각 요청마다 TCP 접속을 새로 열고,
// 응답을 수신한 즉시 연결을 닫는다. 상시 연결을 유지하지 않는다.
//
// 사용 패턴:
//   c := NewClient(addr)
//   resp, err := c.Status("")   // 전체 목록 조회
//   resp, err := c.Start("id")  // 특정 프로세스 시작
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

const (
	defaultAddr    = "127.0.0.1:30001"
	connectTimeout = 3 * time.Second
	rwTimeout      = 5 * time.Second
)

// Client: goTbnova TCP 제어 소켓 접속 정보를 보유하는 클라이언트.
type Client struct {
	addr string
}

// NewClient: 주어진 주소로 Client를 생성한다. addr이 빈 문자열이면 defaultAddr를 사용한다.
func NewClient(addr string) *Client {
	if addr == "" {
		addr = defaultAddr
	}
	return &Client{addr: addr}
}

// send: cmd를 JSON으로 직렬화해 TCP로 전송하고 응답을 역직렬화해 반환한다.
// 접속·송신·수신 각 단계에서 실패하면 즉시 오류를 반환한다.
func (c *Client) send(cmd ControlCmd) (ControlResponse, error) {
	conn, err := net.DialTimeout("tcp", c.addr, connectTimeout)
	if err != nil {
		return ControlResponse{}, fmt.Errorf("연결 실패: %w", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(rwTimeout))

	b, err := json.Marshal(cmd)
	if err != nil {
		return ControlResponse{}, err
	}
	b = append(b, '\n')

	if _, err = conn.Write(b); err != nil {
		return ControlResponse{}, fmt.Errorf("전송 실패: %w", err)
	}

	var resp ControlResponse
	if err = json.NewDecoder(bufio.NewReader(conn)).Decode(&resp); err != nil {
		return ControlResponse{}, fmt.Errorf("응답 파싱 실패: %w", err)
	}
	return resp, nil
}

// Status: id 프로세스(또는 id==""이면 전체)의 상태 스냅샷을 조회한다.
func (c *Client) Status(id string) (ControlResponse, error) {
	return c.send(ControlCmd{Cmd: "status", ID: id})
}

// Start: id 프로세스를 시작한다.
func (c *Client) Start(id string) (ControlResponse, error) {
	return c.send(ControlCmd{Cmd: "start", ID: id})
}

// Stop: id 프로세스를 종료한다.
func (c *Client) Stop(id string) (ControlResponse, error) {
	return c.send(ControlCmd{Cmd: "stop", ID: id})
}

// Restart: id 프로세스를 재시작한다.
func (c *Client) Restart(id string) (ControlResponse, error) {
	return c.send(ControlCmd{Cmd: "restart", ID: id})
}

// Reload: goTbnova에 Process Table 재로드 명령을 보낸다.
func (c *Client) Reload() (ControlResponse, error) {
	return c.send(ControlCmd{Cmd: "reload"})
}

// Env: goTbnova 전역 환경변수 목록을 조회한다.
func (c *Client) Env() (ControlResponse, error) {
	return c.send(ControlCmd{Cmd: "env", ID: ""})
}
