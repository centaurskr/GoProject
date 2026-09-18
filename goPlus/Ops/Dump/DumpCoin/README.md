# goDumpCoin

지정한 코인의 `MEMCOIN`(호가창 전체 구조)을 mmap 파일에서 로드해 표준출력에 덤프하는 운영 디버깅 도구.

## 개요

[centolib/matchingboard](../../../../centolib/matchingboard/README.md)의 `MEMCOIN.LoadMEMCOIN()` + `MEMCOIN.Dump()`를 그대로 호출하는 얇은 래퍼. 매도/매수 호가 전체 목록과 각 호가의 가격/수량/주문수/mmap 상태를 확인할 때 사용한다.

## 빌드

```bash
cd goPlus/Ops/Dump/DumpCoin
make
make install
```

## 사용법

```bash
goDumpCoin <coin>
```

`<coin>`은 심볼(예: `BTC.USDT`). 환경변수 `ORDPATH`(주문 mmap 파일 루트) 아래 `{coin}/`을 로드한다.

## 환경설정

| 환경변수 | 용도 |
|---|---|
| `ORDPATH` | 주문/호가 mmap 파일 루트 경로 |

## 주의사항

- 대상 코인이 [goMatching](../../../Matching/NetLog/README.md)에 의해 사용 중이어도 읽기 전용으로 attach하므로 운영 중 실행해도 안전하지만, 대량 호가 상황에서는 출력이 매우 길어질 수 있다.
