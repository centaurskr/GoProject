# SubZmq (goSubZmq)

ZMQ SUB 소켓으로 지정한 엔드포인트에 연결해 수신되는 메시지를 그대로 출력하는 가장 기본적인 디버깅용 CLI 도구.

## 개요

메시지의 첫 두 파트(`msg[0]`=topic, `msg[1]`=payload)를 `topic-payload` 형태로 계속 출력한다. JSON pretty-print가 필요하면 `../SubJson`을 사용한다.

## 빌드

```bash
cd goPlus/Ops/SubZmq
make            # goSubZmq 빌드
make install    # 빌드 + strip + ~/bin 으로 이동
```

## 사용법

```
goSubZmq <endpoint> [topic]
```

- `<endpoint>`: 연결할 엔드포인트 (예: `tcp://127.0.0.1:23423`)
- `[topic]` (선택): 구독할 topic. 생략하면 전체(`""`) 구독.

인자가 1개 미만이면 사용법을 출력하고 종료한다.

## 환경설정

없음 — 모든 입력을 커맨드라인 인자로 받는다.

## 주의사항

- `socket.Connect()` 직후 1초 고정 대기(slow joiner 대응) 후 구독을 시작한다.
- 무한 루프로 계속 수신하며 별도 종료 처리가 없다 — Ctrl+C 등으로 강제 종료해야 한다.
- 수신 메시지 파트가 2개 미만이면(`msg[1]` 접근 시) 인덱스 아웃오브레인지로 패닉할 수 있다(에러 처리 없음).
