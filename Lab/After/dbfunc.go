package main
import (
   "database/sql"
    _ "github.com/lib/pq"
    Log "centolib/logout"
	Mb "centolib/matchingboard"
	Bd "centolib/bindecimal"
	"centolib/api"
	"strconv"
    "time"
    "fmt"
)
/////////////////  주문 Table에 ERROR UPDATE    /////////////////////////////
/// 취소주문시 오류 Update
func DbUpdateError(dbctx *sql.DB, pkjson Mb.OrderJson)error{

	sqlstr := `
	UPDATE  ordertbl
	SET 
		status = 'ER',
		errcode = $1
	WHERE
		itemcode = $2
	AND
		ordnum = $3;`

	_, err := dbctx.Exec(sqlstr, 
				pkjson.State,
				pkjson.ItemCode,
				pkjson.OrdNum)
	if err != nil { 
		return fmt.Errorf("%v-%s-ECODE(%s) ItemCode(%s) OrdNum(%s)", 
			err, sqlstr, pkjson.State, pkjson.ItemCode, pkjson.OrdNum)
	}
	Log.Infor("주문 오류Update 성공OrdNum[%s]msg[%s]", pkjson.OrdNum, pkjson.Message)

	return nil

}
/////////////////  취소주문 처리/////////////    /////////////////////////////
// 1. 원주문에 취소 수량을 Update한다(canqty = pkjson.Qty, remqty = 0)
// 2. 원주문 상태를   "GO" --> "CC"로 변경 한다
// 3. 취소주문 상태를 "GO" --> "CM"으로 변경한다
// 4. 취소주문의 주문증거금과, 취소수량을 Update한다
//
// 취소주문의 수량 & 금액 정보는 "goMatching"에서 계산한다
// -. 원주문의 주문잔량 - 취소수량 
// pkjson.Qty 
// -.원주문의 주문증거금 잔량(증거금 해소용)
// pkjson.Margin 
// pkjson.Fee
// 
func DbCancelOrder(dbctx *sql.DB, pkjson Mb.OrderJson)error{
	// 원주문 Update (canqty, remqty, status, margin, fee)
	// 주문 증거금 Clear
	// 원주문의 margin 과 fee는 
	// pkjson.Margin 과 pkjson.Fee와 같다.
	// 따라서, "-" 연산보다 "0"으로 Set함.
	sqlstr := `
	UPDATE ordertbl
	SET canqty = $1,
		remqty = 0,
		margin = 0,
		fee    = 0,
		status = 'CC'
	WHERE
		itemcode = $2 
	AND
		ordnum = $3;`
	_,err :=dbctx.Exec(sqlstr, pkjson.Qty, pkjson.ItemCode, pkjson.OrgNum)
	if err != nil{return err;}
Log.Infor("1/2 취소주문. 원주문(%v) Update OK. 취소수량[%s]", pkjson.OrgNum,pkjson.Qty)
	// 취소주문 상태 Update
	sqlstr = `
	UPDATE ordertbl
	SET status = 'CM',
		qty    = $1,
		margin = $2,
		fee    = $3 
	WHERE
		itemcode = $4 
	AND
		ordnum = $5;`
	_,err =dbctx.Exec(sqlstr, 
		pkjson.Qty, pkjson.Margin, pkjson.Fee,
		pkjson.ItemCode, pkjson.OrdNum)

	if err != nil{return err;}
	Log.Infor("2/2 취소주문. 취소주문 Update 'CM' OK.")
	return nil
}
/////////////////  시장가주문 완료처리//////////////////////////////////////////
// 1. 시장가 주문 상태를   "GO" --> "CM"로 변경 한다
// 2. 매수(B) 
//    pkjson.Margin, pkjson.fee에 주문증거금
//    매도(S)
//    pkjson.Margin (주문 완료 후 남은 수량)
// ❗㉼ 시장가 주문 정보 taker.price , taker.qty
//          │   taker.price    │   taker.qty           
// ─────────┼──────────────────┼─────────────────
//  매수(B) │   주문가능 금액  │      0.0                
// ─────────┼──────────────────┼─────────────────
//  매도(S) │      0.0         │  주문가능 수량                
// ─────────┴──────────────────┴─────────────────
// 
func DbUpdateMkDone(dbctx *sql.DB, pkjson Mb.OrderJson)error{
	sqlstr := `
	UPDATE ordertbl
	SET remqty = 0,
		margin = $1,
		fee    = $2,
		status = 'CM'
	WHERE
		itemcode = $3 
	AND
		ordnum = $4;`
	_,err :=dbctx.Exec(sqlstr, pkjson.Margin, pkjson.Fee,
			pkjson.ItemCode, pkjson.OrdNum)
	if err != nil{return err;}

	return nil
}
/////////////////  주문 Table에 주문번호 Update /////////////////////////////
// ordertbl.ordnum  -- 주문번호
// ordertbl.ordtime -- 주문시각 (Microsecond)
// ordertbl.status  -- 주문 상태  RG->GO (접수상태 -> 매매진행중)
// 작업KEY : pk.ItemCode + pk.RegNum
// 빠르기  : SELECT > INSERT > UPDATE
//
func DbUpdateOrdNum(dbctx *sql.DB, pkjson Mb.OrderJson) error{

Log.Debug("DBSEQ : RegNum[%s] OrdNum[%s] time[%s]", 
	pkjson.RegNum, pkjson.OrdNum, pkjson.OrdTime)

	tm, _ := strconv.ParseInt(pkjson.OrdTime, 10, 64)
	ordtime := time.UnixMicro(tm)

	sqlstr := `
	UPDATE  ordertbl
	SET 
		ordnum = $1, 
		ordtime = $2,
		status  = 'GO'
	WHERE
		itemcode = $3
	AND
		regnum = $4;`

	_, err := dbctx.Exec(sqlstr, 
				pkjson.OrdNum,
				ordtime,
				pkjson.ItemCode,
				pkjson.RegNum)
	if err != nil { return err; }
	Log.Infor("주문번호 Update 성공OrdNum[%s]", pkjson.OrdNum)
	return nil
}
///////////////// 체결 내용을 주문 Table에 Update /////////////////////////////
// ordertbl.remqty     -- 체결 후 주문 잔량
// ordertbl.cheqty     -- 체결수량
// ordertbl.status     -- 잔량이 없는 경우  GO -> CM
// 작업 KEY : pk.ItemCode + pk.OrdNum
func DbUpdateOrder( dbctx *sql.DB, pk *Mb.InnerChePacket)error{
	// 체결수량, 주문상태 Update
	sqlstr := `UPDATE ordertbl
		SET 
			cheqty = cheqty + $1, 
			remqty = remqty - $2, 
			margin = margin - $3,
			fee    = fee    - $4,
			status = $5
		WHERE itemcode = $6 AND ordnum = $7;`
	var status string
	if Bd.DecimalIsZero(pk.RemQty) == true {
		status = "CM"
	}else{
		status = "GO"
	}
	_, err := dbctx.Exec(sqlstr, 
		pk.Qty.String(), pk.Qty.String(), 
		pk.Margin.String(), pk.Fee.String(), status,
		pk.ItemCode, pk.OrdNum)
	if err != nil {
		return fmt.Errorf("%v-%s-%+v", err, sqlstr, pk)
	}
	return nil
}
// 체결 Insert ///////////////////////////////////////////////////////////////
// -- 채결 Table : public.tradetbl
//  pairkey     varchar(20)    NULL,     -- 체결 Pair key
//  maketake    bpchar(1)      NULL,     -- Maker Taker (M,T)
//  symbol      varchar(20)    NULL,     -- 종목명
//  ordtype     bpchar(1)      NULL,     -- 주문구분 (L:지정가, M:시장가)
//  buysell     bpchar(1)      NULL,     -- 매도수 구분 (B:매수, S:매도)
//  price       numeric(25, 8) NULL,     -- 체결가격
//  qty         numeric(25, 8) NULL,     -- 체결수량
//  margin      numeric(25, 8) NULL,     -- 주문증거금(차감용)
//  fee         numeric(25, 8) NULL,     -- 체결증거금(차감용)
//  chefee      numeric(25, 8) NULL,     -- 체결 수수료
//  feerate     numeric(25, 8) NULL,     -- 체결수수료율(maker or taker)
//  ordnum      varchar(20)    NULL,     -- 주문번호 yyyymmdd + 9(13)
//  cheseq      int            NULL,     -- 체결번호
//  cntnum      varchar(20)    NULL,     -- 상대주문번호
//  chetime     timestamp(6)   NULL,     -- 체결시각
//  itemcode    int4           NULL,     -- 종목코드
//  coincode    int4           NULL,     -- Coin Code
//  mkcode      int4           NULL,     -- 시장코드(결제 통화)
//  userfield   varchar(60)    NULL,     -- 사용자정의 필드
func DbInsertChe(dbctx *sql.DB, pk *Mb.InnerChePacket)error{
	sqlstr := `
INSERT INTO public.tradetbl(
	pairkey  , maketake , symbol   , ordtype  , buysell  , 
	price    , qty      , margin   , fee      , chefee   , 
	feerate  , ordnum   , cheseq   , cntnum   , chetime  ,
	userid   , itemcode , coincode , mkcode   , userfield
)VALUES(
	$1 ,  $2,  $3,  $4,  $5,  $6,  $7,  $8,  $9, $10,
	$11, $12, $13, $14, $15, $16, $17, $18, $19, $20
);` 

	_, err := dbctx.Exec( sqlstr,
			pk.CheKey,                         // 1
			string(pk.MakeTake),               // 2
			api.CleanString(pk.Symbol[:]),     // 3
			string(pk.OrdType),                // 4
			string(pk.BuySell),                // 5
			pk.Price.String(),                 // 6
			pk.Qty.String(),                   // 7
			pk.Margin.String(),                // 8
			pk.Fee.String(),                   // 9
			pk.CheFee.String(),                // 10
			pk.FeeRate.String(),               // 11
			pk.OrdNum,                         // 12
			pk.CheSeq,                         // 13
			pk.CntOrdNum,                      // 14
			time.UnixMicro(int64(pk.CheTime)), // 15
			pk.UserId,                         // 16
			pk.ItemCode,                       // 17
			pk.CoinCode,                       // 18
			pk.MkCode,                         // 19
			api.CleanString(pk.UserField[:]),  // 20
		)
	if err != nil{
		return fmt.Errorf("Inser Error : %v", err)
	}
	return nil

}

