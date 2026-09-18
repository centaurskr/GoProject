# goMatchingCtl (LocalLog)

매칭 시스템 제어 프로세스 — [centolib/logout](../../../centolib/logout/README.md)(로컬 콘솔/파일 로그) 버전. 전체 아키텍처·환경변수·주의사항은 [../NetLog/README.md](../NetLog/README.md) 참고. 이 문서는 차이만 다룬다.

## NetLog 버전과의 차이

- 로그 백엔드가 `centolib/netlog` 대신 `centolib/logout`이다.
- **CLI 인자로 로그 출력 모드를 받는다**:

  ```bash
  goMatchingCtl <debug mode>
  # <debug mode>: file, console, all
  ```
  인자가 없으면 사용법(`Using()`)을 출력하고 종료한다(NetLog 버전은 인자를 받지 않는 것과의 주요 차이).

## 빌드

```bash
cd goPlus/MatchingCtl/LocalLog
make
make install
```

## 환경설정

NetLog 버전과 동일(`PROXYENV_YAML`, `MATCHING_STREAM`, `AFTERMAT_STREAM`, `ORDER_LIST`) + `LOG_HOME`(file/all 모드 로그 디렉토리).

## 주의사항

NetLog 버전과 동일 — 매칭 시스템의 중앙 조정자이며 단일 인스턴스만 허용(`IsRunning("MATCHINGCTL")`).
