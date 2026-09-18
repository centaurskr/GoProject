# goReqRepProxy (NetLog)

ZeroMQ REQ/REP 메시지 중계 프록시(ROUTER/DEALER) — 다수의 REQ 클라이언트를 소수의 REP 서버로 라우팅하며, 모든 요청/응답을 별도 PUB 채널로 캡처(복제 방송)할 수 있다. 로그 백엔드는 [centolib/netlog](../../../../centolib/netlog/README.md). 로컬 로그 버전은 [../../LocalLog/ReqRepProxy](../../LocalLog/ReqRepProxy/README.md).

## 개요

```
REQ(클라이언트들) ──▶ [ROUTER, bind] goReqRepProxy [DEALER, bind] ──▶ REP(서버)
                            ⇅ control(REP)         │
                     capture(PUB) ◀── inner PUB/SUB 프록시로 REQ/REP 메시지 복제
```

`router`(ROUTER)로 받은 요청을 `dealer`(DEALER)로 그대로 전달하고, `dealer` 응답을 `router`로 돌려준다(표준 ZMQ ROUTER/DEALER 프록시 패턴). 동시에 요청/응답 메시지 본문을 in-process PUB/SUB 프록시를 거쳐 `capture` 엔드포인트(PUB)로 복제 방송한다 — 외부에서 트래픽을 관찰(디버깅/감사용)할 수 있다. [goRecvOrder](../../../RecvOrder/NetLog/README.md)의 `ORDERPASS` 채널이 이 프록시의 대표적 사용처다.

## 빌드

```bash
cd goPlus/Proxy/NetLog/ReqRepProxy
make
make install
```

## 사용법

```bash
goReqRepProxy <ServerName>
```

`<ServerName>`은 `$ENV_HOME/Proxyenv.yaml`에 정의된 프록시 이름. `proxytype`이 정확히 `"REQ/REP"`이어야 한다.

## 제어 프로토콜 (REQ/REP, `control` 엔드포인트)

| 명령 | 동작 |
|---|---|
| `INFOR` | 현재 `ProxyInfor` JSON 응답 |
| `PAUSE` | 릴레이 일시 중단(`"PAUSE OK"` 응답) — 예: [goRecvOrder](../../../RecvOrder/NetLog/README.md)/[goMatchingCtl](../../../MatchingCtl/NetLog/README.md)이 정비 중 주문 접수를 막을 때 이 명령을 쓴다 |
| `RUN` | 릴레이 재개(`"RUN OK"` 응답) |

## 환경설정

- `$ENV_HOME/Proxyenv.yaml` — `front`(ROUTER 바인드), `back`(DEALER 바인드), `control`, `capture`(PUB 바인드) 엔드포인트 정의.

## 주의사항

- `PubSubProxy`와 달리 이 프록시는 `ProxyInfor.CaptureEndPoint`(캡처 채널)를 실제로 사용한다 — [centolib/api/proxyenv.go](../../../../centolib/api/README.md)의 주석에서 "PUB/SUB Proxy의 Capture는 REQ/REP인 경우만"이라 명시한 것과 일치.
- Capture PUB은 **구독자가 없으면(`CaptureCount<=0`) 메시지를 보내지 않는다**(연결 여부만으로 판단, `Pause`와는 무관) — 캡처는 항상 켜져 있는 게 아니라 구독자가 붙어야 데이터가 흐른다.
- `PAUSE` 상태에서는 poller를 돌지 않고 1초 간격으로 sleep한다 — REQ 클라이언트는 프록시가 응답하지 않으므로 자체 타임아웃에 걸릴 수 있다.
