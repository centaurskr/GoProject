# goAfterMatch (NetLog)

매칭 후 DB 정리 프로세스 — `goMatching`이 만든 체결/주문 이벤트를 받아 DB(`ordertbl`/`tradetbl` 등)에 반영하고, 정리된 데이터를 시세/호가 배포용으로 재전송한다. 로그 백엔드는 [centolib/netlog](../../../centolib/netlog/README.md). 로컬 로그 버전은 [../LocalLog](../LocalLog/README.md).

## 개요

```
【AFTERMATCHING Proxy】 🡆 (DBERR/DBCHE/DBSEQ/DBCANCEL/MKDONE) 🡆 goAfterMatch
                                                                    🡇 channel(size 10000, CircularQueue)
                                                              goDataProcess (dataproc.go)
                                                                    🡇
                                                       DB 반영(dbfunc.go) + REALJSON/SISEJSON 재전송
```

`goMatchingCtl`에 raw TCP로 접속해 자신의 인덱스를 알리고, 처리할 구독 토픽(`AFTER{idx}`)을 동적으로 할당받는다(N-Value 변경 시 [goMatchingCtl](../../MatchingCtl/NetLog/README.md)이 재할당). 데이터 수신(`DataHandler`)과 실제 DB 처리(`goDataProcess` 고루틴)를 [CircularQueue](spscqueue.go)로 분리해 ZMQ 수신 루프가 DB I/O에 막히지 않도록 한다.

파일 구성: `afterMatch.go`(부트스트랩, TCP/ZMQ 이벤트 핸들러), `dataproc.go`(큐 소비 고루틴, 수신 타입별 분기), `dbfunc.go`(PostgreSQL 갱신 함수 — 주문 오류/취소/시장가완료/주문번호/체결 반영), `jsonfunc.go`(`InnerChePacket → JsonChePacket` 변환), `spscqueue.go`(단일 생산자/단일 소비자 원형 큐).

## 빌드

```bash
cd goPlus/AfterMatch/NetLog
make
make install
```

## 사용법

```bash
goAfterMatch <Index>
```

`<Index>`는 이 인스턴스의 식별 번호(`goMatchingCtl`에 `AFTERMATCH:{Index}`로 등록되며, 로그 파일명 접미사로도 쓰임). 중복 실행 방지는 `AfterMatch_{Index}` lock으로 수행.

## 환경설정

| 키 | 용도 |
|---|---|
| `PROXYENV_YAML` | `AFTERMATCHING`(체결 데이터 수신, back), `REALJSON`(처리결과 재전송, front), `EXTRAINFOR`(CONTROL 채널) |
| `MATCHING_HOST`, `AFTERMAT_STREAM` | `goMatchingCtl` TCP 접속 주소 |
| `DBCONFIG_YAML` | [centolib/api.ConnectDb](../../../centolib/api/README.md)에 전달되는 PostgreSQL 접속 정보 yaml |

## 주의사항

- **DB 연결(`DBCONFIG_YAML`)이 필수**다 — 접속 실패 시 즉시 종료한다.
- 수신 큐(`CircularQueue`, 크기 10000)가 가득 차면 `DataHandler`가 **10μs 간격으로 재시도하며 블로킹**한다(드롭하지 않음) — DB 처리가 지속적으로 밀리면 ZMQ 수신도 함께 지연된다.
- 구독 토픽은 `goMatchingCtl`이 동적으로 지정하며, 재할당 시 기존 토픽을 `Unsubscribe`하고 새 토픽을 `Subscribe`한다 — `goMatchingCtl` 재기동 시 순간적으로 데이터 수신이 끊길 수 있다.
- `<Index>`는 인자 필수이며 누락 시 사용법만 출력하고 종료한다.
