# goPubSubProxy (LocalLog)

ZeroMQ PUB/SUB 메시지 중계 프록시 — [centolib/logout](../../../../centolib/logout/README.md)(로컬 콘솔/파일 로그) 버전. 전체 아키텍처·제어 프로토콜·환경설정·주의사항은 [../../NetLog/PubSubProxy/README.md](../../NetLog/PubSubProxy/README.md) 참고.

## NetLog 버전과의 차이

- 로그 백엔드가 `centolib/logout`이다.
- **CLI 인자가 1개 더 필요**하다:

  ```bash
  goPubSubProxy <SERVERNAME> <debug mode>
  ```
  `<debug mode>`: `file`/`console`/`all`.

## 빌드

```bash
cd goPlus/Proxy/LocalLog/PubSubProxy
make
make install
```

## 주의사항

NetLog 버전과 동일 — Capture 기능 없음, 구독자 0명이면 전달 스킵, PAUSE 시 poll 중단.
