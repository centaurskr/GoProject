# PubZmq (goPubZmq)

임의의 ZMQ PUB 엔드포인트로 토픽+메시지를 1회 발행(publish)하는 테스트/디버깅용 CLI 도구.

## 개요

지정한 엔드포인트에 ZMQ PUB 소켓으로 연결한 뒤, 커맨드라인으로 받은 topic과 메시지 문자열(들)을 멀티파트 메시지로 전송한다. SUB 구독자가 정상 동작하는지, 특정 이벤트를 수동으로 주입해 테스트할 때 사용한다.

## 빌드

```bash
cd goPlus/Ops/PubZmq
make            # goPubZmq 빌드
make install    # 빌드 + strip + ~/bin 으로 이동
```

## 사용법

```
goPubZmq <endpoint> <topic> <message>
```

- `<endpoint>`: 연결할 엔드포인트 (예: `tcp://127.0.0.1:23423`)
- `<topic>`: 메시지 토픽
- `<message>`: 전송할 메시지 (인자를 더 추가하면 멀티파트 메시지로 함께 전송됨)

인자가 4개 미만이면 사용법을 출력하고 종료한다.

## 환경설정

없음 — 모든 입력을 커맨드라인 인자로 받는다.

## 주의사항

- `socket.Connect()` 직후 구독자가 소켓을 준비할 시간을 주기 위해 1초(`time.Sleep(1000ms)`)를 무조건 대기한 뒤 전송한다 — "slow joiner" 문제 회피용이며, 빠르게 반복 실행할 때는 이 지연을 감안해야 한다.
- PUB 소켓 특성상 연결 시점에 구독자가 없으면 메시지가 유실될 수 있다(ZMQ PUB/SUB의 일반적 동작).
