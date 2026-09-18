# goTbnova/control

TCP 모니터 서버 — 레거시 C버전 monitor client와 호환되는 JSON 제어 프로토콜.

## Import

```go
import "goTbnova/control"
```

## 개요

`/etc/services`의 `tbnova/tcp` 포트로 TCP 연결을 수락하고, 줄바꿈으로 구분된 JSON 명령(`model.ControlCmd`)을 수신해 [process.Manager](../process/README.md)의 `controlCh` 채널로 전달한 뒤 응답(`model.ControlResponse`)을 JSON으로 돌려준다. `goTbmon`이 이 프로토콜의 클라이언트다.

설계 원칙: C버전과 포트를 공유해(중복 기동 시 bind 실패로 방지) 동시 실행을 막고, `127.0.0.1`에만 바인딩(로컬 전용)하며, 연결당 goroutine 1개로 처리하고, 실제 상태 변경은 항상 `manager.Run()` goroutine에서만 수행하도록 요청을 채널로 위임한다.

## API

### `func NewServer(controlCh chan<- process.ControlRequest, bindHost string) (*Server, error)`
- 설명: `/etc/services`에서 `tbnova/tcp` 포트를 조회해 `bindHost:port`에 TCP 리슨한다.
- 파라미터: `controlCh` — manager의 제어 요청 채널(`Manager.ControlCh()`), `bindHost` — 바인딩 IP(`""`이면 `127.0.0.1`, `"0.0.0.0"`이면 모든 인터페이스).
- 반환값: `*Server`, 에러. **포트 바인딩 실패 시 "C버전 또는 다른 goTbnova가 이미 실행 중"이라는 에러 메시지**를 명시적으로 반환한다.

### `func (s *Server) Serve(ctx context.Context)`
- 설명: `Accept()` 루프. `ctx.Done()` 시 내부적으로 리스너를 닫아 루프를 정상 종료한다. `main()`에서 `go srv.Serve(ctx)`로 기동.

### `func (s *Server) Port() int`
- 설명: 실제 바인딩된 포트 번호 반환(`tlog.Startup()` 로깅용).

## 프로토콜

- Request(client→server): `model.ControlCmd{Cmd, ID}` JSON 1줄(`\n` 종료), 최대 4KB.
- Response(server→client): `model.ControlResponse{Result, Message, Processes, Env}` JSON 1줄.
- 허용 명령(`Cmd`): `status`(ID 생략 시 전체), `start`/`stop`/`restart`(ID 필수), `reload`(ID 불필요, 비동기 접수), `env`(ID 생략 시 전역 env, 지정 시 병합 env).

## 환경설정

- 조회 포트: `/etc/services`의 `tbnova/tcp` 항목이 **반드시 등록되어 있어야** `NewServer`가 성공한다.
- 바인딩 IP는 [goTbnova root README](../README.md)의 `gotbnova.toml [server] bind`에서 온다.

## 주의사항

- **보안**: 기본 바인딩은 `127.0.0.1`(로컬 전용)이다. `bind`를 `0.0.0.0`이나 특정 외부 IP로 바꾸면 인증 없는 프로세스 제어 채널이 그대로 노출되므로 신뢰할 수 있는 네트워크에서만 사용해야 한다(코드에도 `[WARN]` 로그가 남는다).
- 요청 크기 4KB, 읽기 타임아웃 30초, manager 응답 대기 타임아웃 10초로 각각 제한되어 있다 — 이 한도를 넘는 요청/응답은 에러로 처리되고 연결이 끊긴다.
- `handleConn`은 각 연결마다 `respCh := make(chan model.ControlResponse, 1)`을 새로 만들어 manager에 전달한다 — manager 쪽에서 이 채널에 정확히 1번 응답을 써야 하며(현재 `manager.handleControl`이 그렇게 동작), 응답을 안 쓰면 `respTimeout`(10초)으로 클라이언트가 타임아웃 에러를 받는다.
