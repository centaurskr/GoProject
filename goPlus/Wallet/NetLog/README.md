# goWallet (NetLog)

증거금(지갑) 정리 프로세스 — `goAfterMatch`가 재전송하는 체결/취소/주문번호 갱신 데이터(`REALJSON`)를 구독해 사용자 증거금(지갑) 테이블을 정산한다. 로그 백엔드는 [centolib/netlog](../../../centolib/netlog/README.md). 로컬 로그 버전은 [../LocalLog](../LocalLog/README.md).

## 개요

```
goAfterMatch ──REALJSON(PUB)──▶ goWallet(SUB, 전체 구독) ──▶ channel(5000) ──▶ goDataProcess ──▶ DB(wallet)
```

`dataHandler`가 ZMQ SUB로 받은 메시지를 `Message{tr, msg}`로 감싸 버퍼 채널(5000)에 넣고, `goDataProcess`(dbfunc.go) 고루틴이 소비하며 트랜잭션 종류(`tr`)에 따라 매도/매수 체결 정산(`DbSellCheProcess`/`DbBuyCheProcess`), 취소(`DbCancelProcess`), 시장가 완료(`DbMarketDone`) 등으로 분기해 [goPlus/SQL/wallet.sql](../../SQL/README.md) 스키마의 지갑 테이블을 갱신한다.

파일 구성: `wallet.go`(부트스트랩, ZMQ 수신), `dbfunc.go`(증거금 조회/갱신 전체 DB 로직 및 `goDataProcess` 소비 고루틴).

## 빌드

```bash
cd goPlus/Wallet/NetLog
make
make install
```

## 사용법

```bash
goWallet
```

인자 없음. 중복 실행 방지: `goWallet` lock(단일 인스턴스만 허용).

## 환경설정

| 키 | 용도 |
|---|---|
| `PROXYENV_YAML` | `REALJSON`(back) — `goAfterMatch`가 정산 대상 데이터를 재전송하는 채널 |
| `DBCONFIG_YAML` | PostgreSQL 접속 정보(`centolib/api.ConnectDb`) |

## 주의사항

- SUB 소켓을 `SetSubscribe("")`(전체 구독)으로 열어 `REALJSON`의 모든 토픽을 수신한다 — 필터링 없이 전량을 받아 채널에 적재하므로, 트래픽이 몰리면 채널(버퍼 5000)이 밀릴 수 있다.
- 단일 인스턴스로만 실행 가능(`IsRunning("goWallet")`) — `goRecvOrder`/`goMatching`처럼 인덱스로 다중화되어 있지 않다.
