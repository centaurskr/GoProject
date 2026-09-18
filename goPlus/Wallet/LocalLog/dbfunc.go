package main
import (
   "database/sql"
    _ "github.com/lib/pq"
    Log "centolib/logout"
	Mb "centolib/matchingboard"
	Bd "centolib/bindecimal"
	"encoding/json"
	"strconv"
    "time"
    "fmt"
)
// DbSelectMargin은 wallet 테이블에서 사용자의 잔고와 보유 증거금을 조회하여 로그에 출력한다.
func DbSelectMargin(msg string, dbctx *sql.DB, userId, coinCode string){
var amt, hold float64
    // SELECT 잔고 - 증거금
    sqlstr := fmt.Sprintf(
    `SELECT amt, hold
    FROM public.wallet 
    WHERE userid = %s AND coincode = %s;`, userId, coinCode)

    err := dbctx.QueryRow(sqlstr).Scan(&amt, &hold)
    if err != nil{
        Log.Error("증거금 Select :%s:%v", sqlstr, err)
		return
    }
	Log.Debug("-(%s)증거금 Amt[%v] hold[%v] [%s:%s]", 
		msg, amt,hold, userId, coinCode)
	return
}
// DbUpdateError는 취소주문 오류 발생 시 ordertbl의 status를 'ER'로 업데이트한다.
func DbUpdateError(dbctx *sql.DB, stream []byte)error{

	var pkjson Mb.OrderJson
	err := json.Unmarshal(stream, &pkjson)	
	if err != nil { return err;}
	sqlstr := `
	UPDATE  ordertbl
	SET 
		status = 'ER',
		errcode = $1,
	WHERE
		itemcode = $2
	AND
		ordnum = $3;`

	_, err = dbctx.Exec(sqlstr, 
				pkjson.State,
				pkjson.ItemCode,
				pkjson.OrdNum)
	if err != nil { return err; }
	Log.Infor("주문 오류Update 성공OrdNum[%s]msg[%s]", pkjson.OrdNum, pkjson.Message)

	return nil

}
// DbCancelOrder는 취소주문을 처리한다.
// 원주문 상태를 'CC'로, 취소주문 상태를 'CM'으로 변경한다.
func DbCancelOrder(dbctx *sql.DB, stream []byte)error{
	var pkjson Mb.OrderJson
	err := json.Unmarshal(stream, &pkjson)	
	if err != nil { return err;}
	// 원주문 주문 잔량 Select
	sqlstr :=`SELECT remqty FROM ordertbl WHERE itemcode = $1 AND ordnum = $2;`
	var remQty float64
	rows, err := dbctx.Query(sqlstr, pkjson.ItemCode, pkjson.OrgNum)
	if err != nil{
		return fmt.Errorf("%v-[%s]-%s(%d,%d) ",
			err, pkjson.Symbol, sqlstr, pkjson.ItemCode, pkjson.OrgNum)
	}
	defer rows.Close()

	rowCnt := 0
	for rows.Next(){
		if err := rows.Scan(&remQty); err != nil {
			return fmt.Errorf(
		"%v:%s-SELECT remqty FROM ordertbl WHERE itemcode=%s AND ordnum =%s",
		err, pkjson.Symbol, pkjson.ItemCode, pkjson.OrgNum)
		}
		rowCnt ++
	}
	if rowCnt != 1 {
		return fmt.Errorf("[%s]-%s(%d,%d) select count(%d)가 1이 아님",
			pkjson.Symbol, sqlstr, pkjson.ItemCode, pkjson.OrgNum, rowCnt)
	}
	// 원주문 Update (canqty, remqty, status)
	sqlstr = `
	UPDATE ordertbl
	SET canqty = $1,
		remqty = 0,
		status = 'CC'
	WHERE
		itemcode = $2 
	AND
		ordnum = $3;`
	_,err =dbctx.Exec(sqlstr, remQty, pkjson.ItemCode, pkjson.OrgNum)
	if err != nil{return err;}
Log.Infor("1/2 취소주문. 원주문(%v) Update OK. 취소수량[%v]", pkjson.OrgNum,remQty)
	// 취소주문 상태 Update
	sqlstr = `
	UPDATE ordertbl
	SET status = 'CM'
	WHERE
		itemcode = $1 
	AND
		ordnum = $2;`
	_,err =dbctx.Exec(sqlstr, pkjson.ItemCode, pkjson.OrdNum)
	if err != nil{return err;}
	Log.Infor("2/2 취소주문. 취소주문 Update 'CM' OK.")
	return nil
}
// DbUpdateOrdNum은 ordertbl에 주문번호(ordnum)·주문시각(ordtime)을 기록하고 상태를 RG→GO로 변경한다.
func DbUpdateOrdNum(dbctx *sql.DB, stream []byte) error{
	var pkjson Mb.OrderJson

	err := json.Unmarshal(stream, &pkjson)	
	if err != nil { return err;}

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

	_, err = dbctx.Exec(sqlstr, 
				pkjson.OrdNum,
				ordtime,
				pkjson.ItemCode,
				pkjson.RegNum)
	if err != nil { return err; }
	Log.Infor("주문번호 Update 성공OrdNum[%s]", pkjson.OrdNum)
	return nil
}
// DbUpdateOrder는 체결 처리 후 ordertbl의 체결수량(cheqty)·잔량(remqty)·상태를 업데이트한다.
// 잔량이 0이 되면 상태를 'CM'으로 변경한다.
func DbUpdateOrder( dbctx *sql.DB, pk *Mb.InnerChePacket)error{

	// 체결수량 Select
	sqlstr :=`SELECT cheqty FROM ordertbl WHERE itemcode = $1 AND ordnum = $2;`
	var cheqty float64

	// 기존 체결 수량 Select
	rows, err := dbctx.Query(sqlstr, pk.ItemCode, pk.OrdNum)
	if err != nil{
		return fmt.Errorf("%v-[%s]-%s(%d,%d) ",
			err, string(pk.Symbol[:]), sqlstr, pk.ItemCode, pk.OrdNum)
	}
	defer rows.Close()

	rowCnt := 0
	for rows.Next(){
		if err := rows.Scan(&cheqty); err != nil {
			return fmt.Errorf(
		"%v:%s-SELECT cheqty FROM ordertbl WHERE itemcode=%d AND ordnum =%d",
		err, string(pk.Symbol[:]), pk.ItemCode, pk.OrdNum)
		}
		rowCnt ++
	}
	if rowCnt != 1 {
		return fmt.Errorf("[%s]-%s(%d,%d) select count(%d)가 1이 아님",
			string(pk.Symbol[:]), sqlstr, pk.ItemCode, pk.OrdNum, rowCnt)
	}

	// 체결수량, 주문상태 Update
	cheQty,_ := pk.Qty.Add(cheqty, Bd.NOROUND, 8)
	sqlstr = `UPDATE ordertbl
		SET cheqty = $1, remqty = $2, status = $3
		WHERE itemcode = $4 AND ordnum = $5;`
	var status string
	if Bd.DecimalIsZero(pk.RemQty) == true {
		status = "CM"
	}else{
		status = "GO"
	}
	_, err = dbctx.Exec(sqlstr, 
		cheQty.String(), 
		pk.RemQty.String(),
		status,
		pk.ItemCode,
		pk.OrdNum)
	if err != nil {
		return fmt.Errorf("%v-%s-%s:(%s,%s,%s,%d,%d)", 
			err, sqlstr, 
			cheQty.String(), 
			pk.RemQty.String(),
			status,
			pk.ItemCode,
			pk.OrdNum)
	}
	return nil
}
// DbSellCheProcess는 매도 체결을 처리한다.
// CoinCode 잔고를 차감하고 MkCode(거래통화) 잔고를 증가시키는 wallet 업데이트를 수행한다.
func DbSellCheProcess(jsonPk Mb.JsonChePacket)error{
fmt.Printf("SELL : %+v\n", jsonPk)
	// CoinCode "-"
DbSelectMargin("S체결 전Co", DbContext, jsonPk.UserId, jsonPk.CoinCode)
DbSelectMargin("S체결 전Mk", DbContext, jsonPk.UserId, jsonPk.MkCode)
	sqlstr := fmt.Sprintf(`
	UPDATE public.wallet
	SET 
		hold = hold - %s
	WHERE userid = %s AND coincode = %s;`,
	jsonPk.Qty, jsonPk.UserId, jsonPk.CoinCode)

	tx, err := DbContext.Begin()
	if err != nil{
		Log.Error("DbContext.Begin()() : %v : %+v", err, jsonPk)
		return err
	}
	_, err = tx.Exec(sqlstr)
	if err != nil{
		tx.Rollback()
		Log.Error("%s:%v:%+v", sqlstr, err, jsonPk)
		return err
	}

	// MkCode "+"
	sqlstr = fmt.Sprintf(`
	UPDATE public.wallet
	SET
		amt = amt + %s - %s
	WHERE userid = %s AND coincode = %s;`,
		jsonPk.CheAmt, jsonPk.CheFee,
		jsonPk.UserId, jsonPk.MkCode)
	_, err = tx.Exec( sqlstr)
	if err != nil{
		tx.Rollback()
		Log.Error("%s:%v:%+v", sqlstr, err, jsonPk)
		return err
	}
	err = tx.Commit()
	if err != nil{
		Log.Error("DbBuyProcecss() Commit :%v:%+v", err, jsonPk)
		return err
	}
DbSelectMargin("S체결 후Co", DbContext, jsonPk.UserId, jsonPk.CoinCode)
DbSelectMargin("S체결 후Mk", DbContext, jsonPk.UserId, jsonPk.MkCode)
	return nil
	
}
// DbBuyCheProcess는 매수 체결을 처리한다.
// CoinCode 잔고를 증가시키고 MkCode(거래통화)의 증거금을 해소한다.
func DbBuyCheProcess(jsonPk Mb.JsonChePacket)error{
fmt.Printf("Buy : %+v\n", jsonPk)
DbSelectMargin("B체결 전Co", DbContext, jsonPk.UserId, jsonPk.CoinCode)
DbSelectMargin("B체결 전Mk", DbContext, jsonPk.UserId, jsonPk.MkCode)
	// CoinCode + (BTC)
	sqlstr := fmt.Sprintf(`
	UPDATE public.wallet
	SET  amt = amt + %s
	WHERE userid = %s and coincode = %s;`, 
	jsonPk.Qty, jsonPk.UserId, jsonPk.CoinCode)

	tx, err := DbContext.Begin()
	if err != nil{
		Log.Error("DbContext.Begin()() : %v : %+v", err, jsonPk)
		return err
	}
	_, err = tx.Exec(sqlstr)
	if err != nil{
		tx.Rollback()
		Log.Error("%s:%v:%+v", sqlstr, err, jsonPk)
		return err
	}
	// MkCode - (USDT)
	marginFree,_ := Bd.NewFromString(jsonPk.Margin)
	marginFree,_  = marginFree.Add(jsonPk.Fee, Bd.NOROUND, 8)
	sqlstr = fmt.Sprintf(`
	UPDATE public.wallet
	SET
		amt = amt - %s + %s - %s,
		hold = hold - %s
	WHERE
		userid = %s AND coincode = %s;`, 
		jsonPk.CheAmt, marginFree.String(), jsonPk.CheFee, 
		marginFree.String(),
		jsonPk.UserId, jsonPk.MkCode)

Log.Debug("CheAmt(%s) CheFee(%s) margin(%s)+Fee(%s)= %s",
	jsonPk.CheAmt, jsonPk.CheFee, jsonPk.Margin, jsonPk.Fee, 
	marginFree.String())

	_, err = tx.Exec( sqlstr)
	if err != nil{
		tx.Rollback()
		Log.Error("%s:%v:%+v", sqlstr, err, jsonPk)
		return err
	}
	err = tx.Commit()
	if err != nil{
		Log.Error("DbBuyProcecss() Commit :%v:%+v", err, jsonPk)
		return err
	}
DbSelectMargin("B체결 후Co", DbContext, jsonPk.UserId, jsonPk.CoinCode)
DbSelectMargin("B체결 후Mk", DbContext, jsonPk.UserId, jsonPk.MkCode)
	return nil
}
// DbCancelProcess는 취소 체결을 처리한다.
// 매수 취소 시 MkCode 증거금을, 매도 취소 시 CoinCode 증거금을 wallet에 반환한다.
func DbCancelProcess(jsonPk Mb.OrderJson)error{
// jsonPk.Qty 취소수량
// 주문증거금
// jsonPk.Margin 
// jsonPk.Fee

	switch jsonPk.BuySell {
	case "B" :
DbSelectMargin("Cancel-매수 전", DbContext, jsonPk.UserId, jsonPk.MkCode)
	// marginFree = jsonPk.Margin + jsonPk.Fee
	// amt = amt + marginFree
	// hold = hold - marginFree 
		marginFree,_ := Bd.NewFromString(jsonPk.Margin)
		marginFree,_  = marginFree.Add(jsonPk.Fee, Bd.NOROUND, 8)
		sqlstr := fmt.Sprintf(`
		UPDATE public.wallet 
		SET
			amt = amt + %s,
			hold = hold - %s
		WHERE
			userid = %s AND coincode = %s;`,
		marginFree.String(), marginFree.String(),
		jsonPk.UserId, jsonPk.MkCode)
		_, err := DbContext.Exec(sqlstr)
		if err != nil{
			Log.Error("CANCEL(B) :%v:%s:%v", err, sqlstr, jsonPk)
			return err
		}
DbSelectMargin("Cancel-매수 후", DbContext, jsonPk.UserId, jsonPk.MkCode)
	case "S" :
DbSelectMargin("Cancel-매도 전", DbContext, jsonPk.UserId, jsonPk.CoinCode)
		sqlstr := fmt.Sprintf(`
		UPDATE public.wallet
		SET
			amt = amt + %s,
			hold = hold - %s
		WHERE
			userid = %s AND coincode = %s;`,
		jsonPk.Qty, jsonPk.Qty,
		jsonPk.UserId, jsonPk.CoinCode)
		_, err := DbContext.Exec(sqlstr)
		if err != nil{
			Log.Error("CANCEL(S) :%v:%s:%v", err, sqlstr, jsonPk)
			return err
		}
DbSelectMargin("Cancel-매도 후", DbContext, jsonPk.UserId, jsonPk.CoinCode)
	}
	return nil
}
// DbMarketDone은 시장가 주문 완료 처리를 한다.
// 매수 시 남은 증거금을, 매도 시 수량 증거금을 wallet에 반환한다.
func DbMarketDone(jsonPk Mb.OrderJson)error{
fmt.Printf("MKDONE : %+v\n", jsonPk)
	switch jsonPk.BuySell {
	case "B" :
	// 매수(B)
	// marginFree = jsonPk.Margin + jsonPk.Fee
	// amt = amt + marginFree
	// hold = hold - marginFree 
DbSelectMargin("MKDONE-전", DbContext, jsonPk.UserId, jsonPk.MkCode)
		marginFree,_ := Bd.NewFromString(jsonPk.Margin)
		marginFree,_  = marginFree.Add(jsonPk.Fee, Bd.NOROUND, 8)
		sqlstr := fmt.Sprintf(`
		UPDATE public.wallet 
		SET
			amt = amt + %s,
			hold = hold - %s
		WHERE
			userid = %s AND coincode = %s;`,
		marginFree.String(), marginFree.String(),
		jsonPk.UserId, jsonPk.MkCode)
		_, err := DbContext.Exec(sqlstr)
		if err != nil{
			Log.Error("MKDONE(B) :%v:%s:%v", err, sqlstr, jsonPk)
			return err
		}
DbSelectMargin("MKDONE-후", DbContext, jsonPk.UserId, jsonPk.MkCode)
	case "S" :
	// 매도(S)
	// JsonPk.Margin :  주문증거금 해지 수량
		sqlstr := fmt.Sprintf(`
		UPDATE public.wallet
		SET
			amt = amt + %s,
			hold = hold - %s
		WHERE
			userid = %s AND coincode = %s;`,
		jsonPk.Margin, jsonPk.Margin,
		jsonPk.UserId, jsonPk.CoinCode)
		_, err := DbContext.Exec(sqlstr)
		if err != nil{
			Log.Error("DONE(S) :%v:%s:%v", err, sqlstr, jsonPk)
			return err
		}
	}
	return nil
}
// Channel Message Format
// type Message struct{
//     tr string
//     msg [][]byte
// }
// tr 
// -. DBCHE    + JsonChePacket
// -. DBCANCEL + OrderJson
// -. DBSEQ    + OrderJson
// -. DBERR    + OrderJson
// -. MKDONE   + OrderJson
func goDataProcess(){
for{
select {
case packet := <- Channel :
switch packet.tr {
	case "DBCHE":
		var jsonChe Mb.JsonChePacket
		err := json.Unmarshal(packet.msg[0], &jsonChe)
		if err != nil{
			Log.Error("DBCHE : %v : %v\n", err, packet)
			continue
		}
		switch jsonChe.BuySell {
		case "B" :
			DbBuyCheProcess(jsonChe)
		case "S" :
			DbSellCheProcess(jsonChe)
		}
	case "DBCANCEL":
		var jsonOrd Mb.OrderJson
		err := json.Unmarshal(packet.msg[0], &jsonOrd)
		if err != nil{
			Log.Error("DBCANCEL : %v : %v\n", err, packet)
			continue
		}
		DbCancelProcess(jsonOrd)
	case "MKDONE" :
		var jsonOrd Mb.OrderJson
		err := json.Unmarshal(packet.msg[0], &jsonOrd)
		if err != nil{
			Log.Error("MKDONE : %v : %v\n", err, packet)
			continue
		}
		Log.Infor("MKDONE : %+v", jsonOrd)
		DbMarketDone(jsonOrd)
	case "DBERR" :
		//Log.Infor("DBERR : %s", string(packet.msg[0][:]))
	case "DBSEQ":
		//Log.Infor("DBSEQ : %s", string(packet.msg[0][:]))

} // switch(packet.tr)
} // select
} // for Loop 
} // goDataProcess
