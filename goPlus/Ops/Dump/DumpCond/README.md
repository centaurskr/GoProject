# goDumpCond

코인별 주문 가능 상태(`OrderCond.map`)를 덤프하는 운영 디버깅 도구.

## 개요

`ORDERCOND_MAP` mmap 파일(헤더 + `OrderCondition` 배열)을 attach해 코인별 그룹/상태(1=정상, 0=주문정지, 3=매도만 가능)를 표 형태로 출력한다. 이 mmap 파일은 [goMatchingCtl](../../../MatchingCtl/NetLog/README.md)이 갱신하고 [goRecvOrder](../../../RecvOrder/NetLog/README.md)가 조회하는 공유 상태다.

## 빌드

```bash
cd goPlus/Ops/Dump/DumpCond
make
make install
```

## 사용법

```bash
goDumpCond
```

인자 없음.

## 환경설정

| 키 | 용도 |
|---|---|
| `ORDERCOND_MAP` | `centolib/api.GetPlatformEnv`로 조회하는 주문 상태 mmap 파일 경로 |

## 주의사항

- 같은 디렉토리에 `dumpcoin.go`, `dumpord.go`라는 **빈 파일(0바이트)**이 남아있다 — `Makefile`은 `dumpcond.go`만 빌드 대상으로 삼으므로 실제 빌드에는 영향이 없지만, 다른 목적으로 만들다 만 잔재로 보이며 실제 내용이 없다.
