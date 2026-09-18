# goDumpOrdmap

임의의 주문 mmap 파일(`MEDO.*`/`MESU.*`, `OrderHeader`+`Order[]`) 하나를 직접 지정해 덤프하는 저수준 디버깅 도구.

## 개요

[goDumpCoin](../DumpCoin/README.md)이 코인 전체(`MEMCOIN`)를 로드하는 것과 달리, 이 도구는 **개별 mmap 파일 1개**를 경로로 직접 열어 헤더와 모든 주문 슬롯을 순서대로 덤프한다(`centolib/matchingboard`의 `OrderHeader.Dump()`/`Order.Dump()` 사용).

## 빌드

```bash
cd goPlus/Ops/Dump/DumpOrdMap
make
make install
```

## 사용법

```bash
goDumpOrdmap <mapname> <mode>
```

- `<mapname>`: mmap 파일 전체 경로(예: `$ORDPATH/BTC.USDT/MEDO.0000`)
- `<mode>`: `1`(전체 덤프) / `2`(JSON 원본 페이로드는 덤프하지 않음) — 단, 실제 `Order.Dump()`/`OrderHeader.Dump()` 구현은 `mode` 값을 사용하지 않고 항상 전체를 출력한다(코드상 `mode`는 출력 헤더 문자열에만 표시됨).

## 환경설정

없음(파일 경로를 인자로 직접 받음).

## 주의사항

- `<mode>` 인자가 사용법 설명과 달리 **실제로는 출력 내용에 영향을 주지 않는다**(항상 전체 덤프) — `mode=2`를 줘도 JSON 페이로드가 함께 출력된다.
