# getbal

`centolib/fulcrum`을 이용해 Bitcoin 주소 잔액을 조회하는 최소 예제/데모 CLI.

## 개요

하드코딩된 주소(`bc1qmnnq9vv4nhd35px2j7dkks5smvgu96a5sakf3v`)와 Fulcrum 서버(`192.168.0.102:50001`)에 대해 [centolib/fulcrum.GetBalance](../centolib/fulcrum/README.md)를 호출하고 결과를 표준출력에 찍는 단일 파일 프로그램이다. 파라미터화되어 있지 않아 실사용 도구라기보다는 `centolib/fulcrum` 사용 예제에 가깝다.

## 빌드

```bash
cd getbal
make          # ./getbal 생성
make install  # 빌드 + strip + ~/bin 으로 이동
```

## 사용법

```bash
./getbal
```

인자를 받지 않는다. 조회할 주소나 서버를 바꾸려면 `getbal.go`의 `address`/`fulcrumServer` 상수를 직접 수정해야 한다.

## 환경설정

없음(설정 파일/환경변수 미사용). `fulcrumServer`가 가리키는 호스트(`192.168.0.102:50001`)에 네트워크로 도달 가능해야 한다.

## 주의사항

- **이 디렉토리에는 `go.mod`가 없고, `go.work`의 `use` 목록에도 포함되어 있지 않다** — 현재 상태로는 `go build`/`make`가 모듈 해석에 실패할 가능성이 높다(주석 처리된 `"libs/fulcrum"` import 흔적으로 보아 과거 다른 모듈 구성에서 쓰던 코드로 추정). 빌드하려면 `go.mod`를 추가하고 `centolib`를 `replace` 지시자로 연결하거나, `go.work`의 `use`에 `./getbal`을 추가해야 한다.
- Bech32(`bc1...`) mainnet 주소만 지원(`centolib/fulcrum`의 제약, [해당 README](../centolib/fulcrum/README.md) 참고).
