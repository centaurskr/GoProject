# goMatching (LocalLog)

코인별 주문 매칭(체결) 엔진 — [centolib/logout](../../../centolib/logout/README.md)(로컬 콘솔/파일 로그)을 사용하는 버전. 기능은 NetLog 버전과 동일하며, 전체 아키텍처·환경변수·주의사항은 [../NetLog/README.md](../NetLog/README.md)를 참고할 것. 이 문서는 LocalLog 버전과의 차이만 다룬다.

## NetLog 버전과의 차이

- 로그 백엔드만 `centolib/netlog` 대신 `centolib/logout`을 사용한다(다른 로직은 동일 — `afterfunc.go`/`fdsfunc.go`/`memefunc.go`/`sisefunc.go`의 import 한 줄 차이).
- **CLI 인자가 1개 더 필요하다**(로그 출력 모드 지정):

  ```bash
  goMatching <Index> <mode>
  # ex) goMatching 1 console
  ```

  | mode | 동작 |
  |---|---|
  | `file` | `$LOG_HOME`(미설정 시 `/tmp/Log`) 아래 파일로 로그 기록 |
  | `console` | 표준출력으로 로그 출력 |
  | `all` | 파일+콘솔 동시 출력 |
  | (그 외/누락) | `console`으로 동작 |

## 빌드

```bash
cd goPlus/Matching/LocalLog
make            # ./goMatching 생성
make install    # 빌드 + ~/bin 으로 이동
```

## 환경설정

NetLog 버전과 동일(`PROXYENV_YAML`, `ORDER_LIST`, `MATCHING_HOST`, `MATCHING_STREAM`, `ORDPATH`)하며, 추가로 `LOG_HOME`(file/all 모드일 때 로그 디렉토리, 기본 `/tmp/Log`)을 사용한다.

## 주의사항

NetLog 버전과 동일한 주의사항(goMatchingCtl 연결 끊김 시 자동 PAUSE, 인덱스 1부터 시작, mmap 기반 order book, 무중단 코인 추가/제거)이 그대로 적용된다. Vector 소켓 의존성이 없으므로 독립/개발 환경에서 사용하기 적합하다.
