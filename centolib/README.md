# centolib

거래소(TugBoat/`tb*`) 시스템의 여러 Go 모듈이 공유하는 공용 라이브러리 모듈.

## 개요

- `module centolib`, `go 1.22.2` (`go.mod`).
- 순수 애플리케이션 코드(빌드되는 실행 파일)가 아니라, `goTbnova`, `goapisvr`, `getbal`, `goPlus`의 각 구성요소 등이 가져다 쓰는 라이브러리 패키지 모음이다.
- 저수준 시스템 유틸리티(mmap 파일, raw TCP socket, epoll 이벤트 루프), ZeroMQ 래퍼, 로깅 백엔드(로컬/네트워크), 매칭엔진 주문/호가(order book) 데이터 구조, Bitcoin Fulcrum 서버 통신, 터미널 제어(ANSI) 등을 제공한다.

## 사용법

이 저장소는 `go.work`로 워크스페이스가 구성되어 있어, `go.work`의 `use` 목록에 `./centolib`가 포함되어 있으면 워크스페이스 내 다른 모듈에서 별도 `replace` 없이 바로 import할 수 있다.

`go.work` 없이 개별 모듈만 빌드하는 경우(`goTbnova`의 예)에는 `go.mod`에 로컬 경로 `replace` 지시자를 추가한다:

```go
require centolib v0.0.0-00010101000000-000000000000
replace centolib => /home/cento/Project/Workspace/centolib
```

이후 코드에서는 패키지 경로로 바로 import한다:

```go
import (
    "centolib/api"
    "centolib/tbzmq"
    "centolib/logout"
)
```

## 패키지 목록

| 패키지 | 설명 |
|---|---|
| [api](api/README.md) | DB 접속, env/yaml 파일 읽기, mmap 파일, raw TCP 소켓, epoll 기반 이벤트 루프, 구조체 정렬/이진탐색, 시간 포맷 등 범용 유틸리티 모음 |
| [bindecimal](bindecimal/README.md) | 고정 길이 바이너리 버퍼(mmap/파일 저장용)에 `shopspring/decimal` 값을 담는 `BinDecimal` 타입과 사칙연산 |
| [ctlzmq](ctlzmq/README.md) | ZeroMQ PUB/SUB 기반 "CONTROL" 채널로 프로세스 간 제어 메시지를 송수신 |
| [fulcrum](fulcrum/README.md) | Fulcrum(Electrum 프로토콜 서버)에 TCP로 접속해 Bitcoin Bech32 주소의 잔액을 조회 |
| [logout](logout/README.md) | zap 기반 로컬 로거. 콘솔/파일/둘다 출력 모드와 레벨 색상 지원 (`goPlus/*/LocalLog` 빌드가 사용) |
| [matchingboard](matchingboard/README.md) | 매칭엔진의 호가/주문 데이터 구조(`MEMCOIN`, `MEMHOGA`, `Order` 등) — mmap 파일 기반 order book |
| [netlog](netlog/README.md) | Unix Domain Socket으로 Vector→VictoriaLogs에 비동기 전송하는 네트워크 로거 (`goPlus/*/NetLog`, `goTbnova/tlog`가 사용) |
| [screen](screen/README.md) | 터미널 raw mode 전환, 커서/색상 제어를 위한 ANSI 이스케이프 시퀀스 상수 및 함수 |
| [tbzmq](tbzmq/README.md) | ZeroMQ 소켓 monitor 및 PUB/SUB 프록시 래퍼 |

## 주의사항

- 다른 모듈들은 이 라이브러리를 (버전 태그가 아니라) `replace` 지시자로 로컬 경로를 직접 참조하는 경우가 많다 (`goTbnova/go.mod` 등). 따라서 `centolib`를 수정하면 별도 배포/버전업 없이 참조하는 모든 모듈에 즉시 반영된다 — 여러 모듈이 동시에 빌드/배포되는 시점을 함께 고려해야 한다.
- `logout`(LocalLog)과 `netlog`(NetLog)는 API가 다르고 둘 다 동시에 쓰이는 패키지가 아니다. `goPlus`의 각 컴포넌트는 두 백엔드 중 하나만 골라 빌드하는 관례(`LocalLog/` `NetLog/` 디렉토리 분리)를 따른다.
- `matchingboard`, `api`(mmap, raw socket, epoll)는 `unsafe`/`syscall`을 직접 다루며 Linux 전용이다. 다른 OS에서는 빌드/동작이 보장되지 않는다.
- 테스트 코드(`*_test.go`)가 없다. 변경 시 의존 모듈(특히 `goPlus`의 매칭 관련 컴포넌트)에서 직접 빌드/동작 확인이 필요하다.
