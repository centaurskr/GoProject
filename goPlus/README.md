# goPlus

암호화폐 거래소 매칭 엔진 및 운영 도구 모음.

## 개요

`goPlus`는 `TugBoat`("tb") 거래소 시스템의 핵심 매칭 파이프라인(주문 수신 →
매칭/체결 → 후처리 → 지갑/환율 갱신)과, 운영·디버깅을 위한 다수의 독립 CLI 도구
(`Ops/`, `tools/`)를 담은 모듈이다. 통합 애플리케이션이 아니라 각자 별도로 빌드되고
배포되는 다수의 소형 데몬/도구 집합이다.

- `module goPlus`, `go 1.22.2` (`go.mod`)
- 직접 의존성: `github.com/creack/pty`, `github.com/kr/pty`, `golang.org/x/sys`,
  `golang.org/x/term` (indirect로 표기되어 있음). 각 하위 도구는 실제로
  `centolib`, `github.com/pebbe/zmq4`, `github.com/bwmarrin/snowflake`,
  `gopkg.in/yaml.v3` 등도 사용한다 — `go.work`를 통해 워크스페이스 루트에서 해석된다.

## 메시징 아키텍처

컴포넌트 간 통신은 ZeroMQ(`centolib/tbzmq`, `github.com/pebbe/zmq4`) 기반이다.
`Proxy/PubSubProxy`(PUB/SUB 중계, XSUB↔XPUB)와 `Proxy/ReqRepProxy`(REQ/REP 중계,
ROUTER↔DEALER, 메시지 캡처 지원)가 각 파이프라인 단계 사이의 메시지를 중계한다.
프록시들의 엔드포인트(front/back/control/capture)는 각 도구가 `PLATFORM_ENV` →
`PROXYENV_YAML` 경로의 YAML 파일에서 이름으로 조회한다(예: `BEFOREMATCHING`,
`AFTERMATCHING`, `ORDERPASS`, `REALJSON`).

## LocalLog / NetLog 이중 빌드 패턴

핵심 컴포넌트(Matching, MatchingCtl, AfterMatch, RecvOrder, Wallet, Exrate60,
Proxy)는 각각 `LocalLog/`와 `NetLog/` 하위 디렉토리에 **거의 동일한 소스가
중복**되어 있다. 두 버전의 실질적 차이는 로깅 백엔드와(대체로) 실행 인자 개수뿐이다:

- `LocalLog/` → `centolib/logout` 사용. `Log.InitLogger(id, mode)`로 초기화하며
  `mode`는 `file`/`console`/`all` 중 하나를 실행 인자로 받는다.
- `NetLog/` → `centolib/netlog` 사용. `Log.InitNetLogout(os.Args[0], id)`로
  초기화하며, 로그 모드 인자를 받지 않는다(항상 네트워크로 전송).

한쪽을 수정할 때는 대응하는 다른 쪽에도 동일한 로직 변경이 필요한지 반드시
확인할 것. 실행 인자 형식은 컴포넌트마다 조금씩 다르므로(예: `goMatching`은
LocalLog가 `<Index> <mode>`, NetLog가 `<Index>`만 받음) 각 디렉토리의 README를
개별로 참고한다.

## 빌드

통합 빌드 시스템이 없다. 각 leaf 디렉토리(실제 소스가 있는 하위 폴더)마다 개별
`Makefile`이 있고, 표준 타겟은 다음과 같다.

```bash
cd goPlus/<컴포넌트>/<LocalLog|NetLog>   # 또는 goPlus/Ops/<도구>, goPlus/tools/매칭검증/<도구>
make            # go build (+ 일부는 strip 포함)
make install    # clean → build → strip → ~/bin 으로 이동
```

## 컴포넌트 인덱스

### 매칭 파이프라인 (LocalLog / NetLog 쌍)

| 컴포넌트 | LocalLog | NetLog | 설명 |
|---|---|---|---|
| Matching | [README](Matching/LocalLog/README.md) | [README](Matching/NetLog/README.md) | 코인별 지정가 매칭(체결) 엔진 |
| MatchingCtl | [README](MatchingCtl/LocalLog/README.md) | [README](MatchingCtl/NetLog/README.md) | 매칭/후처리 프로세스 감독, N-Value·주문상태 관리 |
| AfterMatch | [README](AfterMatch/LocalLog/README.md) | [README](AfterMatch/NetLog/README.md) | 체결 후 DB 반영 및 결과 재발행 |
| RecvOrder | [README](RecvOrder/LocalLog/README.md) | [README](RecvOrder/NetLog/README.md) | 주문 수신 및 전처리 |
| Wallet | [README](Wallet/LocalLog/README.md) | [README](Wallet/NetLog/README.md) | 지갑/잔고 처리 |
| Exrate60 | [README](Exrate60/LocalLog/README.md) | [README](Exrate60/NetLog/README.md) | 환율 갱신(60초 주기 추정, 각 README 참고) |
| Proxy/PubSubProxy | [README](Proxy/LocalLog/PubSubProxy/README.md) | [README](Proxy/NetLog/PubSubProxy/README.md) | ZMQ PUB/SUB 중계 프록시 |
| Proxy/ReqRepProxy | [README](Proxy/LocalLog/ReqRepProxy/README.md) | [README](Proxy/NetLog/ReqRepProxy/README.md) | ZMQ REQ/REP 중계 프록시 |

### 운영/디버깅 도구 (Ops/)

| 도구 | 설명 링크 |
|---|---|
| DumpCoin | [Ops/Dump/DumpCoin/README.md](Ops/Dump/DumpCoin/README.md) |
| DumpCond | [Ops/Dump/DumpCond/README.md](Ops/Dump/DumpCond/README.md) |
| DumpOrdMap | [Ops/Dump/DumpOrdMap/README.md](Ops/Dump/DumpOrdMap/README.md) |
| Hoga | [Ops/Hoga/README.md](Ops/Hoga/README.md) |
| NewCoin | [Ops/NewCoin/README.md](Ops/NewCoin/README.md) |
| PlatformCheck | [Ops/PlatformCheck/README.md](Ops/PlatformCheck/README.md) |
| ProxyCtl | [Ops/ProxyCtl/README.md](Ops/ProxyCtl/README.md) |
| PubZmq | [Ops/PubZmq/README.md](Ops/PubZmq/README.md) |
| RegNumInit | [Ops/RegNumInit/README.md](Ops/RegNumInit/README.md) |
| SubJson | [Ops/SubJson/README.md](Ops/SubJson/README.md) |
| SubZmq | [Ops/SubZmq/README.md](Ops/SubZmq/README.md) |
| ViewJson | [Ops/ViewJson/README.md](Ops/ViewJson/README.md) |

### 매칭 검증 도구 (tools/매칭검증/)

| 도구 | 설명 링크 |
|---|---|
| RcvSEQ | [tools/매칭검증/RcvSEQ/README.md](tools/매칭검증/RcvSEQ/README.md) |
| cancel | [tools/매칭검증/cancel/README.md](tools/매칭검증/cancel/README.md) |
| randorder | [tools/매칭검증/randorder/README.md](tools/매칭검증/randorder/README.md) |
| stepord | [tools/매칭검증/stepord/README.md](tools/매칭검증/stepord/README.md) |

### 데이터

- [SQL 스키마](SQL/README.md) — `coininfor`, `ordertbl`, `tradetbl`, `ticktbl`,
  `wallet`, `exratetbl`, `errtbl` 테이블 정의.

## 환경설정 (공통)

goPlus 하위 대부분의 실행 파일은 환경변수 `PLATFORM_ENV`가 가리키는
`KEY=VALUE` 형식 설정 파일(`centolib/api.GetPlatformEnv`)을 통해 동작에 필요한
값(코인 목록 파일, YAML 경로, TCP 접속 정보 등)을 읽는다. 세부 키는 각 컴포넌트별
README의 "환경설정" 절을 참고한다. 이 파일의 실제 경로/내용은 저장소에 포함되어
있지 않으므로(운영 서버에만 존재), 신규 환경 구축 시 각 컴포넌트가 참조하는 키
목록을 모아 직접 작성해야 한다.

## 주의사항

- `Matching/`, `MatchingCtl/`, `AfterMatch/`, `RecvOrder/`, `Wallet/`, `Exrate60/`,
  `Proxy/`, `Proxy/LocalLog/`, `Proxy/NetLog/`, `Ops/`, `Ops/Dump/`, `tools/`,
  `tools/매칭검증/` 디렉토리 자체에는 소스 파일이 없다 — 하위 leaf 디렉토리를
  묶는 용도이며 위 인덱스의 링크로 이동해서 실제 문서를 참고한다.
- 테스트 코드(`*_test.go`)가 없다. 변경 후에는 대상 바이너리를 직접 빌드해
  확인해야 한다.
- `goPlus/.claude/settings.local.json`은 이 저장소에서 과거 작업 시 사용된 로컬
  권한 설정 파일로, 프로젝트 동작과는 무관하다.
