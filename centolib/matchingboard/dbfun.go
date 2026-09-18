package matchingboard
import (
   "database/sql"
    _ "github.com/lib/pq"
    "fmt"
)
// CoinInfor 구하기
func GetCoinInfor(dbctx *sql.DB, name string)(CoinInfor, error){
    cinfor := CoinInfor{}
    sqlstr := `
        SELECT symbol, itemcode, coincode, mkcode 
        FROM coininfor
        WHERE symbol = $1;`

    //fmt.Println(sqlstr)

    var symbol string
    err := dbctx.QueryRow(sqlstr,name).Scan(
        &symbol, &cinfor.ItemCode, &cinfor.CoinCode, &cinfor.MkCode)
    if err != nil{
        return cinfor, fmt.Errorf("%s(%s)\n%v", sqlstr, name, err)
    }

    copy(cinfor.Symbol[:], []byte(symbol))
    return cinfor, nil
}

