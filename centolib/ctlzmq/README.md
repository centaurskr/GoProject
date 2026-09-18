# centolib/ctlzmq

ZeroMQ PUB/SUB CONTROL 프록시를 통해 프로세스 간 제어 메시지를 주고받는 패키지.

## Import

```go
import "centolib/ctlzmq"
```

## 개요

`goPlus` 계열 데몬(Matching, Proxy 등)이 서로에게 제어 명령(재시작, PAUSE 등으로 추정)을 보내기 위한 메시징 계층이다. 메시지는 중앙 CONTROL 프록시(PUB/SUB 릴레이, `centolib/api/proxyenv.go`의 `ProxyInfor` yaml에 `EXTRAINFOR`라는 이름으로 등록되어 있다고 가정)를 경유한다.

```
"CONTROL" + FROM + TO + message  (PUB) →  [CONTROL PROXY]  → (SUB)
```

## API

### `func MakeControlSocket(ctx *zmq.Context, pubPoint, subPoint, name string) (*zmq.Socket, error)`
- 설명: PUB 소켓을 `pubPoint`에 연결해 패키지 전역 `CtlPubSocket`/`FromName`을 설정하고, SUB 소켓을 `subPoint`에 연결한 뒤 토픽 `"CONTROL"`을 구독해 반환한다. 연결 안정화를 위해 100ms 대기한다.
- 파라미터: `pubPoint`/`subPoint` — 프록시의 front/back 엔드포인트, `name` — 이 프로세스의 식별 이름(FROM).
- 반환값: 구독용 SUB 소켓, 에러.

### `func InitControlMessage(ctx *zmq.Context, name string) (*zmq.Socket, error)`
- 설명: 환경변수 `PLATFORM_ENV`가 가리키는 yaml에서(`centolib/api.GetPlatformEnv("PROXYENV_YAML")`) `EXTRAINFOR` 항목의 `front`/`back` 엔드포인트를 읽어 `MakeControlSocket`을 호출하는 고수준 초기화 함수.
- 파라미터: `name` — FROM 이름.
- 반환값: 구독용 SUB 소켓, 에러.

### `func SendControlMessage(toName, cmd string)`
- 설명: `"CONTROL"`, `FromName`, `toName`, `cmd` 4개 프레임을 `CtlPubSocket`으로 전송한다. `MakeControlSocket`/`InitControlMessage`로 먼저 초기화해야 한다(하지 않으면 `nil` 소켓 사용으로 패닉).

### `func CloseControlMessage()`
- 설명: `CtlPubSocket.Close()`.

## 의존성

`github.com/pebbe/zmq4`, `centolib/api`(환경변수/yaml 조회).

## 주의사항

- `CtlPubSocket`/`FromName`은 패키지 전역 변수다 — 프로세스당 하나의 컨트롤 채널만 가정하며, 동시에 여러 컨트롤 컨텍스트를 다루는 용도로는 쓸 수 없다.
- `InitControlMessage`를 쓰려면 `PLATFORM_ENV` 환경변수와 그 파일 안의 `PROXYENV_YAML` 키, 그리고 그 yaml 안의 `EXTRAINFOR` 프록시 항목이 모두 사전에 구성되어 있어야 한다.
