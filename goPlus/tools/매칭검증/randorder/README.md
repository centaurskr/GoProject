# randorder (genord)

`coins.json`에 정의된 코인 정보를 바탕으로 무작위 주문 5,000건을 생성해두고, 그중 무작위로 골라 최대 10,000회까지 REQ 소켓으로 매칭엔진에 주문을 전송하는 부하/검증용 도구.

## 개요

이 디렉토리에는 `package main`을 가진 독립 파일이 3개 있으나, **실제로 사용하는 것은 `genord.go` 하나뿐**이다.

- `genord.go` — 실행 대상. `coins.json`을 읽어(`readCoin`) 코인 목록을 만들고, `makeOrder(5000)`으로 무작위 주문 5,000건을 미리 생성한 뒤, 최대 10,000회 반복하며 무작위로 하나를 골라 ZMQ REQ로 매칭엔진에 전송하고 응답(`RtnJson.State`)을 확인한다. 응답 상태가 `"0"`이 아니면 루프를 중단한다.
- `readcoin.go` — `genord.go`와 별개의 `readCoin()`/`Coin` 타입을 가진 예전/실험 버전으로 보인다. **`genord.go`와 함께 빌드하면 `package main` 중복 선언 충돌이 난다** — 사용하지 않는 잔재 파일로 추정.
- `randprc2.go` — 가격 변동 실험용 독립 스니펫(`getPrice`/`getDir`), `genord.go`와 별개의 `main()`을 가짐. 마찬가지로 함께 빌드 불가.

## 빌드 / 실행

Makefile 없음. `run.sh`가 곧 실행 방법이다:

```bash
cd "goPlus/tools/매칭검증/randorder"
cat run.sh
# go run genord.go tcp://127.0.0.1:31001

./run.sh
# 또는 직접:
go run genord.go tcp://127.0.0.1:31001
```

`readcoin.go`, `randprc2.go`를 실행하려면 각각 별도 디렉토리로 옮기거나 `go run readcoin.go` / `go run randprc2.go`처럼 단일 파일만 지정해서 빌드해야 한다(같은 디렉토리에서 `go build .`나 `go run *.go`를 하면 실패한다).

## 사용법

```
genord <REQ endpoint>
```

- `<REQ endpoint>`: 매칭엔진(RecvOrder 등)의 주문 접수 REQ 엔드포인트 (예: `tcp://127.0.0.1:31001`)

## 환경설정

- 실행 디렉토리에 `coins.json` 파일이 있어야 한다(상대경로 `"coins.json"`로 오픈). 심볼/종목코드/시장코드/기준가/기준수량을 담은 JSON 배열.
- `coins.json.8`이라는 백업/샘플 파일도 같이 있다(직접 사용되지 않음).

## 주의사항

- **부하성 도구다**: 최대 10,000건의 주문을 REQ/REP로 연속 전송한다(호출 사이 지연 코드는 모두 주석 처리되어 비활성화됨). 운영 엔드포인트에 실수로 실행하면 실제 주문이 대량 유입될 수 있다 — 반드시 테스트 환경 엔드포인트로만 실행할 것.
- 가격/수량은 기준값의 ±0.1%(가격) / ±50%(수량) 범위에서 무작위로 흔든다(`getPrice`/`getQty`).
- 사용자 ID는 `UID_BASE(42280)` ~ `+9` 범위에서만 무작위 선택된다 — 고정된 소수의 테스트 계정을 전제로 한다.
- `readcoin.go`/`randprc2.go`는 미사용 잔재 파일로 보이며, 정리 대상인지 확인이 필요하다.
