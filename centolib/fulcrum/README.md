# centolib/fulcrum

Fulcrum(Electrum 프로토콜 서버)에 TCP JSON-RPC로 접속해 Bitcoin 주소 잔액을 조회하는 패키지.

## Import

```go
import "centolib/fulcrum"
```

## 개요

Bech32(SegWit, `bc1...`) 주소를 Electrum scripthash로 변환한 뒤 Fulcrum 서버(기본 포트 50001)에 `blockchain.scripthash.get_balance` JSON-RPC를 호출한다. `getbal`, `goapisvr`가 이 패키지를 사용한다.

## API

### `func GetBalance(address, fulcrumAddress string) (string, error)`
- 설명: `address`(Bech32 형식)를 scripthash로 변환하고, `fulcrumAddress`(`"host:port"`)에 TCP로 접속해 잔액 조회 요청을 보낸 뒤 응답을 받아 반환한다.
- 파라미터: `address` — 조회할 Bitcoin Bech32 주소, `fulcrumAddress` — Fulcrum 서버 주소(예: `"192.168.0.102:50001"`).
- 반환값: `{"confirmed":..,"unconfirmed":..}` 형태의 JSON 문자열, 에러.
- 사용 예 (`getbal/getbal.go` 실사용):
  ```go
  rtnstr, err := fulcrum.GetBalance(
      "bc1qmnnq9vv4nhd35px2j7dkks5smvgu96a5sakf3v",
      "192.168.0.102:50001")
  ```

내부 비공개 함수 `decodeBech32Address`/`generateScripthash`는 Bech32 디코드 → witness program 추출 → `OP_0 <program>` 스크립트 SHA256 → 바이트 역순(Electrum 프로토콜 규약)으로 scripthash를 생성한다.

## 의존성

`github.com/btcsuite/btcutil/bech32`.

## 주의사항

- Bech32(`bc1...`, P2WPKH/P2WSH) 주소만 지원한다. HRP(human-readable part)가 `"bc"`가 아니면 에러 — 즉 mainnet 주소만, Legacy(`1...`)/P2SH(`3...`) 주소나 testnet은 지원하지 않는다.
- 매 호출마다 새 TCP 연결을 맺고 요청 1개를 보낸 뒤 응답 1줄(`\n` 구분)을 읽고 닫는다 — 연결 재사용이나 타임아웃 설정이 없으므로, Fulcrum 서버가 응답하지 않으면 호출이 블로킹될 수 있다.
