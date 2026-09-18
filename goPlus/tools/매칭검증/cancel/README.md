# cancel

DB의 `ordertbl`에서 `status = 'GO'`(체결 대기 중)인 주문을 전부 조회해, 각각에 대해 취소 주문을 매칭엔진(ORDERPASS 프록시)으로 전송하는 매칭엔진 검증용 도구.

## 개요

1. `DBCONFIG_YAML` 설정으로 DB에 연결.
2. `SELECT ordnum, symbol, itemcode, mkcode, coincode, userid, channel, ordtype, buysell, price FROM ordertbl WHERE status = 'GO';` 로 대기 중 주문을 전부 조회.
3. 조회된 주문 각각에 대해 `SendCancel()`이 `Mb.OrderJson`(채널 `CN03`, `Cancel: "1"`) 취소 패킷을 만들어 `tcp://127.0.0.1:31001`(ORDERPASS 요청 포트, **소스에 하드코딩**)로 REQ/REP 전송.
4. 응답을 콘솔에 출력하고 10ms 대기 후 다음 주문으로 진행.

## 빌드

이 디렉토리에는 Makefile이 없다.

```bash
cd "goPlus/tools/매칭검증/cancel"
go build -o cancel cancel.go
```

## 사용법

```
cancel
```

인자 없이 실행한다. 실행 즉시 DB 조회 및 취소 전송이 시작된다.

## 환경설정

- `DBCONFIG_YAML` (플랫폼 환경 설정, `centolib/api.GetPlatformEnv`로 조회): DB 접속 설정 YAML 경로.
- ORDERPASS 요청 엔드포인트(`tcp://127.0.0.1:31001`)가 **소스코드에 하드코딩**되어 있음 — 다른 호스트/포트 환경에서는 소스를 직접 수정해야 한다.

## 주의사항

- **⚠️ 확인 프롬프트가 없다.** 소스에 `y/n` 확인을 받는 코드가 있으나 주석 처리되어 비활성화된 상태다 — 실행 즉시 `status='GO'`인 **모든** 대기 주문에 취소를 전송한다. 운영 DB를 대상으로 실수로 실행하면 실주문이 전부 취소될 수 있다.
- 취소 요청에 `UserIp`, `Uuid` 등이 고정 값으로 하드코딩되어 있어 실제 요청 출처와 무관한 값이 전송된다 — 검증/테스트 환경 전용 도구이며 운영 트래픽으로 오인되지 않도록 주의.
- 검증(테스트) 목적의 도구이므로 **반드시 테스트 DB/환경에서만 사용**하고 운영 DB를 가리키는 `DBCONFIG_YAML`로 실행하지 않도록 주의해야 한다.
- 개별 주문 전송 실패 시 루프를 즉시 `break`하고 종료한다(이후 주문은 처리되지 않음).
