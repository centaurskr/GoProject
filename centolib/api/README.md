# centolib/api

DB 접속, 환경설정 파일 읽기, mmap 파일, raw TCP 소켓, epoll 이벤트 루프, 구조체 정렬/탐색, 시간 포맷 등을 제공하는 범용 유틸리티 패키지.

## Import

```go
import "centolib/api"
```

## 개요

특정 도메인에 종속되지 않는 저수준 헬퍼 함수들을 모아둔 패키지이며, 파일별로 역할이 나뉜다: `dbapi.go`(PostgreSQL 접속), `getbytes.go`(구조체 ↔ byte 변환), `getenv.go`(KEY=VALUE 환경파일), `goevent.go`(epoll+zmq 통합 이벤트 루프), `isrun.go`(중복 실행 방지 flock), `maxcpu.go`(GOMAXPROCS), `mkmap.go`(mmap 파일), `network.go`(raw TCP socket), `proxyenv.go`(ZMQ 프록시 정보 yaml), `sortst.go`(unsafe 기반 구조체 슬라이스 정렬/탐색), `timelib.go`(날짜 포맷).

외부 의존성: `database/sql` + `github.com/lib/pq`(Postgres 드라이버), `gopkg.in/yaml.v3`, `github.com/pebbe/zmq4`.

## API

### DB 접속 (dbapi.go)

#### `func ConnectDb(name string) (*sql.DB, error)`
- 설명: `name`으로 지정한 YAML 파일(`DbConfig`: host/port/user/password/database)을 읽어 PostgreSQL(`sslmode=disable`)에 접속한다.
- 파라미터: `name` — DB 접속 정보 YAML 파일 경로.
- 반환값: `*sql.DB` 커넥션, 에러.

```go
db, err := api.ConnectDb("/etc/tb/db.yaml")
```

### 구조체 ↔ Byte (getbytes.go)

- `func CleanByte(input []byte) []byte` — byte 슬라이스에서 `0x00`을 제거.
- `func CleanString(input []byte) string` — `0x00`을 제거하고 string으로 변환 (고정 길이 `[N]byte` 필드 → 문자열 변환에 사용. 사용례: `str := api.CleanString(item[:])`).
- `func GetBytes[T any](v *T) []byte` — 임의 타입 포인터를 `unsafe`로 byte slice로 재해석(zero-copy).
- `func SizeOf[T any](v T) int` — `unsafe.Sizeof` 래퍼.
- `func GetTypeName(a interface{}) string` — `reflect.TypeOf(a).String()`.

### 환경파일 읽기 (getenv.go)

- `func GetEnvValue(name string, inkey string) (string, error)` — `name` 파일에서 `KEY=VALUE` 형식(줄 단위, `#` 주석 지원)으로 `inkey` 값을 찾는다. 없으면 에러.
- `func GetPlatformEnv(name string) (string, error)` — OS 환경변수 `PLATFORM_ENV`가 가리키는 파일에서 `name` 키 값을 읽는다 (`PLATFORM_ENV` 미설정 시 에러).

### 이벤트 루프 (goevent.go)

`GoEvent` 구조체: TCP fd는 Linux `epoll`, ZMQ 소켓은 `zmq.Poller`로 통합 감시하는 단일 스레드 이벤트 루프.

- `func (p *GoEvent) InitGoEvent() error` — epoll fd 및 zmq Poller 초기화.
- `func (p *GoEvent) AddZmqClient(socket *zmq.Socket, h func(*GoEvent, *zmq.Socket)) error` — ZMQ 소켓을 POLLIN 감시 대상으로 등록.
- `func (p *GoEvent) AddTcpClient(fd int, h func(*GoEvent, int)) error` — TCP fd를 epoll에 등록.
- `func (p *GoEvent) DelTcpClient(fd int) error` — epoll에서 제거 후 fd `Close`.
- `func (p *GoEvent) DelZmqClient(socket *zmq.Socket) error` — Poller에서 제거 후 소켓 `Close`.
- `func (p *GoEvent) GoEventDown()` — `Run` 플래그를 내려 루프 종료 요청.
- `func (p *GoEvent) GoEventLoop() error` — 블로킹 메인 루프. ZMQ는 10ms 폴, TCP는 `EpollWait` 10ms 타임아웃으로 순차 처리하며 등록된 핸들러를 호출한다.

### 중복 실행 방지 (isrun.go)

- `func IsRunning(name string) (bool, error)` — `/tmp/{name}.lock` 파일에 `flock(LOCK_EX|LOCK_NB)`을 시도한다. `true`면 이미 실행 중(락 실패), `false`면 락을 걸고 실행 계속. `name`은 호스트 내에서 유일해야 한다.

### CPU (maxcpu.go)

- `func UsingMaxProcess()` — `runtime.GOMAXPROCS(runtime.NumCPU())` 호출.

### mmap 파일 (mkmap.go)

- `func CreateMapfile(fname string, size int) ([]byte, error)` — `fname`을 `size` 바이트로 생성/truncate하고 `PROT_READ|WRITE`, `MAP_SHARED`로 mmap.
- `func AttachMapfile(fname string) ([]byte, int, error)` — 기존 파일을 열어 현재 파일 크기만큼 mmap. 반환값은 (데이터, 파일크기, error).
- `func UnMapfile(data []byte)` — `syscall.Munmap`.

### Raw 네트워크 (network.go)

- `func GetMyIpv4Address() (string, error)` — 루프백이 아닌 첫 IPv4 주소.
- `func WriteDataStr4(fd int, data string) (int, error)` — 앞에 4자리 문자열 길이 헤더를 붙여 전송(`syscall.Write`). 반환 size는 헤더 4바이트 제외.
- `func ReadDataStr4(fd int) (int, []byte, error)` — 4바이트 길이 헤더 + 본문을 읽는다. 헤더/본문 각각 읽기 실패 시 -1/-2.
- `func WaitConnectRaw(listenfd int) (int, string, error)` — raw `syscall.Accept`, `peer`는 `"ip"` 문자열.
- `func OpenTcpServerRaw(port string) (int, error)` — `SO_REUSEADDR`+`SO_KEEPALIVE`를 설정한 raw listen 소켓 fd 생성. `port`는 `":32767"`처럼 콜론 포함 형식.
- `func WaitConnect(listener *net.TCPListener) (*net.TCPConn, string, error)` — 표준 `net.TCPListener.Accept()` 래퍼.
- `func OpenTcpServer(port string) (*net.TCPListener, error)` — raw 소켓으로 bind/listen한 뒤 `net.FileListener`로 감싸 `*net.TCPListener`를 반환(REUSEADDR/KEEPALIVE 옵션이 필요할 때 표준 `net.Listen` 대신 사용).
- `func OpenTcpClient(endpoint string) (int, error)` — raw 논블로킹 커넥트(`EINPROGRESS` 허용) 후 소켓 fd 반환.

### ZMQ 프록시 환경정보 (proxyenv.go)

`ProxyInfor`/`ProxyYaml`: ZMQ 프록시(PUB/SUB, REQ/REP)의 front/back/control/capture 엔드포인트 정보.

- `func ReadYaml(yamlfile, name string) (*ProxyYaml, error)` — YAML 파일에서 최상위 키 `name` 항목을 읽는다.
- `func NewProxyInfor(file, name string) (*ProxyInfor, error)` — `ReadYaml` 결과에 현재 호스트 IP·시작시각·카운터 초기값을 채운 런타임 `ProxyInfor`를 생성.
- `func GetProxyInfor(file, name, key string) (string, error)` — `key`(`"id"|"comment"|"proxytype"|"front"|"back"|"control"|"capture"`)에 해당하는 값만 문자열로 조회.
- `func GetProxyList(yamlFile string) ([]string, error)` — YAML 최상위 키(프록시 이름) 목록을 순서대로 반환.

### 구조체 정렬/탐색 (sortst.go)

`SortSlice`: `unsafe.Pointer` 기반으로 임의의 고정 크기 구조체 배열(주로 mmap 버퍼)을 `sort.Interface`로 감싸 정렬·이진탐색하기 위한 타입.

```go
type SortSlice struct {
    StBase   uintptr // 배열 시작 주소
    StCount  int     // 원소 개수
    StSize   uintptr // 원소 크기
    LessFunc func(unsafe.Pointer, unsafe.Pointer) bool
}
```

- `Len()/Less(i,j)/Swap(i,j)` — `sort.Interface` 구현 (`sort.Sort(&sortslice)`로 사용).
- `func (ss *SortSlice) Search(key unsafe.Pointer) unsafe.Pointer` — `LessFunc` 기준 이진탐색으로 동일 키 원소 포인터 반환(없으면 `nil`).
- `func (ss *SortSlice) SearchMode(key unsafe.Pointer, mode SearchMode) unsafe.Pointer` — `SMODE_EQ/GT/LT/GE/LE` 모드로 범위 탐색.

### 시간 포맷 (timelib.go)

- `func GetNumYyyymmdd() int` / `func GetStrYyyymmdd() string` — 오늘 날짜 `yyyymmdd` (정수/문자열).
- `func GetNumYymmdd() int` / `func GetStrYymmdd() string` — 오늘 날짜 `yymmdd` (정수/문자열).

## 의존성

`gopkg.in/yaml.v3`, `github.com/lib/pq`, `github.com/pebbe/zmq4`. `network.go`/`mkmap.go`/`goevent.go`는 Linux `syscall` API(epoll, raw socket, mmap)에 의존하므로 Linux 전용이다.

## 주의사항

- `GetBytes`, `SortSlice`, `mkmap.go` 계열은 `unsafe.Pointer` 산술을 직접 사용한다 — 대상 구조체의 메모리 레이아웃(정렬/패딩)이 파일 포맷과 정확히 일치해야 한다.
- `ReadDataStr4`/`WriteDataStr4`는 길이 헤더가 정확히 4자리 숫자 문자열이라고 가정한다(길이 10000바이트 이상인 데이터는 포맷이 깨진다).
- `IsRunning`은 lock 파일을 정리(`unlock`/삭제)하지 않는다 — 프로세스가 살아있는 동안 fd를 들고 있는 것으로 락을 유지하는 방식이다.
