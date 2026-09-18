# Lab

Go/ZeroMQ/여러 라이브러리를 익히고 검증하기 위한 개인 실험·스크래치 코드 모음. **정식 프로덕션 모듈이 아니다** — `goPlus`/`goTbnova` 등에 들어가기 전 아이디어를 시험하는 용도.

## 개요

- `module Lab`, `go 1.24.0`(toolchain 1.24.5). 단일 `go.mod`로 묶여 있지만 실제로는 서로 무관한 수십 개의 독립 실험이 한 모듈에 섞여 있다.
- 루트에 100개 이상의 단독 `.go` 파일(각각 `package main`인 독립 스니펫: 채널, 클로저, 인터페이스, ZMQ 스트리밍, 정렬, JSON, mmap 맵파일 등 Go 언어/라이브러리 기능 실험)이 직접 놓여 있다.
- 하위 폴더는 좀 더 완결된 실험 단위(디렉토리당 보통 `Makefile` + 여러 `.go` 파일)이다. 대표적으로:
  - `After/` — `goPlus/AfterMatch`의 원형으로 보이는 체결 후처리 실험(큐, DB, JSON 처리).
  - `Zmq/`, `ReqRepSys/`(`Cli/DbSvr/OrdBrg/Proxy/Rcv`) — ZeroMQ REQ/REP·PUB/SUB·프록시 패턴 실험.
  - `HeartBeatZmq/` — ZMQ 연결 상태 추적/하트비트 패턴에 대한 자체 분석 노트(`README.TXT`, 한글 기술 메모).
  - `WAL/btc_crawler/` — 자체 `README.md`가 있는 별도 소형 프로젝트(Arkham Intel 웹사이트에서 BTC 전송 내역을 크롤링하는 Go 프로그램). 자체 `go.mod`를 가진 독립 서브모듈이다.
  - `ordermap/`, `Wallet/`, `exrate/`, `decimal/` — `centolib/matchingboard`/`bindecimal` 관련 프로토타입.
  - `Kafka/`, `Timescale/`, `Netlog/`, `TimeRand/`, `Visit-sim/`, `TCP/`, `BuffIO/`, `Event/`, `Array/`, `ReadDir/`, `keytest/`, `yaml/` — 각각 이름이 가리키는 라이브러리/기능(Kafka, TimescaleDB, 텍스트 로그, 난수, 크롤링 시뮬레이션, raw TCP, 버퍼드 IO, epoll 이벤트, 슬라이스, 디렉토리 읽기, 키 입력, YAML) 단발성 실험.

## 빌드/실행

통합 빌드 규칙이 없다. 루트의 개별 파일은 보통 그 자체로 하나의 `package main`이므로:

```bash
cd Lab
go run 특정파일.go
```

하위 폴더 중 `Makefile`이 있는 곳은 `make`로 빌드 가능(각 폴더의 `Makefile` 참고). `WAL/btc_crawler`는 독자적인 `go.mod`를 가지므로 그 디렉토리 안에서 `go mod tidy && go run btc_web_crawler.go`로 실행한다(자세한 내용은 [WAL/btc_crawler/README.md](WAL/btc_crawler/README.md)).

## 주의사항

- **실험/학습 코드이며 품질·안정성이 보장되지 않는다.** 프로덕션(`goPlus`, `goTbnova` 등)에 있는 동일 아이디어의 완성본을 우선 참고할 것 — 예: `centolib/matchingboard`가 `Lab/ordermap`의 완성형이고, `goPlus/AfterMatch`가 `Lab/After`의 완성형이다.
- 컴파일된 바이너리(예: `TimeRand/goTimeRandom`, `Netlog/goNetlog` 등)를 커밋하지 말 것 — 과거 이런 산출물이 하위 폴더에 남아있던 적이 있으므로 새 실험 후 정리 시 주의한다.
- `go.mod`의 `require` 목록이 매우 크고 대부분 `// indirect`다 — 하위 폴더 하나를 위해 저장소 전체의 의존성이 함께 받아진다. 새 실험에 새 의존성을 추가하면 이 루트 `go.mod`/`go.sum`에 반영된다.
- 한글로 된 자유 형식 메모(`HeartBeatZmq/README.TXT` 등)가 군데군데 있다 — 정식 문서가 아니라 당시의 기술 분석 노트이므로 최신 상태를 보장하지 않는다.
