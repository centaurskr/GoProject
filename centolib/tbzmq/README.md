# centolib/tbzmq

ZeroMQ 소켓 모니터링 및 PUB/SUB 프록시 구동을 위한 얇은 래퍼 패키지.

## Import

```go
import "centolib/tbzmq"
```

## 개요

`github.com/pebbe/zmq4` 위에 두 가지 헬퍼를 제공한다: 임의 소켓의 연결/해제 이벤트를 감시하는 모니터 소켓 생성, 그리고 XSUB/XPUB 기반 PUB/SUB 프록시를 백그라운드 고루틴으로 구동하는 함수.

## API

### `func MakeMonitorSocket(base string, sock *zmq4.Socket) (*zmq4.Socket, error)`
- 설명: `sock`에 대해 `zmq4.EVENT_ACCEPTED | zmq4.EVENT_DISCONNECTED` 이벤트를 발행하는 `inproc://` PAIR 모니터 채널을 만들고, 그 채널에 연결된 별도의 PAIR 소켓을 반환한다.
- 파라미터: `base` — inproc 채널 이름 접두사(빈 문자열이면 나노초 타임스탬프로 자동 생성해 고유성 보장), `sock` — 모니터링 대상 소켓.
- 반환값: 이벤트를 수신할 PAIR 소켓, 에러.

### `func RunPubSubProxy(ctx *zmq4.Context, f, b string) error`
- 설명: `f`(front, XSUB로 bind)와 `b`(back, XPUB로 bind) 두 엔드포인트 사이에서 `zmq4.Proxy`를 백그라운드 고루틴으로 실행하는 PUB/SUB 중계 프록시를 구동한다. `goPlus/Proxy/*/PubSubProxy`가 사용하는 패턴과 동일한 역할.
- 파라미터: `f` — front 바인드 주소, `b` — back 바인드 주소.
- 반환값: 소켓 생성/bind 실패 시 에러(함수 자체는 프록시 고루틴을 기동만 하고 즉시 반환하므로, 프록시 동작 중 에러는 감지하지 못한다).

## 의존성

`github.com/pebbe/zmq4`.

## 주의사항

- `RunPubSubProxy`는 프록시를 고루틴으로 띄우고 바로 리턴한다 — 호출자가 프로세스를 종료시키지 않도록 별도의 대기(블로킹) 로직을 가지고 있어야 한다.
- `MakeMonitorSocket`이 만드는 `inproc://` 채널은 프로세스 내부 전용이며, 모니터 대상 소켓과 같은 `zmq.Context`를 공유해야 한다(내부적으로 `sock.Context()`로 가져와 사용).
