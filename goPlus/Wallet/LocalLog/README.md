# goWallet (LocalLog)

증거금(지갑) 정리 프로세스 — [centolib/logout](../../../centolib/logout/README.md)(로컬 콘솔/파일 로그) 버전. 전체 아키텍처·환경변수·주의사항은 [../NetLog/README.md](../NetLog/README.md) 참고.

## NetLog 버전과의 차이

- 로그 백엔드가 `centolib/logout`이다.
- **CLI 인자로 로그 모드를 받는다**:

  ```bash
  goWallet <mode>
  ```
  `<mode>`: `file`/`console`/`all` (인자 없으면 사용법 출력 후 종료).

## 빌드

```bash
cd goPlus/Wallet/LocalLog
make
make install
```

## 환경설정

NetLog 버전과 동일(`PROXYENV_YAML`, `DBCONFIG_YAML`) + `LOG_HOME`(file/all 모드).

## 주의사항

NetLog 버전과 동일 — 전체 구독(무필터), 단일 인스턴스 제한(`IsRunning("goWallet")`).
