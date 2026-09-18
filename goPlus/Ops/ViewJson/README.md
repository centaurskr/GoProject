# ViewJson (goViewJson)

한 줄에 JSON 하나씩 기록된 로그 파일을 읽어, 각 줄을 들여쓰기된 형태로 보기 좋게 출력하는 로그 열람용 CLI 도구.

## 개요

파일을 줄 단위로 스캔(`bufio.Scanner`)하며 각 줄을 `encoding/json.Indent`로 정렬해 출력한다. JSON 파싱에 실패한 줄은 원본 그대로와 에러를 함께 출력하고 다음 줄로 넘어간다(중단하지 않음).

## 빌드

```bash
cd goPlus/Ops/ViewJson
make            # goViewJson 빌드
make install    # 빌드 + strip + ~/bin 으로 이동
```

## 사용법

```
goViewJson <json 로그 파일>
```

인자가 정확히 1개(파일 경로)가 아니면 사용법을 출력하고 종료한다.

## 환경설정

없음 — 파일 경로 하나만 인자로 받는다.

## 주의사항

- 파일 전체가 아니라 **줄 단위(1줄 = JSON 객체 1개)** 포맷을 전제로 한다 — 여러 줄에 걸친 JSON이나 JSON 배열 전체 파일에는 적합하지 않다.
- 파싱 실패한 줄도 계속 진행하며 종료하지 않으므로, 대상 파일이 JSON 로그가 아니어도 에러만 반복 출력하며 끝까지 실행된다.
