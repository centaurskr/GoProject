# RegNumInit (goRegNumInit, goReadRegnum)

주문 접수번호(등록번호) 맵 파일을 생성/조회하는 두 개의 독립 운영 도구. 이 디렉토리는 서로 다른 `main()`을 가진 두 소스 파일(`RegNumInit.go`, `ReadRegnum.go`)로 구성되며, Makefile이 각각 별도 바이너리로 빌드한다.

## 개요

접수번호 포맷(공통 주석): `yymmdd + line no(2자리) + seq(11자리, 0으로 채움)` — 총 uint64. 맵 파일은 `MapHeader` + `RegNumber{Index int, RegNum uint64}` 배열(`Mb.MAX_LINE_COUNT`개) 구조이며, 파일 경로는 `${PLATFORM_ENV}` 설정 파일의 `REGNUM_MAP` 키로 지정한다.

- **goRegNumInit** (`RegNumInit.go`): 접수번호 맵 파일을 새로 생성(`api.CreateMapfile`)하고, 각 라인 인덱스별 초기 시퀀스 값을 채운다.
- **goReadRegnum** (`ReadRegnum.go`): 기존 접수번호 맵 파일을 attach해서 전체 인덱스/접수번호를 조회·출력한다.

## 빌드

```bash
cd goPlus/Ops/RegNumInit
make            # goReadRegnum, goRegNumInit 둘 다 빌드
make install    # 빌드 + ~/bin 으로 이동 (strip 없음)
```

## 사용법

```
goRegNumInit            # 최초 생성 — ${PLATFORM_ENV} 파일이 이미 있으면 거부하고 종료
goRegNumInit force       # 강제 재생성

goReadRegnum             # 인자 없음, 기존 맵 파일 내용을 전부 출력
```

`goRegNumInit`은 인자가 없거나 `force` 하나일 때만 동작하며, 그 외(인자 2개 이상 또는 `force`가 아닌 값)는 사용법 출력 후 종료한다.

## 환경설정

- `PLATFORM_ENV` 환경변수(필수, 두 도구 공통): 플랫폼 환경 설정 파일 경로.
- 해당 환경 파일의 `REGNUM_MAP` 키: 접수번호 맵 파일 경로.
- `centolib/logout` 콘솔 로깅 모드로 초기화됨.

## 주의사항

- `goRegNumInit`의 "이미 존재함" 검사는 `REGNUM_MAP` 맵 파일이 아니라 **`PLATFORM_ENV`가 가리키는 파일 자체**를 `os.Open`으로 열어 성공 여부로 판단한다(인자 없이 실행 시). 실제 맵 파일 존재 여부와는 무관한 로직이므로 의도와 다르게 동작할 수 있다 — 강제 재생성이 필요하면 `force` 인자를 사용한다.
- `force` 실행은 기존 접수번호 맵 파일을 **덮어쓴다** — 운영 중 실행 시 이미 발급된 접수번호 상태가 초기화될 수 있으므로 반드시 서비스 중단 상태에서만 실행해야 한다.
- 두 도구 모두 에러 발생 시 `os.Exit(0)`으로 종료한다(비정상 상황에서도 exit code 0).
