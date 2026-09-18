# goNewCoin

신규 코인 상장 시 주문 관련 초기 mmap 파일 구조를 생성하는 운영 도구.

## 개요

다음을 순서대로 수행한다:
1. DB(`coininfor` 테이블)에서 코인 정보 조회(`Mb.GetCoinInfor`)
2. `$ORDPATH/{coin}` 디렉토리 생성(없으면)
3. `{coin}.MASTER` 주문 마스터 mmap 파일 생성(`Mb.CreateOrderMasterInfor`)
4. 매도(`MEDO.00~03`)/매수(`MESU.00~03`) 호가 mmap 파일 생성(`Mb.CreateHoga`, [centolib/matchingboard](../../../../centolib/matchingboard/README.md)의 `ORDER_COUNT`개만큼)

[goMatchingCtl](../../../MatchingCtl/NetLog/README.md)의 신규상장 처리(`newListProcess`)가 호출하는 `MakeNewCoinOrder` 로직과 동등한 작업을 수동으로 수행하는 CLI 버전이다.

## 빌드

```bash
cd goPlus/Ops/NewCoin
make
make install
```

## 사용법

```bash
goNewCoin <coin>
```

`<coin>`: 거래 코인 이름(예: `BTC.KRW`). 로그는 [centolib/logout](../../../../centolib/logout/README.md) `ALL_MODE`(콘솔+파일 동시)로 고정.

## 환경설정

| 환경변수 | 용도 |
|---|---|
| `PLATFORM_ENV` | KEY=VALUE 환경파일 경로. 이 안에서 `ORDPATH`, `DBCONFIG_YAML`을 조회(`centolib/api.GetEnvValue`) |

## 종료 코드

인자 오류(1), env 미설정(1), `ORDPATH`/`DBCONFIG_YAML` 없음(2/3), DB 접속 실패(4), DB에 코인 정보 없음(5), 디렉토리 생성 실패(6), 마스터/호가 파일 생성 실패(7~10).

## 주의사항

- **이미 존재하는 코인에 대해 재실행하면 실패한다**(`CreateOrderMasterInfor`가 기존 `.MASTER` 파일이 있으면 에러 반환) — 재초기화가 필요하면 기존 mmap 파일을 먼저 정리해야 한다.
- DB에 해당 코인 정보(`coininfor` 테이블)가 먼저 등록되어 있어야 한다.
