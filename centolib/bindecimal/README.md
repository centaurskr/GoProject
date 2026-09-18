# centolib/bindecimal

파일/mmap에 고정 길이 바이너리로 저장하기 위한 `decimal.Decimal` 래퍼 타입.

## Import

```go
import Bd "centolib/bindecimal"
```

## 개요

`shopspring/decimal` 값은 가변 길이라 mmap 구조체 필드로 직접 쓰기 어렵다. `BinDecimal`은 36바이트 고정 버퍼(`Val [36]byte` + `ValLen int`)에 `Decimal.MarshalBinary()` 결과를 담아, 매칭엔진의 주문/호가 구조체(`centolib/matchingboard`)처럼 고정 크기 레코드가 필요한 곳에서 소수 값을 다룰 수 있게 한다.

## API

### 생성

- `func New() BinDecimal` — 값 `"0"`으로 초기화된 `BinDecimal`.
- `func NewFromString(num string) (BinDecimal, int)` — 문자열로부터 생성. 반환 `int`는 에러코드(`0`=성공, `-1`=36바이트 초과).

### 값 추출/설정

- `func (p *BinDecimal) Decimal() Dc.Decimal` — 내부 바이너리를 `decimal.Decimal`로 언마샬.
- `func (p *BinDecimal) String() string` — 위와 동일하되 문자열로.
- `func (p *BinDecimal) SetFromString(num string) int` — 문자열 파싱 후 저장. `-1`이면 36바이트 초과.
- `func (p *BinDecimal) SetFromDecimal(num Dc.Decimal) int` — `Decimal` 값 직접 저장.
- `func (p *BinDecimal) Set(a interface{}) int` — `BinDecimal | Decimal | string | float32 | float64 | int64` 아무 타입이나 받아 저장(내부적으로 `getDecimal`로 타입 스위치).

### 비교/판정

- `func DecimalIsZero(a interface{}) bool`
- `func DecimalEqual(a, b interface{}) int` — `1`=같음, `0`=다름, `-1`/`-2`=a/b 타입 변환 오류.
- `func DecimalCompare(a, b interface{}) int` — `1`=a>b, `0`=a==b, `-1`=a<b.
- `func DecimalMax(a, b interface{}) interface{}` — 둘 중 큰 값을 원래 타입 그대로 반환.

### 사칙연산

각 연산자(`Add`/`Sub`/`Mul`/`Div`)마다 4가지 출력 변형이 있다 — 반환 타입만 다르고 동작은 동일:

| 연산 | `BinDecimal` 반환 | 반환값 검증용 별칭 | `string` 반환 | `decimal.Decimal` 반환 |
|---|---|---|---|---|
| 더하기 | `Add(val, mode, pl) (BinDecimal, int)` | `AddOutBinDec(...)` | `AddOutString(...)` | `AddOutDecimal(...)` |
| 빼기 | `Sub(...)` | `SubOutBinDec(...)` | `SubOutString(...)` | `SubOutDecimal(...)` |
| 곱하기 | `Mul(...)` | `MulOutBinDec(...)` | `MulOutString(...)` | `MulOutDecimal(...)` |
| 나누기 | `Div(...)` | `DivOutBinDec(...)` | `DivOutString(...)` | `DivOutDecimal(...)` |

모두 시그니처는 `(p *BinDecimal) Xxx(val interface{}, mode, pl int) (반환타입, int)`이며, `p`(수신자) 값에 `val`을 연산한다.
- `val`: 피연산자 (`BinDecimal | Decimal | string | float32/64 | int64`).
- `mode`: 반올림 모드 상수 — `ROUND_UP`(올림), `ROUND_DN`(버림), `ROUND`(반올림), `NOROUND`(반올림 없음).
- `pl`: 반올림 자릿수(place).
- 반환 `int`: `0`=성공, 음수=에러(`val` 타입 변환 실패 등).

나머지: `func (p *BinDecimal) Mod(val interface{}) (BinDecimal, int)` — 나머지 연산(반올림 없음).

## 의존성

`github.com/shopspring/decimal`.

## 주의사항

- `Val [36]byte`를 초과하는 정밀도/자릿수의 decimal은 `SetFromString`/`SetFromDecimal`/`Set`이 `-1`(또는 `-3`)을 반환하며 저장되지 않는다 — 반환 에러코드를 반드시 확인해야 한다.
- 사칙연산 계열 함수들은 에러 시 `string` 변형은 `"0"`을, `BinDecimal`/`Decimal` 변형은 빈 값을 반환하므로, 실패를 값으로 오인하지 않도록 항상 두 번째 반환값(에러코드)을 확인할 것.
