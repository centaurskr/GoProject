# ProxyCtl (goProxyCtl)

ZMQ REQ 소켓으로 지정한 프록시의 제어(control) 포트에 명령을 보내고 JSON 응답을 정리해 출력하는 운영용 CLI 도구.

## 개요

`PROXY_YAML`에 정의된 프록시 이름의 `control` 엔드포인트를 `centolib/api.GetProxyInfor`로 조회한 뒤, ZMQ REQ 소켓으로 명령 문자열을 전송하고 응답(JSON)을 `api.ProxyInfor` 구조체로 파싱해 보기 좋게(`MarshalIndent`) 출력한다.

## 빌드

```bash
cd goPlus/Ops/ProxyCtl
make            # goProxyCtl 빌드
make install    # 빌드 + strip + ~/bin 으로 이동
```

## 사용법

```
goProxyCtl <Proxyname> <cmd>
```

- `<Proxyname>`: `${PROXY_YAML}`에 정의된 PROXY 서버 이름
- `<cmd>`: 전송할 명령 — `INFOR`, `PAUSE`, `RUN`

인자가 2개가 아니면 사용법을 출력하고 종료한다.

## 환경설정

- `PROXY_YAML` 환경변수(필수): 프록시 접속 정보가 담긴 YAML 파일 경로. 없으면 즉시 종료.

## 주의사항

- REQ 소켓 연결/전송에 타임아웃이 설정되어 있지 않다 — 대상 프록시가 응답하지 않으면 `RecvMessage(0)`에서 무기한 블록될 수 있다.
- 응답 JSON 파싱 실패 시 에러 메시지 포맷 문자열에 `%v` 대신 `$v`로 오타가 있어(`"Err : Unmarshal() :$v\n"`, `"Err : MaeshalIndent() :$v\n"`) 실제 에러 값이 출력되지 않는다 — 디버깅 시 참고할 것.
- 파일 하단에 주석 처리된 JSON map 순회 코드가 남아있다(미사용, 삭제 후보).
