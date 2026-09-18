# goAfterMatch (LocalLog)

매칭 후 DB 정리 프로세스 — [centolib/logout](../../../centolib/logout/README.md)(로컬 콘솔/파일 로그) 버전. 전체 아키텍처·환경변수·주의사항은 [../NetLog/README.md](../NetLog/README.md) 참고.

## NetLog 버전과의 차이

- 로그 백엔드가 `centolib/logout`이다.
- **CLI 인자가 1개 더 필요**하다:

  ```bash
  goAfterMatch <Index> <mode>
  ```
  `<mode>`: `file`/`console`/`all` (그 외/누락 시 `console`).

## 빌드

```bash
cd goPlus/AfterMatch/LocalLog
make
make install
```

## 환경설정

NetLog 버전과 동일(`PROXYENV_YAML`, `MATCHING_HOST`, `AFTERMAT_STREAM`, `DBCONFIG_YAML`) + `LOG_HOME`(file/all 모드).

## 주의사항

NetLog 버전과 동일(DB 연결 필수, 큐 가득 참 시 블로킹 재시도, 동적 토픽 재할당). `<Index> <mode>` 두 인자 모두 필수.
