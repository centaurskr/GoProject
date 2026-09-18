# goRecvOrder (NetLog)

주문 수신 프로세스 — 외부에서 들어오는 주문(REQ/REP)을 검증·접수번호 발급·증거금 확인·DB 저장한 뒤, 매칭엔진으로 전달(PUB)하는 거래소 입구 컴포넌트. 로그 백엔드는 [centolib/netlog](../../../centolib/netlog/README.md). 로컬 로그 버전은 [../LocalLog](../LocalLog/README.md).

## 개요

```
┌──────────┐  ①    ┌────────────────┐
│ OrdBrg   │  ⇄    │ REQ/REP Proxy  │
└──────────┘  ✓    │ (goReqRepProxy)│       ┌────────────────┐
         접수번호  └────────────────┘    ┌─ │  wallet using  │ 증거금 Locking
                           ⑤ ⇅ ②          │  └────────────────┘
               ⑥    ┌────────────────┐ ③  │
   ORDER_CALL   🡄   │ 🔴goRecvOrder  │ ⇄  ┤ (매칭 대상 코인마다 wallet 조회)
  (주문 Insert)     └────────────────┘ ④  │
                           ⑦ ↓            │
                    ┌────────────────┐    └─
                    │  PUB/SUB Proxy │
                    │ (goPubSubProxy)│ ──▶ goMatching(BEFOREMATCHING)
                    └────────────────┘
```

처리 순서(`ReqHandler`): ① JSON 파싱 → ② 접수번호 채번(mmap `REGNUM_MAP`, 인덱스별 슬롯) → ③ `ValidateOrder`(정합성 검증, 환율/증거금 계산) → ④ `DbCheckMargin`(증거금 부족 확인) → ⑤ `DbInsertOrder`(ORDER_CALL 테이블 INSERT) → ⑥ REQ 응답 전송 → ⑦ 매칭엔진으로 내부 PUB(`BEFOREMATCHING`).

접수번호 체계: `uint64`, `yymmdd+00+00000000000`(11자리, index 2자리 + 시퀀스). 주문 가능 여부는 mmap 공유 파일(`ORDERCOND_MAP`, [goMatchingCtl](../../MatchingCtl/NetLog/README.md)이 갱신)을 `centolib/api.SortSlice`로 심볼 이진탐색해 확인한다(1=정상, 0=주문정지, 3=매도만 가능).

파일 구성: `recvOrder.go`(부트스트랩, REQ 핸들러, 접수번호/주문상태 mmap 관리), `dbfunc.go`(환율 조회, 주문 INSERT, 증거금 확인 DB 로직), `jsonfunc.go`(`ValidateOrder` — 수신 JSON 검증 및 내부 패킷 변환), `exrate.go`(환율 SUB 처리 보조, Makefile 빌드 목록에는 없음 — 아래 주의사항 참고).

## 빌드

```bash
cd goPlus/RecvOrder/NetLog
make
make install
```

## 사용법

```bash
goRecvOrder <index>
```

`<index>`는 0~99 범위(여러 인스턴스를 두어 부하 분산). 범위를 벗어나거나 인자가 없으면 사용법을 출력하고 종료. 중복 실행 방지: `goRecvOrder-{index:02d}` lock.

## 환경설정

| 키 | 용도 |
|---|---|
| `PROXYENV_YAML` | `ORDERPASS`(주문 REQ/REP, back), `BEFOREMATCHING`(매칭엔진 전달, front), `EXTRAINFOR`(환율/CONTROL, back) |
| `ORDERCOND_MAP` | 코인별 주문 가능 상태 mmap 파일 경로 |
| `REGNUM_MAP` | 접수번호 mmap 파일 경로(인덱스별 슬롯) |
| `DBCONFIG_YAML` | PostgreSQL 접속 정보(`centolib/api.ConnectDb`) |

## 주의사항

- `Makefile`의 `SRC`에 **`exrate.go`가 빠져 있다**(`recvOrder.go dbfunc.go jsonfunc.go`만 명시) — `go build`(패키지 단위 컴파일)로는 문제없이 포함되지만 `make` 의존성 추적 목적으로는 빠진 상태다.
- DB INSERT(`DbInsertOrder`) 실패 시 코드 주석에 `증거금 rollback해야함-- TODO`가 있다 — **증거금 롤백 로직이 미구현** 상태로 남아있는 것으로 보이므로, 이 경로에서 증거금 정합성 이슈가 발생할 수 있다.
- `<index>`가 0~99를 벗어나면 즉시 종료한다.
- 매일 자정 부근에서 5초 주기로 날짜 변경을 감지해 접수번호 시퀀스를 초기화한다(`goCheckDate`).
