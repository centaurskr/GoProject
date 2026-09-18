# goRecvOrder (LocalLog)

주문 수신 프로세스 — [centolib/logout](../../../centolib/logout/README.md)(로컬 콘솔/파일 로그) 버전. 전체 아키텍처·환경변수·주의사항은 [../NetLog/README.md](../NetLog/README.md) 참고.

## NetLog 버전과의 차이

- 로그 백엔드가 `centolib/logout`이다.
- **CLI 인자가 1개 더 필요**하다:

  ```bash
  goRecvOrder <index> <debug mode>
  ```
  `<index>`: 0~99, `<debug mode>`: `file`/`console`/`all`.

## 빌드

```bash
cd goPlus/RecvOrder/LocalLog
make
make install
```

## 환경설정

NetLog 버전과 동일(`PROXYENV_YAML`, `ORDERCOND_MAP`, `REGNUM_MAP`, `DBCONFIG_YAML`) + `LOG_HOME`(file/all 모드).

## 주의사항

NetLog 버전과 동일 — DB INSERT 실패 시 증거금 롤백 미구현(TODO), 인덱스 범위 0~99, 일변경 시 접수번호 초기화.
