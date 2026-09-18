# goExrate60 (NetLog)

Alpha Vantage API에서 USD/KRW 환율을 주기적으로 조회해 DB에 저장하고, 내부 ZMQ로 환율 변경을 배포하는 프로세스. 로그 백엔드는 [centolib/netlog](../../../centolib/netlog/README.md). 로컬 로그 버전은 [../LocalLog](../LocalLog/README.md).

## 개요

30초마다 tick하여 "60분 단위 구간"(`DATA_MIN=60`)이 바뀌었는지 확인하고, 바뀌었으면 Alpha Vantage `CURRENCY_EXCHANGE_RATE` API(USD→KRW)를 호출해 `exratetbl`(PostgreSQL, upsert)에 저장한 뒤 `EXTRAINFOR` 프록시로 `EXRATE {"CURRENCY":"USD","TIME":..,"RATE":..}` 메시지를 PUB한다. 이 메시지는 [goRecvOrder](../../RecvOrder/NetLog/README.md) 등이 구독해 증거금 계산에 사용한다.

## 빌드

```bash
cd goPlus/Exrate60/NetLog
make
make install
```

## 사용법

```bash
goExrate60
```

인자 없음. 중복 실행 방지: `goExchange60` lock.

## 환경설정

다른 `goPlus` 컴포넌트와 달리 `centolib/api.GetPlatformEnv`(즉 `PLATFORM_ENV` 파일 경유)를 쓰지 않고 **환경변수를 직접 읽는다**:

| 환경변수 | 용도 |
|---|---|
| `PROXY_YAML` | ZMQ 프록시 정보 yaml 경로(`EXTRAINFOR`의 `front` 조회) — 다른 컴포넌트의 `PROXYENV_YAML`과 이름이 다르다 |
| `PLATFORM_ENV` | `DBCONFIG_YAML` 값을 얻기 위한 KEY=VALUE 환경파일 경로 |

소스에 하드코딩된 Alpha Vantage API 키(`apikey=CGKF7AX2H13T4UHN`)를 사용한다.

## 주의사항

- **다른 `goPlus` 컴포넌트와 환경변수 접근 방식이 다르다** — 대부분은 `PLATFORM_ENV` 파일을 경유하는 `GetPlatformEnv("PROXYENV_YAML")`을 쓰지만, 이 프로세스는 `os.Getenv("PROXY_YAML")`로 OS 환경변수를 직접 읽는다. 배포 시 이 차이를 놓치면 프록시 정보를 못 찾아 기동에 실패한다.
- **API 키가 소스에 하드코딩**되어 있다 — 키 교체 시 재빌드가 필요하고, 이 저장소를 공유하면 키가 함께 노출된다.
- Alpha Vantage 무료 API는 호출량 제한이 있다 — 30초 tick + 60분 단위 게이트로 실제 API 호출은 시간당 최대 1회 수준으로 제한되어 있지만, API 응답이 지연/실패해도 재시도 로직 없이 다음 구간까지 대기한다.
