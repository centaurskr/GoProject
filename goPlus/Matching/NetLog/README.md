# goMatching (NetLog)

코인별 주문 매칭(체결) 엔진. `goPlus` 거래소 시스템의 핵심 컴포넌트 — 주문을 수신해 호가창(order book)에 반영하고 매칭·체결 결과를 후속 컴포넌트로 전송한다. 로깅 백엔드는 [centolib/netlog](../../../centolib/netlog/README.md)(Vector→VictoriaLogs)를 사용하는 버전이며, 로컬 로그 버전은 [../LocalLog](../LocalLog/README.md) 참고.

## 개요

코인 1개당 고루틴 1개(`gofCoinMatching`)로 매칭을 처리하며, 프로세스 1개가 `process.toml`/`ORDER_LIST` 환경파일의 인덱스에 해당하는 코인 그룹 전체를 담당한다. 내부 통신은 in-process ZMQ PUB/SUB 프록시(`inproc://IPC_FRONT-{pid}`, `IPC_BACK-{pid}`)로 일원화되어 있다.

```
                                   ┌───────────────┐
【PUB/SUB Proxy(RecvOrder)】      │ 🔴 goMatching │  ⇆  CONTROL 메시지 (goMatchingCtl)
         🡇                        └───────────────┘
  (코인별 gofCoinMatching) ──IPCFDS──▶ gofFdsFunction  ──▶ (FDS, 현재 미사용/스텁)
                           ──IPCSEQ/IPCCHE/IPCCANCEL/IPCMKDONE/IPCERR──▶
                              gofAfterFunction ──▶ AFTERMATCHING Proxy (DBCHE/DBSEQ/DBCANCEL/DBERR/MKDONE)
                           ──IPCCURR/IPCHOGA──▶ gofSiseFunction ──▶ SISEJSON/HOGAJSON Proxy
```

파일 구성: `MatMain.go`(부트스트랩, 일자변경/제어메시지 처리), `memefunc.go`(코인별 매칭 알고리즘 — 지정가/시장가/취소 처리, 호가창 조작), `afterfunc.go`(체결 후처리: 증거금 계산, `AFTERMATCHING`으로 DB 저장용 패킷 전송), `sisefunc.go`(현재가/호가 시세 전송), `fdsfunc.go`(FDS 이상거래탐지 훅, 현재 로직은 주석 처리된 스텁), `clsmapfile.go`(일자변경 시 주문 mmap 파일 정리/재배치), `utilfunc.go`(경로 확인 유틸).

## 빌드

```bash
cd goPlus/Matching/NetLog
make            # ./goMatching 생성 (strip 포함)
make install    # 빌드 + ~/bin 으로 이동
```
`Makefile`의 `SOURCE`는 `MatMain.go afterfunc.go fdsfunc.go memefunc.go sisefunc.go clsmapfile.go`이며, **`utilfunc.go`는 목록에 없다**(같은 디렉토리에 있지만 `make`의 명시적 소스 목록에서 빠져 있음 — `go build`가 패키지 단위로 컴파일하므로 실제 빌드에는 포함되지만 Makefile의 의존성 추적 대상에서는 빠져 있는 상태).

## 사용법

```bash
goMatching <Index>
```

`<Index>`는 1부터 시작하는 코인 그룹 인덱스(`ORDER_LIST` 환경파일의 `ORD_{Index}` 라인에 해당하는 코인들을 처리). 인자 개수가 다르면 사용법을 출력하고 종료.

중복 실행 방지: `Matching_{Index}` 이름으로 `/tmp/*.lock` flock을 사용하므로, 동일 인덱스로 두 번 기동하면 즉시 종료된다.

## 환경설정

`centolib/api.GetPlatformEnv`로 읽는 환경변수들(`$PLATFORM_ENV` 파일 안에 정의):

| 키 | 용도 |
|---|---|
| `PROXYENV_YAML` | ZMQ 프록시 엔드포인트 정의 yaml 경로. `BEFOREMATCHING`(주문 수신), `AFTERMATCHING`(체결 후처리 전송), `SISEJSON`/`HOGAJSON`(시세/호가 전송), `EXTRAINFOR`(CONTROL 채널) 항목을 조회 |
| `ORDER_LIST` | 코인 인덱스별 목록 파일(`ORD_N=coin1:coin2:...`) — [centolib/matchingboard](../../../centolib/matchingboard/README.md)의 `GetCoinListByIndex` 형식 |
| `MATCHING_HOST`, `MATCHING_STREAM` | `goMatchingCtl` TCP 접속 주소(host+stream 문자열 결합) |
| `ORDPATH` | 주문/호가 mmap 파일 저장 루트 경로 |

## 주의사항

- **NetLog 버전은 `$LOG_HOME` 대신 Vector UDS 소켓**(`centolib/netlog`, 기본 `/tmp/VECTOR.UNIX.DGRAM`)을 사용한다 — 소켓이 없으면 `InitNetLogout` 실패로 즉시 종료된다.
- `goMatchingCtl`과의 TCP 접속이 끊기면(`MatchingCtlHandler`) `ORDERPASS` 프록시에 `PAUSE`를 보내 **주문 수신을 자동 중단**시킨다 — `goMatchingCtl` 없이 장시간 운영하면 신규 주문이 막힐 수 있다.
- 코인 그룹을 담당하는 인덱스(`<Index>`)는 **1부터 시작**해야 한다(0 이하 입력 시 즉시 에러 종료).
- 체결/주문 데이터는 [centolib/matchingboard](../../../centolib/matchingboard/README.md)의 mmap 기반 order book을 직접 조작한다 — `ORDPATH`의 파일시스템 상태와 프로세스 실행 상태가 강하게 결합되어 있으므로, 비정상 종료 후 재기동 시 mmap 파일 일관성에 유의해야 한다.
- `RELOAD` 제어 메시지(`goMatchingCtl`발) 수신 시 코인 목록을 다시 읽어 신규 코인은 고루틴을 추가 기동하고, 목록에서 빠진 코인에는 `STOP`을 보내 종료시킨다(런타임 중 무중단 코인 추가/제거 지원).
