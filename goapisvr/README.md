# goapisvr

`centolib/fulcrum`을 이용해 Bitcoin 잔액 조회를 WebSocket으로 제공하는 소형 API 서버.

## 개요

- `module goapisvr`, `go 1.22.2`. 단일 파일(`goapisvr.go`) 실행 프로그램. `github.com/gorilla/websocket` + [centolib/fulcrum](../centolib/fulcrum/README.md).
- 엔드포인트는 `/get_balance` 하나뿐이며, WebSocket 연결을 맺으면 클라이언트가 보내는 메시지마다 `{METHOD, PARAM}` JSON을 받아 `PARAM`(Bitcoin 주소)의 잔액을 Fulcrum 서버에서 조회해 응답한다.

## 빌드

```bash
cd goapisvr
go build -o goapisvr
# 또는
make          # ./goapisvr 생성
make install  # 빌드 + strip + ~/bin 으로 이동
```

## 사용법

```bash
./goapisvr
```

인자 없이 실행하면 `:50005` 포트에서 HTTP 서버가 뜨고 `ws://<host>:50005/get_balance`로 WebSocket 연결을 받는다.

요청 메시지(JSON):
```json
{"Method": "get_balance", "Param": "bc1q..."}
```

응답 메시지(JSON, `Param`에 대한 `fulcrum.GetBalance` 결과를 다시 JSON 문자열로 감싼 값):
```json
"{\"Confirmed\":..,\"Unconfirmed\":..}"
```

## 환경설정

- 리슨 포트: `:50005` (코드에 하드코딩, 변경하려면 소스 수정 필요).
- Fulcrum 서버 주소: `192.168.0.102:50001` (코드에 하드코딩).
- 그 외 환경변수/설정 파일 없음.

## 주의사항

- **CORS 검사가 비활성화**되어 있다(`CheckOrigin: func(r *http.Request) bool { return true }`) — 모든 Origin의 WebSocket 연결을 허용하므로, 신뢰할 수 없는 네트워크에 노출하면 임의 웹페이지에서 이 API를 호출할 수 있다.
- WebSocket Upgrade 실패 시 `log.Fatalf`를 호출한다 — **프로세스 전체가 종료**된다(연결 1건의 업그레이드 실패가 서버 전체를 내림). 운영 환경에서는 위험한 패턴이다.
- Fulcrum 서버 주소와 리슨 포트가 소스에 하드코딩되어 있어, 배포 환경이 바뀌면 재빌드가 필요하다.
- 응답 페이로드는 JSON 객체가 아니라 **JSON 문자열로 이중 인코딩**된 값이다(`json.Marshal(outstr)`에서 `outstr`이 이미 JSON 문자열이므로) — 클라이언트는 수신 후 한 번 더 언마샬해야 실제 객체를 얻는다.
