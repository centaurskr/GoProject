# goHogaRcv

터미널 화면에 실시간 현재가/호가창을 그려 보여주는 콘솔 모니터. `SISEJSON`/`HOGAJSON` 프록시를 구독해 지정 코인의 체결가와 매도/매수 호가 10단계를 ANSI 화면 제어로 출력한다.

## 개요

- [centolib/screen](../../../../centolib/screen/README.md)로 커서 이동/색상/화면 지우기 제어.
- `SISEJSON`(현재가, [goMatching](../../../Matching/NetLog/README.md)의 `gofSiseFunction`이 생성)과 `HOGAJSON`(호가, 동일)을 코인 심볼로 구독한다.
- 수신은 채널(버퍼 5000)로 비동기 처리하며, 화면 갱신 로직(`displayCurr`/`displayHoga`)이 별도 고루틴에서 소비한다.

## 빌드

```bash
cd goPlus/Ops/Hoga
make
make install
```

## 사용법

```bash
goHogaRcv <coin>
```

`<coin>`은 심볼(예: `BTC.USDT`). **터미널 크기가 폭 80·높이 36 미만이면 즉시 종료**한다(실제 검사 조건은 `x < 80 || y < 36`). 종료 시 출력되는 안내 문구는 "최소 Size 폭(80) 높이(40)"라고 표시하지만, 실제 강제되는 최소 높이는 36이다(안내 문구와 코드의 실제 조건이 다름 — 아래 주의사항 참고). 실행 중에는 raw 모드가 아니라 일반 출력이며 `Ctrl+C` 등으로 종료.

## 환경설정

| 키 | 용도 |
|---|---|
| `PROXYENV_YAML` | `SISEJSON`(back), `HOGAJSON`(back) 엔드포인트 조회 |

## 주의사항

- 실제 강제 조건(`x < 80 || y < 36`)과 종료 시 출력 문구("최소 Size 폭(80) 높이(40)")가 서로 다르다 — 코드 자체의 메시지/조건 불일치이며, 터미널 높이가 36~39 사이여도 정상 실행된다(문구만 40을 요구하는 것처럼 보임).
- 좌표가 화면 크기에 맞춰 하드코딩(`InitApplication()`)되어 있어, 실행 중 터미널 크기를 바꾸면(리사이즈 시그널 처리 없음) 레이아웃이 깨질 수 있다.
