# goTbnova

레거시 C 프로세스 관리자를 대체하는 Go 프로세스 슈퍼바이저 데몬. `process.toml`에 정의된 하위 프로세스들의 기동·종료·재기동·시간대 제어·종속관계를 관리한다.

## 개요

- `module goTbnova`, `go 1.24.5`. 실행 파일(`package main`, `gotbnova.go`)이며 `centolib`를 로컬 `replace`로 참조한다.
- C버전(`tb_nova`)을 대체하기 위해 작성되었으며, 모니터 TCP 프로토콜은 C버전과 호환(`/etc/services`의 `tbnova/tcp` 포트 공유 — 동시 기동 시 포트 바인딩 실패로 중복 실행을 방지)된다.
- 자식 프로세스는 `Setsid=true`로 독립 세션 기동되므로, goTbnova가 종료되어도 자식은 계속 실행된다(고아 프로세스 정리는 [goClsTbnova](../goClsTbnova/README.md) 담당).
- 내부 구조: [model](model/README.md)(타입/상태머신 정의) · [config](config/README.md)(TOML 로드/검증/리로드 diff) · [process](process/README.md)(생명주기 관리 핵심 루프) · [control](control/README.md)(TCP 모니터 서버) · [tlog](tlog/README.md)(로거).

## 빌드

```bash
cd goTbnova
go build -o goTbnova
```

운영 배포용 스크립트(`build.sh`): 서비스 중지 → 잔여 프로세스 정리(`goClsTbnova`) → 빌드 → 서비스 재시작까지 한 번에 수행한다.

```bash
cd goTbnova
./build.sh
```
```bash
# build.sh 내용
sudo systemctl stop gotbnova.service
~/bin/goClsTbnova -f ~/Project/env/process.toml
go build -o goTbnova && strip goTbnova && mv goTbnova ~/bin
sudo systemctl start gotbnova.service
```

## 사용법

```
goTbnova [options] <table_path>
```

`table_path`는 `process.toml`의 **절대경로**(필수). 상대경로나 인자 누락 시 usage 출력 후 종료한다.

| 옵션 | 설명 |
|---|---|
| `-t` | `process.toml` 검증만 수행 후 종료 (nginx `-t` 상당). 라인 번호별 OK/ERR 및 오류 원인 출력 |
| `-d` | `process.toml` 파싱 결과(환경변수 `$VAR` 치환 완료값 포함) 표준출력 덤프 후 종료 |
| `-te` | `gotbnova.toml`(환경 설정 파일) 검증 및 내용 출력 후 종료. `process.toml` 인자 없이 단독 실행 가능 |
| `-v` | 버전 정보 출력 |

```bash
export ENV_HOME=/home/cento/Project/env
goTbnova /home/cento/conf/process.toml        # 정상 기동
goTbnova -t  /home/cento/conf/process.toml    # process.toml 검증
goTbnova -d  /home/cento/conf/process.toml    # process.toml 덤프
goTbnova -te                                   # gotbnova.toml 검증
```

## 환경설정

### `$ENV_HOME/gotbnova.toml` (실행 환경 설정, 없으면 기본값)

탐색 순서: `$ENV_HOME/gotbnova.toml` → 실행 파일 디렉토리의 `gotbnova.toml` → (둘 다 없으면 기본값).

```toml
[server]
bind = "127.0.0.1"                     # 모니터 TCP 바인딩 IP (기본: 로컬 전용)

[log]
type = "net"                            # "net"(기본) 또는 "file"
vector_socket = "/tmp/VECTOR.UNIX.DGRAM"  # type=net 일 때 Vector UDS 소켓 경로
dir = ""                                 # type=file 일 때 로그 디렉토리 (없으면 실행파일경로/log)

[work]
dir = ""                                 # 작업 디렉토리 (없으면 현재 디렉토리 유지)
```

### `process.toml` (관리 대상 프로세스 테이블)

`[[process]]` 섹션마다 하나의 관리 대상을 정의한다. 필드는 [model.ProcessConfig](model/README.md)에 대응한다: `id, group, desc, mode(P/R/O), weekend, stmode(Q/S), st/et, path, args, env, depends_on, depends_healthy, [process.restart], [process.cascade]`. `[default]` 섹션의 `def_dir`(상대경로 기준), `def_cnt`(기본 재기동 횟수)와 `[env]`(전역 환경변수, `$VAR` 형식 상호 참조 가능)도 지원한다.

## 주의사항

- **`table_path`는 반드시 절대경로**여야 한다(상대경로 입력 시 즉시 에러 종료).
- 포트 바인딩 실패는 **C버전 `tb_nova` 또는 다른 goTbnova 인스턴스가 이미 실행 중**임을 의미한다 — 별도 포트 설정이 없고 `/etc/services`의 `tbnova/tcp`를 그대로 사용한다.
- 로그 모드가 `net`(기본값)인데 Vector UDS 소켓(`/tmp/VECTOR.UNIX.DGRAM`)이 없으면 **기동이 즉시 실패**한다 — 독립 환경에서는 `gotbnova.toml`에 `[log] type = "file"`을 설정하거나 Vector 에이전트를 먼저 띄워야 한다.
- goTbnova 종료 시 **관리 중이던 자식 프로세스에는 신호를 보내지 않는다**(설계 의도). 자식까지 정리하려면 [goClsTbnova](../goClsTbnova/README.md)를 사용하거나 개별 `stop` 명령을 TCP로 보낸다.
- **goTbnova 실행 중에 `goClsTbnova`를 실행하지 말 것** — goTbnova가 즉시 재기동을 시도해 충돌한다(반드시 goTbnova 종료 후 사용).
- 설정 리로드(`reload` 명령)는 실행 중인 프로세스의 `path/mode/st/et` 등 변경사항은 **즉시 반영하지 않는다**(ENV만 갱신, 재기동 없음) — 다음 자연 재기동 시점에 적용된다. 상세 규칙은 [process/manager.go](process/README.md)의 `handleReload` 참고.
