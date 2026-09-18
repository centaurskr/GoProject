# goPlus/SQL — DB 스키마 문서

goPlus 매칭 시스템이 사용하는 PostgreSQL(TimescaleDB 확장 포함) 테이블 정의 모음.
DDL 파일만 존재하며 마이그레이션 도구나 실행 스크립트는 없다 — 각 `.sql` 파일을
DB에 직접 적용(`psql -f`)해서 사용한다.

## 테이블 목록

| 파일 | 테이블 | 설명 |
|---|---|---|
| `coininfor.sql` | `public.coininfor` | 코인/종목 마스터 정보 |
| `ordertbl.sql` | `public.ordertbl` | 주문 정보 |
| `tradetbl.sql` | `public.tradetbl` | 체결(거래) 정보 |
| `ticktbl.sql` | `public.ticktbl` | 틱(체결) 시계열 정보 — TimescaleDB 하이퍼테이블 |
| `wallet.sql` | `public.wallet` | 사용자 지갑/잔고 |
| `exrate.sql` | `public.exratetbl` | 환율 정보 |
| `errtbl.sql` | `public.errtbl` | 오류 코드 정의 |

## coininfor — 코인정보

| 컬럼 | 타입 | 설명 |
|---|---|---|
| `symbol` | varchar(20) | 코인 심볼 (UNIQUE) |
| `itemcode` | int4 | 종목코드 (PK) |
| `coincode` | int4 | Coin Code |
| `mkcode` | int4 | 시장코드 |

## ordertbl — 주문정보

PK: `(itemcode, ordnum)`. 주요 컬럼:

| 컬럼 | 타입 | 설명 |
|---|---|---|
| `recseq` | bigserial | 레코드 일련번호(자동) |
| `symbol` / `itemcode` | varchar/int4 | 종목명 / 종목코드 |
| `userid` | int4 | 사용자 ID |
| `status` | varchar(2) | 주문상태: `GO`(진행중) / `CC`(취소) / `CM`(체결완료) / `ER`(오류) |
| `ordnum` / `orgnum` / `regnum` | varchar(20) | 주문번호(`yyyymmdd`+9자리) / 원주문번호(취소 시) / 접수번호 |
| `ordtype` | bpchar(1) | `1`:지정가, `2`:시장가 |
| `buysell` | bpchar(1) | `B`:매수, `S`:매도 |
| `price` / `qty` / `remqty` / `cheqty` / `canqty` | numeric(25,8) | 주문가격 / 주문수량 / 잔량 / 체결수량 / 취소수량 |
| `margin` / `fee` | numeric(25,8) | 주문 증거금 / 체결 증거금 |
| `makerate` / `takerate` / `callunit` / `mintrade` | numeric | 수수료율(Maker/Taker), 호가단위, 최소주문금액 |
| `cancel` | bpchar(1) | `0`:정상, `1`:취소주문 |
| `errcode` | varchar(12) | `errtbl.errcode` 참조 (status=ER일 때) |
| `regtime` / `ordtime` | timestamp(6) | 접수시각 / 주문번호 채번 시각 |

인덱스: `recseq`, `userid`, `ordnum`, `(itemcode, regnum)`.

## tradetbl — 체결정보

PK: `(itemcode, ordnum, cheseq)`. 체결 1건당 1행(Maker/Taker 각각 별도 행,
`maketake` 컬럼으로 구분). 주요 컬럼: `pairkey`(체결 페어 키), `price`/`qty`(체결가/
체결수량), `margin`/`fee`/`chefee`/`feerate`(정산 관련), `cntnum`(상대 주문번호),
`chetime`(체결시각).

## ticktbl — 틱정보 (TimescaleDB 하이퍼테이블)

PK: `(itemcode, chetime, seq)` — TimescaleDB 하이퍼테이블 생성을 위해 PK에 시간
컬럼(`chetime`)이 포함되어야 한다. `create_hypertable`로 3일 단위 청크 생성,
`itemcode`로 세그먼트 압축, 7일 경과 청크는 자동 압축, 1년 경과 데이터는 자동 삭제
(`add_retention_policy`). 컬럼: `chetime`, `itemcode`, `symbol`, `side`(B/S),
`price`, `qty`, `seq`(자동증가).

## wallet — 사용자 지갑/잔고

PK: `(userid, coincode)`. 컬럼: `amt`(잔고), `hold`(증거금으로 묶인 금액),
`address`(지갑주소).

## exratetbl — 환율

PK: `(basetime, currency)`. `basetime`은 분 단위 기준시각(`yyyymmddhhmm`),
`exrate`는 환율, `timestr`은 갱신 시각.

## errtbl — 오류코드

PK: `errcode`. `class1`/`class2`(대/중분류) + `code`(순번)로 구성된 오류 코드 체계
(예: `COM-IF-2`, `ORD-MA-1`). `ordertbl.errcode`, 매칭 처리 오류 응답 등에서 참조된다.

## 주의사항

- 파일에 `INSERT` 시드 데이터가 포함된 경우가 있다: `errtbl.sql`은 실제 운영에
  필요한 오류코드 테이블(정상 데이터)이지만, `wallet.sql`의 `INSERT`(userid
  42280~42289, 각 10,000,000,000 잔고)는 테스트/개발용 더미 지갑 데이터로 보이며
  운영 DB에 그대로 적용하지 않도록 주의한다.
- 마이그레이션 이력 관리 도구가 없으므로, 스키마 변경 시 이 디렉토리의 `.sql`
  파일을 직접 갱신해 실제 DB 상태와 맞춰야 한다(자동 동기화 없음).
- `ticktbl.sql`은 TimescaleDB 확장이 설치된 PostgreSQL에서만 적용 가능하다
  (`create_hypertable`, `timescaledb.compress` 등 TimescaleDB 전용 구문 사용).
