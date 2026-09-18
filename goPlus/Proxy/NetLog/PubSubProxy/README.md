# goPubSubProxy (NetLog)

ZeroMQ PUB/SUB 메시지 중계 프록시 — `goPlus` 시스템 내 여러 컴포넌트가 서로 직접 연결하지 않고 이 프록시를 거쳐 pub/sub 하도록 하는 범용 릴레이. 로그 백엔드는 [centolib/netlog](../../../../centolib/netlog/README.md). 로컬 로그 버전은 [../../LocalLog/PubSubProxy](../../LocalLog/PubSubProxy/README.md).

## 개요

```
PUB(publisher들) ──▶ [front: SUB, bind] goPubSubProxy [back: PUB, bind] ──▶ SUB(구독자들)
                                            ⇅ control(REP)
```

`front`(SUB, publisher들이 연결)로 받은 메시지를 그대로 `back`(PUB, 구독자들이 연결)으로 재전송한다. `EXTRAINFOR`, `BEFOREMATCHING`, `AFTERMATCHING`, `SISEJSON`, `HOGAJSON` 등 `goPlus` 전역의 pub/sub 채널이 모두 이 프로그램의 인스턴스다(같은 바이너리를 서버명만 바꿔 여러 번 기동). Capture 기능은 없다(필요 시 `back` 엔드포인트에서 직접 구독).

## 빌드

```bash
cd goPlus/Proxy/NetLog/PubSubProxy
make
make install
```

## 사용법

```bash
goPubSubProxy <ServerName>
```

`<ServerName>`은 `$ENV_HOME/Proxyenv.yaml`에 정의된 프록시 이름(`centolib/api.ProxyYaml`의 최상위 키). 해당 항목의 `proxytype`이 정확히 `"PUB/SUB"`이어야 하며 아니면 즉시 에러 종료.

## 제어 프로토콜 (REQ/REP, `control` 엔드포인트)

| 명령 | 동작 |
|---|---|
| `INFOR` | 현재 `ProxyInfor`(연결 수, 읽기/쓰기 카운트 등) JSON 응답 |
| `PAUSE` | 메시지 릴레이 일시 중단(수신은 계속하지만 전달 루프에서 `sleep` 후 skip) |
| `RUN` | 릴레이 재개 |

## 환경설정

- `$ENV_HOME/Proxyenv.yaml` — 이 프록시의 `front`/`back`/`control` 엔드포인트가 정의된 YAML. `ENV_HOME` 환경변수 필수.

## 주의사항

- `front`가 `SUB`로 bind되어 있다(코드 주석엔 `XSUB` 고려 흔적이 있으나 실제로는 표준 `SUB` 사용) — publisher 수가 많아지는 fan-in 구조에는 XSUB/XPUB 프록시가 더 적합할 수 있음을 유의.
- `back` 구독자가 0명(`BackCount==0`)이면 메시지를 읽기는 하되 전달은 건너뛴다(`ReadCount`만 증가) — 구독자가 붙기 전 발행된 메시지는 유실된다(버퍼링 없음).
- `PAUSE` 상태에서는 poller 자체를 돌지 않고 500ms 간격으로 sleep만 한다 — front 소켓에 쌓이는 메시지는 ZMQ 내부 HWM(High Water Mark)까지만 버퍼링된다.
