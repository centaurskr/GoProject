package main
import (
   "database/sql"
    _ "github.com/lib/pq"
    //Log "centolib/netlog"
	Mb "centolib/matchingboard"
	Bd "centolib/bindecimal"
	"centolib/api"
	"strconv"
    "time"
    "fmt"
)
// DbUpdateError는 취소주문 오류 발생 시 ordertbl의 status를 'ER'로 업데이트한다.
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
	//Log.Infor("주문 오류Update 성공OrdNum[%s]msg[%s]", pkjson.OrdNum, pkjson.Message)

	return nil

}
// DbCancelOrder는 취소주문을 처리한다.
// 원주문 상태를 'CC'로, 취소주문 상태를 'CM'으로 변경하고 잔량·증거금을 정산한다.
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
//Log.Infor("1/2 취소주문. 원주문(%v) Update OK. 취소수량[%s]", pkjson.OrgNum,pkjson.Qty)
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
	//Log.Infor("2/2 취소주문. 취소주문 Update 'CM' OK.")
	return nil
}
// DbUpdateMkDone은 시장가 주문 완료 후 ordertbl의 잔량을 0으로 설정하고 상태를 'CM'으로 변경한다.
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
	result,err :=dbctx.Exec(sqlstr, pkjson.Margin, pkjson.Fee,
			pkjson.ItemCode, pkjson.OrdNum)
	if err != nil{return err;}
	rows, err := result.RowsAffected()
	if err != nil{return err;}
	if rows == 0 {
		return fmt.Errorf("MKDONE Update Row = 0 : %+v", pkjson)
	}

	return nil
}
// DbUpdateOrdNum은 ordertbl에 주문번호(ordnum)·주문시각(ordtime)을 기록하고 상태를 RG→GO로 변경한다.
func DbUpdateOrdNum(dbctx *sql.DB, pkjson Mb.OrderJson) error{

//Log.Debug("DBSEQ : RegNum[%s] OrdNum[%s] time[%s]", 
//pkjson.RegNum, pkjson.OrdNum, pkjson.OrdTime)

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

	result, err := dbctx.Exec(sqlstr, 
				pkjson.OrdNum,
				ordtime,
				pkjson.ItemCode,
				pkjson.RegNum)
	if err != nil { return err; }
	rows, err := result.RowsAffected()
	if err != nil{ return err;}
	if rows == 0{
		return fmt.Errorf("UPDATE ROW = 0, OrdNum(%s) itemcode(%s) RegNum(%s)", 
			pkjson.OrdNum, pkjson.ItemCode, pkjson.RegNum)
	}
	//Log.Infor("주문번호 Update 성공OrdNum[%s] rows[%d]", pkjson.OrdNum, rows)
	return nil
}
// DbUpdateOrder는 체결 처리 후 ordertbl의 체결수량(cheqty)·잔량(remqty)·상태를 업데이트한다.
// 잔량이 0이 되면 상태를 'CM'으로 변경한다.
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
	result, err := dbctx.Exec(sqlstr, 
		pk.Qty.String(), pk.Qty.String(), 
		pk.Margin.String(), pk.Fee.String(), status,
		pk.ItemCode, pk.OrdNum)
	if err != nil {
		return fmt.Errorf("%v-%s-%+v", err, sqlstr, pk)
	}
	rows, err := result.RowsAffected()
	if err != nil{ return err;}
	if rows == 0{
		return fmt.Errorf("DbUpdateOrder rows = 0-%s-%+v", sqlstr, pk)
	}
	return nil
}
// DbInsertChe는 체결 데이터를 tradetbl에 삽입한다.
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

