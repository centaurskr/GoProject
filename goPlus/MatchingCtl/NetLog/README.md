# goMatchingCtl (NetLog)

매칭 시스템 제어 프로세스 — 여러 `goMatching`/`goAfterMatch` 인스턴스를 TCP로 관리하고, 코인별 주문 가능 상태(`OrderCond.map`)와 신규 상장을 조율하는 중앙 컨트롤러. 로그 백엔드는 [centolib/netlog](../../../centolib/netlog/README.md). 로컬 로그 버전은 [../LocalLog](../LocalLog/README.md).

## 개요

```
                      
 "AFTER"+idx        N-Number       "ORDCTL"
【goAfterMatch】       ↑            "CONTROL"(SUB)
            ↘  ┌────────────────────┐       ┌─────────────┐
tcp-connect    │  🔴 goMatchingCtl  │  ⇆    │  EXTRAINFOR │  (ZMQ CONTROL 프록시)
            ↗  └───────────┬────────┘       └─────────────┘
【goMatching】             │ "CONTROL" + "MATCHINGCTL"(PUB)
"MATCHING"+idx+Symbol      ↓
                    OrderCond.map (코인별 그룹/상태)
```

- **N-Value**: 현재 실행 중인 `goAfterMatch` 개수. `goMatching`↔`goAfterMatch` 간 토픽 분산에 쓰이며(`AFTER{UserId % N}`), 같은 유저의 주문/체결은 항상 같은 `goAfterMatch`가 처리하도록 보장해 데이터 순서 역전을 방지한다.
- `goMatching`/`goAfterMatch`는 기동 시 이 프로세스에 raw TCP로 접속해 그룹/인덱스를 알리고, 이 프로세스가 N-Value·토픽 할당을 계산해 응답한다.
- 신규 상장(`ORDCTL: NEWLIST:{symbol}`) 시: 주문정보 생성 → `ORDER_LIST`에 코인 추가 → `ORDERPASS` 프록시 PAUSE → 코인 목록 재로드 → `ORDERPASS` RESUME → `goRecvOrder`(REATTACH)/`goMatching`(RELOAD)/`goAfterMatch`(RECONNECT)에 CONTROL 메시지 전송.

## 빌드

```bash
cd goPlus/MatchingCtl/NetLog
make            # ./goMatchingCtl 생성
make install    # 빌드 + ~/bin 으로 이동
```

## 사용법

```bash
goMatchingCtl
```

인자 없음. 중복 실행 시(`api.IsRunning("MATCHINGCTL")`) 즉시 종료된다.

## 환경설정

| 키 | 용도 |
|---|---|
| `PROXYENV_YAML` | `ORDERPASS`(주문 흐름 PAUSE/RUN 제어), `EXTRAINFOR`(CONTROL 채널) 프록시 정보 |
| `MATCHING_STREAM` | `goMatching` 클라이언트를 받는 raw TCP 리슨 주소 |
| `AFTERMAT_STREAM` | `goAfterMatch` 클라이언트를 받는 raw TCP 리슨 주소 |
| `ORDER_LIST` | 신규 상장 시 코인을 추가하는 코인 목록 파일 |

## 주의사항

- **매칭 시스템의 중앙 조정자**다 — 이 프로세스가 죽으면 `goMatching`은 접속 끊김을 감지해 `ORDERPASS`에 PAUSE를 보내 주문 수신을 자동 중단시킨다([../../Matching/README](../../Matching/NetLog/README.md) 참고).
- `goAfterMatch`가 전부 종료되면(N-Value=0) 자동으로 `ORDERPASS`에 PAUSE를 보낸다 — 체결 후처리 인스턴스가 하나도 없는 상태로 주문이 계속 들어오는 것을 방지하기 위함.
- 단일 인스턴스로만 실행되어야 한다(`IsRunning` 락).
