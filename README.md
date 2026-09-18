# GoProject — TugBoat 워크스페이스

암호화폐 거래소 매칭엔진/운영 시스템("TugBoat", `tb*`) 관련 Go 프로그램 모음. 하나의 애플리케이션이 아니라 `go.work`로 연결된 여러 독립 Go 모듈과 실험용 코드(`Lab/`)로 구성된 개인용 멀티모듈 워크스페이스이다.

레거시 C 시스템(`clschaos`, `tb_nova` 등)을 Go로 재작성하는 작업이 진행 중이며, 각 소스 파일 상단 주석에 설계 배경과 C 버전 대응 관계가 상세히 기록되어 있다.

## 구성 모듈

| 모듈 | 설명 | 문서 |
|---|---|---|
| [centolib](centolib/README.md) | 공용 라이브러리 (DB/ZMQ/네트워크/로깅 등) — 다른 모듈이 참조 | 9개 하위 패키지별 API 사양서 포함 |
| [goPlus](goPlus/README.md) | 거래소 매칭 엔진 + 운영/검증 도구 모음 | 하위 50여 개 디렉토리 각각 README 보유 |
| [goTbnova](goTbnova/README.md) | 레거시 C 프로세스 관리자를 대체하는 Go 프로세스 슈퍼바이저 데몬 | 내부 패키지별 API 사양서 포함 |
| [goClsTbnova](goClsTbnova/README.md) | goTbnova 종료 후 고아가 된 관리 프로세스 정리 유틸리티 | [상세 매뉴얼](goClsTbnova/goClsTbnova.md) |
| [goTbmon](goTbmon/README.md) | goTbnova를 TCP로 제어/모니터링하는 Bubble Tea 터미널 UI | |
| [goapisvr](goapisvr/README.md) | `centolib/fulcrum` 기반 잔액 조회 WebSocket API 서버 | |
| [getbal](getbal/README.md) | Fulcrum 서버로 비트코인 잔액을 조회하는 독립 CLI 예제 | ⚠️ 빌드 불가 (아래 참고) |
| [Lab](Lab/README.md) | Go 언어/라이브러리 실험용 스크래치 코드 모음 | 정식 모듈 아님 |

## 요구사항

- Go 1.24.5 (`go.work`), 일부 모듈은 `go.mod`에 `go 1.22.2` 명시
- PostgreSQL (`github.com/lib/pq`) — `centolib/api`, `centolib/matchingboard`, goPlus 다수 컴포넌트가 사용
- ZeroMQ C 라이브러리(libzmq) — `github.com/pebbe/zmq4`(CGO 바인딩)를 사용하는 `centolib/tbzmq`, goPlus 매칭/프록시 계열이 필요
- Vector(로그 수집기) — `netlog` 백엔드 사용 시 `/tmp/VECTOR.UNIX.DGRAM` UNIX Datagram 소켓 필요 (없으면 `file` 백엔드로 전환하거나 기동 실패)

## 빌드

통합 빌드 시스템은 없다. 워크스페이스 루트에서 `go.work`에 포함된 모듈을 한 번에 대상으로 할 수도 있고, 각 도구 디렉토리의 `Makefile`/`build.sh`로 개별 빌드할 수도 있다.

```bash
# go.work 범위 전체
go build ./...
go vet ./...

# 개별 모듈
cd goTbnova && go build -o goTbnova

# Makefile이 있는 goPlus 하위 도구
cd goPlus/Matching/NetLog && make install   # 빌드 + strip + ~/bin 이동
```

세부 빌드/사용법은 각 모듈 README를 참고할 것.

## 알려진 이슈 / 빌드 전 확인 사항

- **`centolib/go.mod`에 `require` 항목이 없고 `go.sum`도 없다** — 소스는 `github.com/lib/pq`, `github.com/pebbe/zmq4`, `github.com/btcsuite/btcutil/bech32`, `go.uber.org/zap`, `gopkg.in/natefinch/lumberjack.v2` 등 8개 외부 패키지를 임포트하지만 모듈 메타데이터에 반영되어 있지 않다. `go build`/`go vet` 전에 `cd centolib && go mod tidy`가 필요할 가능성이 높다.
- **`goPlus/go.mod`에도 `github.com/pebbe/zmq4`, `github.com/lib/pq` 등이 빠져 있다** — 마찬가지로 `go mod tidy` 필요.
- **`getbal/`은 자체 `go.mod`가 없고 `go.work`에도 포함되지 않아** 현재 상태로는 빌드되지 않는다.
- `goPlus/Exrate60/{LocalLog,NetLog}/exrate60.go`에 **Alpha Vantage API 키가 소스에 하드코딩**되어 있다 — 공유/배포 전 키 분리 및 교체 권장.
- `goPlus/tools/매칭검증/cancel/`은 DB의 대기 주문을 실제로 취소 전송하는 도구다 — 반드시 테스트 DB/환경에서만 사용할 것. `go build`(패키지 경로 지정)는 디렉토리명이 한글이라 import path 규칙에 걸려 실패하며, `go build <파일명>.go` 형태로만 빌드해야 한다.
- `go/` 디렉토리는 소스가 아니라 `GOPATH`(`go/pkg/mod`, `go/bin`) 데이터이며 `.gitignore`에 포함되어 있다.

## 저장소 구조

```
go.work            워크스페이스 모듈 정의 (Lab, centolib, goClsTbnova, goPlus, goTbmon, goTbnova, goapisvr)
CLAUDE.md          Claude Code용 작업 가이드 (git에는 포함 안 됨, .gitignore)
<각 모듈>/README.md  모듈별 사용법·환경설정·주의사항
```

각 하위 디렉토리에 대한 더 자세한 아키텍처 설명(goTbnova 프로세스 슈퍼바이저 구조, goPlus의 LocalLog/NetLog 이중 빌드 패턴 등)은 해당 모듈의 README.md를 참고할 것.
