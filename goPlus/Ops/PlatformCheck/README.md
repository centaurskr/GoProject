# goPlatformCheck

운영 상태 점검용 콘솔 메뉴 TUI. 화살표/Enter로 탐색하는 텍스트 메뉴에서 주문 상태, 프록시 정보, 코인별 주문 파일을 조회한다.

## 개요

파일 구성: `platformCheck.go`(메인 메뉴 루프, 화면 제어 공통 함수), `ordCond.go`(주문상태 메뉴 — `ORDERCOND_MAP` 덤프), `coinord.go`(코인별주문정보 메뉴 — `ORDER_LIST`의 전체 코인 목록 탐색), `proxyMenu.go`(PROXY 정보 메뉴 — `Proxyenv.yaml`의 모든 프록시 목록 + 선택한 프록시에 `INFOR` 제어 명령을 보내 실시간 상태 조회).

메뉴 구성:

| 메뉴 | 기능 |
|---|---|
| 주문상태 | 코인별 주문 가능 상태(`ORDERCOND_MAP`) 조회 |
| PROXY 정보 | `Proxyenv.yaml`의 프록시 목록 + 선택 시 REQ/REP로 실시간 `INFOR` 조회 |
| 코인별주문정보 | `ORDER_LIST`에 등록된 코인 목록 탐색 |
| MENU3/MENU4 | 플레이스홀더(빈 함수, 아무 동작 없음) |

## 빌드

```bash
cd goPlus/Ops/PlatformCheck
make
make install
```

## 사용법

```bash
./goPlatformCheck
```

키 조작: `↑↓←→` 메뉴 이동, `Enter` 선택, `Ctrl+L` 화면 재출력, `q`/`Q` 종료(하위 메뉴에서도 동일). 터미널 리사이즈(`SIGWINCH`)를 감지해 레이아웃을 재계산한다.

## 환경설정

| 키 | 용도 |
|---|---|
| `ORDERCOND_MAP` | 주문상태 메뉴가 조회하는 mmap 파일 |
| `ORDER_LIST` | 코인별주문정보 메뉴가 조회하는 코인 목록 파일 |
| `PROXYENV_YAML` | PROXY 정보 메뉴가 조회하는 프록시 정의 yaml |

## 주의사항

- `MENU3`/`MENU4`는 미구현 상태(빈 함수)로 메뉴에 노출만 되어 있다.
- [centolib/screen](../../../../centolib/screen/README.md) raw mode를 사용하므로, 비정상 종료 시 터미널이 원상복구되지 않을 수 있다.
