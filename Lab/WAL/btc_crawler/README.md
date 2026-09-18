# BTC 웹 크롤러

Arkham Intel 웹사이트에서 BTC 최근 전송 내역을 수집하는 Go 프로그램입니다.

## 기능

- Arkham Intel 웹사이트 크롤링
- BTC 트랜잭션 데이터 추출
- JSON 형식으로 데이터 저장
- 콘솔에 결과 출력

## 설치 및 실행

### 1. 의존성 설치

```bash
cd btc_crawler
go mod tidy
```

### 2. 프로그램 실행

```bash
go run btc_web_crawler.go
```

### 3. 실행 파일 빌드

```bash
go build -o btc_crawler btc_web_crawler.go
./btc_crawler
```

## 출력 데이터

프로그램은 다음 정보를 수집합니다:

- **해시**: 트랜잭션 해시
- **From**: 발신 주소
- **To**: 수신 주소  
- **Amount**: 전송 수량 (BTC)
- **Timestamp**: 전송 시간
- **Block Number**: 블록 번호

## 저장 파일

결과는 `btc_transfers_YYYYMMDD_HHMMSS.json` 형식으로 저장됩니다.

## 주의사항

- 웹사이트 구조가 변경되면 크롤링이 실패할 수 있습니다
- 과도한 요청은 IP 차단의 원인이 될 수 있습니다
- 실제 운영 환경에서는 적절한 지연 시간을 추가하는 것이 좋습니다

## 트러블슈팅

### 데이터가 수집되지 않을 경우

1. 웹사이트 구조 변경 확인
2. 네트워크 연결 확인
3. 선택자 업데이트 필요

### 의존성 오류

```bash
go clean -modcache
go mod download
```

## 라이선스

MIT License
