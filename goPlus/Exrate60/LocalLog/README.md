# goExrate60 (LocalLog)

USD/KRW 환율 조회·배포 프로세스 — [centolib/logout](../../../centolib/logout/README.md)(로컬 콘솔/파일 로그) 버전. 전체 아키텍처·환경변수·주의사항은 [../NetLog/README.md](../NetLog/README.md) 참고.

## NetLog 버전과의 차이

- 로그 백엔드가 `centolib/logout`이다.
- **CLI 인자로 로그 모드를 받는다**:

  ```bash
  goExrate60 <debug mode>
  ```
  `<debug mode>`: `file`/`console`/`all` (인자 없으면 사용법 출력 후 종료).

## 빌드

```bash
cd goPlus/Exrate60/LocalLog
make
make install
```

## 환경설정

NetLog 버전과 동일 — `PROXY_YAML`(주의: `PROXYENV_YAML`이 아님), `PLATFORM_ENV`(→`DBCONFIG_YAML`) + `LOG_HOME`(file/all 모드).

## 주의사항

NetLog 버전과 동일 — 환경변수 접근 방식이 다른 컴포넌트와 다름(`os.Getenv` 직접 사용), API 키 하드코딩, 단일 인스턴스 제한(`IsRunning("goExchange60")`).
