# RcvSEQ (recvseq)

AfterMatch(사후처리) 컴포넌트가 발행하는 패킷을 수신해 `DBSEQ`/`DBCHE` 메시지 개수를 실시간 카운트하는 매칭엔진 검증용 도구.

## 개요

`PROXYENV_YAML`에서 `AFTERMATCHING` 프록시의 `back` 엔드포인트를 조회해 ZMQ SUB 소켓으로 전체 구독(`""`)한다. `centolib/api.GoEvent` 이벤트 루프에 SUB 소켓을 등록하고, 수신 메시지의 두 번째 파트(`msg[1]`)가 `DBSEQ`/`DBCHE`인지에 따라 누적 카운트를 콘솔에 출력한다. AfterMatch가 DB 반영 순번(SEQ)과 체결(CHE) 메시지를 빠짐없이 보내는지 수량으로 검증하는 용도다.

## 빌드

이 디렉토리에는 Makefile이 없다.

```bash
cd goPlus/tools/매칭검증/RcvSEQ
go build -o recvseq recvseq.go
```

## 사용법

```
recvseq
```

인자 없이 실행한다. 실행하면 안내 문구를 출력하고 즉시 수신을 시작해, 메시지가 올 때마다 누적 카운트를 계속 출력한다.

## 환경설정

- `PROXYENV_YAML` (플랫폼 환경 설정, `centolib/api.GetPlatformEnv`로 조회): 프록시 접속 정보 YAML 경로. 이 안에 `AFTERMATCHING` 항목이 정의되어 있어야 한다.

## 주의사항

- `centolib/api.GoEvent` 이벤트 루프(`GoEventLoop()`)로 동작하며 별도 종료 처리가 없다 — Ctrl+C 등으로 강제 종료해야 한다.
- 카운트는 프로세스 메모리에만 유지되며 파일/DB에 기록하지 않는다 — 재시작하면 초기화된다.
- 수신 메시지가 3파트(`msg[0..2]`) 미만이면 `msg[1]` 접근 시 패닉할 수 있다(에러 처리 없음).
