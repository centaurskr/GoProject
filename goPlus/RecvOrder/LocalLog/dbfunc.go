package main
import (
   "database/sql"
    _ "github.com/lib/pq"
	"time"
	"strconv"
	Mb "centolib/matchingboard"
    "fmt"
	"centolib/api"
	Log "centolib/logout"
	Bd "centolib/bindecimal"
)
/*
		COALESCE(NULLIF($1 ,''), 'N/A'),
		COALESCE(NULLIF($2 ,''), 'N/A'),
		COALESCE(NULLIF($3 ,''), '-1'),
		$4 ,
		$5 ,
		COALESCE(NULLIF($6 ,''), '0'),   //rem
		COALESCE(NULLIF($7 ,''), '0'),   //che
		COALESCE(NULLIF($8 ,''), '0'),   //can
		COALESCE(NULLIF($9 ,''), '-1'),   
		COALESCE(NULLIF($10,''), '-1'),
		COALESCE(NULLIF($11,''), '-1'),  // UserType
		COALESCE(NULLIF($12,''), '-1'),
		COALESCE(NULLIF($13,''), '@'),   // OrdType
		COALESCE(NULLIF($14,''), '@'),   // BuySell
		COALESCE(NULLIF($15,''), '-1'),  // Price
		COALESCE(NULLIF($16,''), '-1'),  // Qty
		COALESCE(NULLIF($17,''), '-1'),  // M rate
		COALESCE(NULLIF($18,''), '-1'),  // T Rate
		COALESCE(NULLIF($19,''), '-1'),  // C unit
		COALESCE(NULLIF($20,''), '-1'),  // Min Trade
		COALESCE(NULLIF($21,''), '@'),   // Cancel
		COALESCE(NULLIF($22,''), 'N/A'),
		COALESCE(NULLIF($23,''), 'N/A'),
		COALESCE(NULLIF($24,''), 'N/A'),
		COALESCE(NULLIF($25,''), '0'),   // OrgNum
		$26,
		$27,
		COALESCE(NULLIF($28,''), '0'),  // Margin
		COALESCE(NULLIF($29,''), '0'),  // Fee
		$30,
		$31,
		$32,
		COALESCE(NULLIF($33,''), 'N/A'));`)
*/
func printParameters(params []interface{}) {
    for i, param := range params {
        switch v := param.(type) {
        case sql.NullString:
            if v.Valid {
                fmt.Printf("Param %d: %s = %s\n", 
					i+1, api.GetTypeName(param), v.String)
            } else {
                fmt.Printf("Param %d: %s = NULL\n", 
					i+1, api.GetTypeName(param))
            }
        case sql.NullInt64:
            if v.Valid {
                fmt.Printf("Param %d: %s = %d\n", 
					i+1, api.GetTypeName(param), v.Int64)
            } else {
                fmt.Printf("Param %d: %s = NULL\n", 
					i+1, api.GetTypeName(param))
            }
        case sql.NullFloat64:
            if v.Valid {
                fmt.Printf("Param %d: %s = %f\n", 
					i+1, api.GetTypeName(param), v.Float64)
            } else {
                fmt.Printf("Param %d: %s = NULL\n", 
					i+1, api.GetTypeName(param))
            }
        default:
            fmt.Printf("Param %d: %s = %+v\n", 
				i+1, api.GetTypeName(param), param)
        }
    }
}
// DbSelectLastExrate는 exratetbl에서 지정 currency의 가장 최근 환율을 조회한다.
// 반환값: 기준시각 문자열, 환율(float64), error
func DbSelectLastExrate(dbctx *sql.DB, currency string)(string, float64, error){
	var sqlstr string

	sqlstr = fmt.Sprintf(`
		SELECT basetime, exrate
		FROM public.exratetbl
		where currency = '%s'
		ORDER BY TO_TIMESTAMP(basetime, 'YYYYMMDDHH24MI') DESC
		LIMIT 1;`, currency)

	var basetime string
	var exrate float64

	err := dbctx.QueryRow(sqlstr).Scan(&basetime, &exrate)
	if err != nil{
		return basetime, exrate, fmt.Errorf("환율 %s-Select : %v", sqlstr, err)
	}
	return basetime, exrate, nil
}
// DbInsertOrder는 주문 JSON 데이터를 ordertbl에 삽입한다.
// 오류 주문은 status='ER', 정상 주문은 status='RG'로 저장한다.
func DbInsertOrder(dbctx *sql.DB, ordjson *Mb.OrderJson)error{
var sqlstr string = fmt.Sprintf(`
	INSERT INTO public.ordertbl ( 
		symbol   , channel  , itemcode , ordnum   , ordtime  , 
		remqty   , cheqty   , canqty   , mkcode   , userid   ,
		usertype , coincode , ordtype  , buysell  , price    ,
		qty      , makerate , takerate , callunit , mintrade ,
		cancel   , ip       , uuid     , userfield, orgnum   ,
		regnum   , regtime  , margin   , fee      , status   ,
		errcode  , exrate
	)VALUES(
		COALESCE(NULLIF($1 ,''), 'N/A'),
		COALESCE(NULLIF($2 ,''), 'N/A'),
		COALESCE(NULLIF($3 ,''), '-1')::INTEGER,
		$4 ,
		$5 ,
		COALESCE(NULLIF($6 ,''), '0.0')::NUMERIC,
		COALESCE(NULLIF($7 ,''), '0.0')::NUMERIC,
		COALESCE(NULLIF($8 ,''), '0.0')::NUMERIC,
		COALESCE(NULLIF($9 ,''), '-1')::INTEGER,   
		COALESCE(NULLIF($10,''), '-1')::INTEGER,
		COALESCE(NULLIF($11,''), '-1')::INTEGER,
		COALESCE(NULLIF($12,''), '-1')::INTEGER,
		COALESCE(NULLIF($13,''), '@'),
		COALESCE(NULLIF($14,''), '@'),
		COALESCE(NULLIF($15,''), '-1.0')::NUMERIC,
		COALESCE(NULLIF($16,''), '-1.0')::NUMERIC,
		COALESCE(NULLIF($17,''), '-1.0')::NUMERIC,
		COALESCE(NULLIF($18,''), '-1.0')::NUMERIC,
		COALESCE(NULLIF($19,''), '-1.0')::NUMERIC,
		COALESCE(NULLIF($20,''), '-1.0')::NUMERIC,
		COALESCE(NULLIF($21,''), '@'),
		COALESCE(NULLIF($22,''), 'N/A'),
		COALESCE(NULLIF($23,''), 'N/A'),
		COALESCE(NULLIF($24,''), 'N/A'),
		COALESCE(NULLIF($25,''), 'N/A'),
		$26,
		$27,
		COALESCE(NULLIF($28,''), '0.0')::NUMERIC,
		COALESCE(NULLIF($29,''), '0.0')::NUMERIC, 
		$30,
		$31,
		COALESCE(NULLIF($32,''), '1.0')::NUMERIC);`)
	var ordnum, remqty ,cheqty ,canqty ,status string

	ordnum  = ordjson.RegNum     // 주문번호
	remqty  = ordjson.Qty        // 주문잔량

	unixTimestamp,err := strconv.ParseInt(ordjson.RegTime, 10, 64)
	if err != nil{return err;}
	regtime := time.UnixMicro(unixTimestamp) //  접수시각
	
	// 주문 상태 & 오류 코드 설정
	if ordjson.State != "0" {
		status = "ER"
	}else{
		status = "RG"   // Matching 에서 "GO"로 변경
	}
	_, err = dbctx.Exec(sqlstr, 
				ordjson.Symbol   ,
				ordjson.Channel  ,
				ordjson.ItemCode ,
					ordnum   ,
					regtime  ,
					remqty   ,
					cheqty   ,
					canqty   ,
				ordjson.MkCode   ,
				ordjson.UserId   ,
				ordjson.UserType ,
				ordjson.CoinCode ,
				ordjson.OrdType  ,
				ordjson.BuySell  ,
				ordjson.Price    ,
				ordjson.Qty      ,
				ordjson.MakeRate ,
				ordjson.TakeRate ,
				ordjson.CallUnit ,
				ordjson.MinTrade ,
				ordjson.Cancel   ,
				ordjson.UserIp   ,
				ordjson.Uuid     ,
				ordjson.UserField,
				ordjson.OrgNum   ,
				ordjson.RegNum   ,
				    regtime  ,
				ordjson.Margin   ,
				ordjson.Fee      ,
					status   ,
				ordjson.State    ,
				ordjson.ExRate)
	if err != nil{
		if Log.GetConsoleLogLevel() != "NULL"{
			fmt.Printf("Executing SQL: %s\n",sqlstr)
 			params := []interface{}{
        	ordjson.Symbol, ordjson.Channel, ordjson.ItemCode,
        	ordnum, regtime, remqty,
        	cheqty, canqty, ordjson.MkCode, ordjson.UserId,
        	ordjson.UserType, ordjson.CoinCode, ordjson.OrdType,
        	ordjson.BuySell, ordjson.Price, ordjson.Qty,
        	ordjson.MakeRate, ordjson.TakeRate, ordjson.CallUnit,
        	ordjson.MinTrade, ordjson.Cancel, ordjson.UserIp,
        	ordjson.Uuid, ordjson.UserField, ordjson.OrgNum,
        	ordjson.RegNum, regtime, ordjson.Margin,
        	ordjson.Fee, status, ordjson.State, ordjson.Message,
			ordjson.ExRate,
    		}
			printParameters(params)
		}
		return fmt.Errorf("Insert Error:%v", err)
	}
	return nil
}

// DbCheckMargin은 wallet 테이블에서 증거금 가용 여부를 확인하고, 충분하면 증거금을 차감한다.
// 반환값: 증거금 충분 여부(bool), error
func DbCheckMargin(dbctx *sql.DB, pk *Mb.OrderJson)(bool, error){
	var (
		userId    int
		coinCode  int
		amt       float64
		margin    Bd.BinDecimal
	)
	userId, _ = strconv.Atoi(pk.UserId)

	// BTC.USDT
	switch pk.BuySell {
	case "S" : // 매도 -- 주문 수량 
		coinCode, _=strconv.Atoi(pk.CoinCode) // BTC
		rtn := margin.SetFromString(pk.Qty)
		if rtn != 0{
			return false, fmt.Errorf("Convert Qty:%s", pk.Qty)
		}
	case "B" : // 매수 -- margin and fee
		coinCode, _=strconv.Atoi(pk.MkCode) // USDT
		rtn := margin.SetFromString(pk.Margin)
		if rtn != 0{
			return false, fmt.Errorf("Convert Margin:%s", pk.Margin)
		}
		margin, rtn = margin.Add(pk.Fee, Bd.NOROUND, 8)
		if rtn != 0{
			return false, fmt.Errorf("Margin.Add(fee):%s", pk.Fee)
		}
	}
	// SELECT 잔고 - 증거금
	sqlstr := fmt.Sprintf(
	`SELECT amt - hold
	FROM public.wallet 
	WHERE userid = %d AND coincode = %d;`, userId, coinCode)

	err := dbctx.QueryRow(sqlstr).Scan(&amt)
	if err != nil{
		return false, fmt.Errorf("매도 Select :%s:%v", sqlstr, err)
	}
	rtn := Bd.DecimalCompare(amt, margin)
	if rtn == -1{ // 증거금 부족
		return false, nil
	}
//Log.Debug("증거금 rtn[%d]  amt :%v  margin;%v", rtn, amt, margin.String())	
	
	// 증거금 UPDATE
	sqlstr =`
	UPDATE public.wallet
	SET
		amt  = amt  - $1,
		hold = hold + $2 
	WHERE
		userid = $3 AND coincode = $4;` 
	_, err = dbctx.Exec(sqlstr, 
			margin.String(), 
			margin.String(), 
			userId, coinCode)
	if err != nil{ 
		return false, fmt.Errorf("증거금 Update 오류")
	}
	return true, nil
}
