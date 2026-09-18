# centolib/matchingboard

매칭엔진(호가/주문 체결)의 핵심 데이터 구조 — mmap 파일 기반 order book(호가창) 엔진.

## Import

```go
import "centolib/matchingboard"
```

## 개요

`goPlus/Matching`, `goPlus/MatchingCtl` 등이 사용하는 저수준 order book 라이브러리. 코인 하나(`MEMCOIN`)는 매도(`Medo`)/매수(`Mesu`) 호가 배열을 가지고, 각 호가(`MEMHOGA`)는 동일 가격대 주문들을 담은 mmap 파일(`OrderHeader` + `Order` 고정 배열)에 대응한다. 파일 구성: `dbfun.go`(Postgres에서 코인 메타정보 조회), `maporder.go`(주문/호가 헤더/JSON 페이로드 타입 정의), `matchingboardata.go`(`MEMCOIN`/`MEMHOGA` 런타임 로직, 코인 리스트 관리).

```
MEMCOIN (코인 1개, 예: BTC.USDT)
├── Medo []MEMHOGA   // 매도 호가 배열 (가격 오름차순)
└── Mesu []MEMHOGA   // 매수 호가 배열 (가격 내림차순)
       └── MEMHOGA (호가 1개 = 동일가 주문 묶음)
             └── mmap 파일 = OrderHeader + []Order (고정 버퍼, ORDER_COUNT씩 증가)
```

가격/수량은 [centolib/bindecimal](../bindecimal/README.md)의 `BinDecimal`로 mmap 파일에 고정 길이 저장되고, 메모리(런타임) 상에서는 `github.com/shopspring/decimal`로 다룬다.

## API

### 상수 (matchingboardata.go)

- `MAX_ORDTIME = 9999999999999999999`, `MIN_ORDTIME = 0`
- `ORDER_COUNT = 3` — 호가 mmap 파일 생성/확장 시 한 번에 늘리는 주문 슬롯 수(기본값).
- `WINDOW_COUNT = 2` — "윈도우"(mmap attach 상태) 안에 유지할 최상위 호가 개수. fd 수 제한 때문에 상위 N개 호가만 attach하고 나머지는 unmap 상태로 둔다.
- `ORDSEQ_MAX`/`ORDSEQ_MIN`(maporder.go) — 주문번호 상/하한(빈 슬롯 표시용).
- `MAX_LINE_COUNT = 100`(maporder.go) — 주문 수신 라인 최대 수.
- `MAX_PRICE`/`MIN_PRICE` (전역 var, `init()`에서 각각 `"9999999999999"`/`"-1.0"`로 설정) — 매도/매수 호가의 빈 가격 슬롯 표시값.

### 타입: `BuySell` (매수/매도 구분)
- `MEDO BuySell = 1` (매도), `MESU BuySell = 2` (매수). `func (bs BuySell) String() string` — `"MEDO"`/`"MESU"`.

### 코인 목록 관리 (텍스트 설정 파일, `ORD_N=coin1:coin2:...` 형식)

- `func AddNewCoin(fileName, newCoin string) error` — 가장 코인 수가 적은 `ORD_N` 라인에 `newCoin`을 추가한다(신규 상장 시 사용). 원본을 `.BACK`으로 백업 후 덮어쓴다. 이미 등록된 코인이면 에러.
- `func GetCoinListByIndex(fileName string, idx int) ([]string, error)` — `ORD_{idx}=...` 라인의 코인 목록만 반환.
- `func GetAllCoins(fileName string) ([]string, error)` — 모든 `ORD_*` 라인의 코인을 합쳐 반환.

### DB 조회 (dbfun.go)

- `func GetCoinInfor(dbctx *sql.DB, name string) (CoinInfor, error)` — `coininfor` 테이블에서 `symbol`로 `itemcode/coincode/mkcode`를 조회해 `CoinInfor`로 반환 (`goPlus/SQL/coininfor.sql` 스키마 참조).

### 주문/호가 데이터 타입 (maporder.go)

- `EXRATE` — 환율 정보(JSON: `TIME/CURRENCY/RATE`).
- `MapHeader` — mmap 파일 공통 헤더(생성일/버퍼수/데이터수, 현재 미사용 흔적).
- `RegNumber` — 접수번호(Index + RegNum).
- `OrderCondition` — 코인별 주문 가능 상태(`Group`, `Status`: 0=정지,1=가능,3=매도만가능).
- `OrderHeader` — 호가 mmap 파일의 헤더 레코드. `Price`(정렬키), `Symbol`, `BuySell`, `BufCount`, `OrderCount`, `SumQty`.
  - Setter: `SetPrice(v interface{}) int`, `SetSumQty(v interface{}) int`, `SetSymbol(val string) int`, `AddSumQty(a interface{}, mode, pl int)`, `SubSumQty(a interface{}, mode, pl int)`.
  - Getter: `GetDcPrice()/GetStrPrice()/GetDcSumQty()/GetStrSumQty()/GetSymbol() `.
  - `func (p *OrderHeader) Dump()` — 표준출력에 상태 덤프.
- `OrderJson` — 클라이언트로부터 수신하는 주문 JSON 전체 필드(TR/COIN/CHANNEL/MKIDX/MIDX/PRC/QTY/UUID 등 23개 필드).
- `Order` — 메모리/mmap 상의 개별 주문 레코드(정렬키: `OrdTime`). `UserId, OrdNum, BuySell, OrdType, OrdTime, RegNum, CheSeq, Price, Qty, RemQty, MakeRate, TakeRate, Margin, Fee, CallUnit, MinTrade, OrdPacket([1024]byte 원본 JSON)`.
  - Setter: `SetPrice/SetQty/SetRemQty/SetMakeRate/SetTakeRate/SetCallUnit/SetMinTrade(v interface{}) int`, `SetBuySell(bs string)`, `SetOrdType(bs string)`, `SetUserId(str string)`.
  - Getter: `IsZeroRemQty() bool`, `GetStrPrice/GetStrQty/GetStrRemQty/GetStrMakeRate/GetStrTakeRate/GetStrCallUnit() string`, `GetDcPrice/GetDcQty/GetDcRemQty/GetDcMakeRate/GetDcTakeRate/GetDcCallUnit() Dc.Decimal`.
  - `func (p *Order) Dump(idx int)` — 표준출력 덤프(빈 슬롯 `OrdNum`은 스킵).
- `InnerOrdPacket` — 내부통신용 취소여부+원주문번호+`Order`.
- `InnerChePacket` — 체결(Maker/Taker) 결과 패킷(체결가/수량/수수료/환율 등).
- `JsonChePacket` — `InnerChePacket`의 JSON 직렬화용 축약 필드명 버전.
- `func GetDecimal(val interface{}) (Dc.Decimal, int)` — `BinDecimal | Decimal | string`을 `Decimal`로 통일 변환(그 외 타입은 에러코드 `-1`).

### `MEMHOGA` (호가 1개, matchingboardata.go)

`Price, SumQty, OrderCount, IsWindow, MapName, MapPoint, Hd *OrderHeader, OrdList uintptr`

- `func (p *MEMHOGA) InWindow() error` — mmap attach.
- `func (p *MEMHOGA) OutWindow()` — mmap unmap(윈도우 밖으로).
- `func (p *MEMHOGA) IncOrderBuff() error` — 파일을 `ORDER_COUNT`개 슬롯만큼 늘리고 재mmap(attach 상태에서 호출).
- `func (p *MEMHOGA) InsertOrder(ord *Order) error` — 이 호가에 주문 1건 추가. 버퍼 부족 시 자동으로 `IncOrderBuff()`, `SumQty`/`OrderCount` 갱신. 호출 시 윈도우 밖이었으면 처리 후 다시 밖으로 되돌린다.
- `func (p *MEMHOGA) GetOrdPointerByIndex(index int) (*Order, error)` — 인덱스로 주문 포인터(zero-copy) 획득.
- `func (p *MEMHOGA) AddSumQty(a interface{}) error` / `SubSumQty(a interface{}) error` — 호가 총수량 증감(메모리+mmap 동시 갱신).
- `func (p *MEMHOGA) SetPrice(prc interface{})` — mmap 헤더 우선 설정 후 메모리 값 동기화(순서 중요, 주석 명시).
- `func (p *MEMHOGA) IncOrderCount()` / `DecOrderCount()` — 주문 건수 증감(메모리+mmap).
- `func (p *MEMHOGA) LoadMEMHOGA() error` — `MapName` 설정 후 헤더만 읽어 메모리에 로드(`IsWindow=false` 상태로).
- `func (p *MEMHOGA) SortOrder()` — `centolib/api.SortSlice`로 주문번호 오름차순 정렬.
- `func (p *MEMHOGA) Dump()` — 표준출력 덤프.

### `CoinInfor` / `OrderMasterInfor`

- `CoinInfor{Symbol [32]byte, ItemCode, CoinCode, MkCode int}` — `func (p *CoinInfor) GetSymbol() string`(우측 0바이트 트림), `Dump(dep int)`.
- `OrderMasterInfor{Coin, OrdNum, BfPrice, BfTime, MedoNameIdx, MesuNameIdx}` — 코인별 주문 마스터(mmap 파일 1개, `{coin}.MASTER`). `Dump(dep int)`.
- `func CreateOrderMasterInfor(path string, coin CoinInfor) error` — `{path}/{coin}/{coin}.MASTER` 신규 생성, 주문번호 초기값 `오늘날짜+0`.
- `func GetOrderMasterInfor(path, name string) ([]byte, error)` — 마스터 파일 attach, 심볼 불일치 시 에러.

### `MEMCOIN` (코인 1개 최상위 구조)

`MstP []byte, MstInfor *OrderMasterInfor, Path string, Medo []MEMHOGA, Mesu []MEMHOGA`

- `func (p *MEMCOIN) LoadMEMCOIN(path, name string) error` — **최상위 진입점.** 마스터 attach → `Medo`/`Mesu` 전체 호가 mmap 파일(`MEDO.*`/`MESU.*`) 로드+정렬 → 상위 `WINDOW_COUNT`개만 윈도우 안으로.
- `func (p *MEMCOIN) UnLoadMEMCOIN()` — 마스터 + 모든 윈도우 상태 호가 unmap.
- `func CreateHoga(path, name string, mode BuySell, idx int) error` — `{path}/{MEDO|MESU}.{idx:04d}` 신규 호가 mmap 파일 생성(빈 슬롯 가격은 매도=`MAX_PRICE`, 매수=`MIN_PRICE`로 채움).
- `func (p *MEMCOIN) NewHoga(prc Bd.BinDecimal, mode BuySell) (*MEMHOGA, error)` — 처리 중 해당 가격 호가가 없을 때 즉석 생성(파일명 인덱스는 `MstInfor.MedoNameIdx`/`MesuNameIdx`에서 자동 채번).
- `func (p *MEMCOIN) FindHoga(mode BuySell, prc interface{}) *MEMHOGA` — 가격으로 호가 포인터 검색(선형 탐색).
- `func (p *MEMCOIN) FindHogaIndex(mode BuySell, prc interface{}) int` — 위와 동일하되 인덱스(-1=없음).
- `func (p *MEMCOIN) FindEmptyHoga(mode BuySell) *MEMHOGA` — `SumQty==0 && OrderCount==0`인 빈 호가 슬롯 검색.
- `func (p *MEMCOIN) DeleteHogaByPrice(mode BuySell, prc interface{}) error` — 주문이 남아있으면 실패, 없으면 mmap 파일 삭제 + 배열에서 제거.
- `func (p *MEMCOIN) SetWindow(mode BuySell)` — 상위 `WINDOW_COUNT`개만 `InWindow()`, 나머지 `OutWindow()`.
- `func (p *MEMCOIN) SortHoga(mode BuySell)` — 매도는 오름차순, 매수는 내림차순 정렬.
- `func (p *MEMCOIN) InitOrderSeq()` — 주문번호를 `오늘날짜+0`으로 리셋.
- `func (p *MEMCOIN) GetOrderSeq() uint64` — 주문번호 채번(증가 후 반환).
- `func (p *MEMCOIN) GetSymbol() string`
- `func (p *MEMCOIN) Dump()` / `DumpHoga()` — 표준출력 덤프(호가 전체 목록 포함).

## 의존성

`centolib/api`(mmap, 구조체 정렬), `centolib/bindecimal`, `centolib/logout`, `github.com/shopspring/decimal`, `github.com/lib/pq`(dbfun.go).

## 주의사항

- **전부 Linux `unsafe`/mmap 기반**이다. `Order`/`OrderHeader`의 필드 순서·타입을 바꾸면 기존에 생성된 mmap 파일과 바이너리 호환이 깨진다.
- `MEMHOGA.InsertOrder`/`GetOrdPointerByIndex` 등은 **`MEMHOGA`가 attach(윈도우 안) 상태여야** 유효하다 — 윈도우 밖 상태에서 `Hd`/`MapPoint`는 `nil`이다.
- `matchingboardata.go`에는 `NewMEMCOIN`과 `SortOrderAscByOrdNum` 함수가 **주석(`/* ... */`)으로 통째로 비활성화**되어 있다 — 현재 패키지 API에는 포함되지 않으므로 호출할 수 없다(대신 `LoadMEMCOIN`이 사실상의 생성/로드 진입점 역할을 한다).
- `DeleteHogaByPrice`는 주문이 남아있는 호가는 삭제하지 않는다(에러 반환) — 반드시 주문 정리 후 호출.
- 동시성 보호(mutex 등)가 코드 내에 없다 — 호출자가 order book 갱신을 단일 스레드/고루틴으로 직렬화한다고 가정한다.
