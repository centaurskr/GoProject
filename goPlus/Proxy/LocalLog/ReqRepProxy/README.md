# goReqRepProxy (LocalLog)

ZeroMQ REQ/REP 메시지 중계 프록시(ROUTER/DEALER, capture 지원) — [centolib/logout](../../../../centolib/logout/README.md)(로컬 콘솔/파일 로그) 버전. 전체 아키텍처·제어 프로토콜·환경설정·주의사항은 [../../NetLog/ReqRepProxy/README.md](../../NetLog/ReqRepProxy/README.md) 참고.

## NetLog 버전과의 차이

- 로그 백엔드가 `centolib/logout`이다.
- **CLI 인자가 1개 더 필요**하다:

  ```bash
  goReqRepProxy <SERVERNAME> <debug mode>
  ```
  `<debug mode>`: `file`/`console`/`all`.

## 빌드

```bash
cd goPlus/Proxy/LocalLog/ReqRepProxy
make
make install
```

## 주의사항

NetLog 버전과 동일 — capture는 구독자가 붙어야 전송, PAUSE 시 REQ 클라이언트가 응답 타임아웃될 수 있음.
