# SubJson (goSubJson)

ZMQ SUB 소켓으로 멀티파트 메시지를 수신해, 그 중 JSON 파트를 보기 좋게(pretty-print) 정렬해 출력하는 디버깅용 CLI 도구.

## 개요

`SubZmq`(원문 메시지 그대로 출력)와 달리, 메시지 파트 중 JSON이 담긴 인덱스를 지정하면 `encoding/json.Indent`로 들여쓰기하여 출력한다. Pub/Sub 프록시가 내보내는 JSON 페이로드를 사람이 읽기 좋게 확인할 때 사용한다.

## 빌드

```bash
cd goPlus/Ops/SubJson
make            # goSubJson 빌드
make install    # 빌드 + strip + ~/bin 으로 이동
```

## 사용법

```
goSubJson <endpoint> <json index> [topic]
```

- `<endpoint>`: 연결할 엔드포인트 (예: `tcp://127.0.0.1:23423`)
- `<json index>`: 메시지 파트 중 JSON이 위치한 인덱스
  - Topic + JSON 구조면 `1`
  - Topic + MSG + JSON 구조면 `2`
- `[topic]` (선택): 구독할 topic. 생략하면 전체(`""`) 구독.

인자가 3개 미만이면 사용법을 출력하고 종료한다. `<json index>`가 `0`이거나 숫자로 변환 실패 시("index 오류") 종료한다.

## 환경설정

없음 — 모든 입력을 커맨드라인 인자로 받는다.

## 주의사항

- 소스 내 사용법 메시지에 프로그램 이름이 `goSubZmq`로 잘못 표기되어 있다(실제 빌드 결과물은 `goSubJson`) — 복사 과정에서 남은 오타로 보인다.
- `socket.Connect()` 직후 1초 고정 대기(slow joiner 대응) 후 구독을 시작한다.
- 무한 루프로 계속 수신하며 별도 종료 처리가 없다 — Ctrl+C 등으로 강제 종료해야 한다.
- 수신 메시지의 파트 개수가 `<json index>`보다 적으면 `msg[idx]` 접근 시 인덱스 아웃오브레인지로 패닉할 수 있다(에러 처리 없음).
