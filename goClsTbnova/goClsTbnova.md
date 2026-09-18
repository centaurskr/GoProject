# goClsTbnova

goTbnova 관리 프로세스 정리 유틸리티

---

## 개요

`goClsTbnova` 는 `goTbnova` 종료 후 독립 실행 중인 관리 대상 프로세스를  
탐지하고 종료하는 유틸리티이다.

### 배경

`goTbnova` 는 `Setsid=true` 로 자식 프로세스를 독립 세션으로 기동한다.  
`goTbnova` 가 종료되면 자식들은 `ppid=1 (init)` 하위로 재배치되어 계속 실행된다.  
이 유틸리티는 그 프로세스들을 `process.toml` 기준으로 탐지하여 정리한다.

```
goTbnova 종료
    │
    └── 자식 프로세스들의 ppid → 1 (init)
                │
                └── goClsTbnova 로 탐지 및 종료
```

### C버전 clschaos 와의 차이

| 항목 | clschaos (C버전) | goClsTbnova |
|---|---|---|
| 설정 파일 | 콜론 구분 tbl | TOML |
| OS | Solaris (psinfo_t) | Ubuntu/Linux (/proc) |
| 프로세스 탐지 | 프로세스명 (pr_fname) | exe 절대경로 + basename |
| dry-run | 없음 | `-n` 옵션 지원 |
| 확인 프롬프트 | 없음 | 기본 제공 (`-f` 로 생략) |
| 시그널 선택 | SIGKILL 고정 | `-s` 로 SIGTERM 선택 가능 |
| 상세 출력 | 없음 | `-v` 옵션 지원 |

---

## 빌드

```bash
# goTbnova 프로젝트 루트에서
$ cd goClsTbnova
$ go build -o goClsTbnova

# 설치 디렉토리로 직접 빌드
$ go build -o /home/cento/bin/goClsTbnova
```

> **참고**: `go.mod` 는 상위 goTbnova 프로젝트의 것을 공유한다.

---

## 사용법

```
goClsTbnova [options] <table_path>
```

### 옵션

| 옵션 | 기본값 | 설명 |
|---|---|---|
| `-u <user>` | 현재 유저 | 대상 유저명 |
| `-n` | false | dry-run: 목록만 출력, 종료 안함 |
| `-f` | false | 확인 없이 즉시 종료 |
| `-s` | false | SIGKILL 대신 SIGTERM 발송 |
| `-v` | false | 상세 출력 (cmdline 전체 표시) |

### 인자

| 인자 | 설명 |
|---|---|
| `table_path` | `process.toml` 전체 경로 (필수, 절대경로) |

---

## 사용 예시

### 1. 목록만 확인 (dry-run)

실제 종료 없이 탐지 결과만 출력한다. 운영 전 반드시 확인 권장.

```bash
$ ./goClsTbnova -n /home/cento/conf/process.toml

탐지된 프로세스 3개 (ppid=1, goTbnova 관리 대상):
────────────────────────────────────────────────────────────────────
  PID       ID                EXE
────────────────────────────────────────────────────────────────────
  12345     TBAGENT           /home/cento/bin/tb_agent
  12346     NPX-QP-ORDRCV     /home/cento/bin/goReqRepProxy
  12347     NMATCHINCTL       /home/cento/bin/goMatchingCtl
────────────────────────────────────────────────────────────────────

[dry-run] 실제 종료하지 않았습니다. -f 옵션으로 종료하세요.
```

### 2. 확인 후 종료 (기본 동작)

```bash
$ ./goClsTbnova /home/cento/conf/process.toml

탐지된 프로세스 3개 (ppid=1, goTbnova 관리 대상):
────────────────────────────────────────────────────────────────────
  PID       ID                EXE
────────────────────────────────────────────────────────────────────
  12345     TBAGENT           /home/cento/bin/tb_agent
  12346     NPX-QP-ORDRCV     /home/cento/bin/goReqRepProxy
  12347     NMATCHINCTL       /home/cento/bin/goMatchingCtl
────────────────────────────────────────────────────────────────────

위 3개 프로세스에 SIGKILL 를 발송합니다. 계속하시겠습니까? [y/N] y
  [ OK ] PID 12345     ID=TBAGENT           SIGKILL 발송
  [ OK ] PID 12346     ID=NPX-QP-ORDRCV     SIGKILL 발송
  [ OK ] PID 12347     ID=NMATCHINCTL       SIGKILL 발송

완료: 성공 3 / 실패 0
```

### 3. 강제 종료 (확인 없음)

```bash
$ ./goClsTbnova -f /home/cento/conf/process.toml
```

### 4. SIGTERM 으로 종료 (graceful)

```bash
$ ./goClsTbnova -s /home/cento/conf/process.toml
```

### 5. 특정 유저 지정

```bash
$ sudo ./goClsTbnova -u cento /home/cento/conf/process.toml
```

### 6. 상세 출력

```bash
$ ./goClsTbnova -v -n /home/cento/conf/process.toml

탐지된 프로세스 3개 (ppid=1, goTbnova 관리 대상):
────────────────────────────────────────────────────────────────────
  PID       ID                EXE
────────────────────────────────────────────────────────────────────
  12345     TBAGENT           /home/cento/bin/tb_agent
            cmdline: /home/cento/bin/tb_agent TugBoat
  12346     NPX-QP-ORDRCV     /home/cento/bin/goReqRepProxy
            cmdline: /home/cento/bin/goReqRepProxy ORDERPASS
  12347     NMATCHINCTL       /home/cento/bin/goMatchingCtl
            cmdline: /home/cento/bin/goMatchingCtl
────────────────────────────────────────────────────────────────────

[dry-run] 실제 종료하지 않았습니다. -f 옵션으로 종료하세요.
```

---

## 탐지 기준

다음 세 조건을 **모두** 충족하는 프로세스를 탐지한다.

| 조건 | 내용 | 확인 위치 |
|---|---|---|
| 1 | `ppid == 1` | `/proc/{pid}/status` 의 `PPid` 필드 |
| 2 | `uid == 대상 유저 UID` | `/proc/{pid}/status` 의 `Uid` 필드 |
| 3 | 실행 파일이 `process.toml` 의 `path` 와 일치 | `/proc/{pid}/exe` 심볼릭링크 |

### 경로 비교 우선순위

```
1순위: /proc/{pid}/exe 의 절대경로 == process.toml path 의 절대경로
2순위: /proc/{pid}/exe 의 basename == process.toml path 의 basename
```

---

## 운영 시나리오

### goTbnova 재기동 전 정리

```bash
# 1. 실행 중인 관리 프로세스 확인
$ goClsTbnova -n /home/cento/conf/process.toml

# 2. 전체 종료
$ goClsTbnova -f /home/cento/conf/process.toml

# 3. goTbnova 재기동
$ systemctl start goTbnova
```

### 비상 전체 정리 (스크립트)

```bash
#!/bin/bash
# emergency_cls.sh — 비상 시 전체 관리 프로세스 정리

TABLE=/home/cento/conf/process.toml

echo "=== goClsTbnova 비상 정리 ==="
/home/cento/bin/goClsTbnova -n "${TABLE}"
echo ""
read -p "위 프로세스를 모두 종료합니까? [y/N] " ans
if [ "${ans}" = "y" ]; then
    /home/cento/bin/goClsTbnova -f "${TABLE}"
fi
```

---

## 주의사항

> **⚠️ goTbnova 실행 중에는 사용하지 말 것**  
> goTbnova 가 실행 중인 상태에서 goClsTbnova 를 실행하면  
> 관리 대상 프로세스가 종료되고 goTbnova 가 즉시 재기동을 시도한다.  
> 반드시 goTbnova 종료 후 사용한다.

> **ℹ️ ppid=1 조건**  
> goTbnova 종료 후 자식들의 ppid 가 1 로 변경된다.  
> goTbnova 가 실행 중이면 자식들의 ppid 는 goTbnova 의 PID 이므로  
> goClsTbnova 의 탐지 대상에 포함되지 않는다.

---

## 파일 구조

```
goTbnova/
└── goClsTbnova/
    ├── main.go      ← 소스 파일
    └── README.md    ← 이 문서
```

---

## 버전

| 버전 | 날짜 | 내용 |
|---|---|---|
| 1.0.0 | 2026-04-27 | 최초 작성 |

---

*goClsTbnova — goTbnova 프로젝트의 일부*
