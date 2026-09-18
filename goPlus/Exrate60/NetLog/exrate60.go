// 환율 정보를 60분마다 Alpha Vantage API에서 조회하여 DB에 저장하는 프로세스.
package main

import (
	"fmt"
	"os"
	"time"
	"net/http"
	"io/ioutil"
	"strconv"
	"centolib/api"
	Log "centolib/netlog"
	zmq "github.com/pebbe/zmq4"
	"database/sql"
	_ "github.com/lib/pq"
	"encoding/json"
)
const (
	TICK_SEC = 30
	DATA_MIN = 60
)

// DbInsertUsdExrate는 USD 환율 정보를 exratetbl에 삽입하거나 갱신한다.
func DbInsertUsdExrate(dbctx *sql.DB, basetime int, getstr string, exrate string)error{
	sqlstr :=`
	INSERT INTO public.exratetbl (basetime, currency, exrate, timestr)
	VALUES($1, $2, $3, $4)
	ON CONFLICT (basetime, currency)
	DO UPDATE SET exrate = EXCLUDED.exrate, timestr = EXCLUDED.timestr;`

	btime :=fmt.Sprintf("%d", basetime) // yyyymmddhhmi (10분단위)
	_,err := dbctx.Exec(sqlstr, btime, "USD", exrate, getstr)
	if err != nil {
		return err
	}
	return nil
}
	
// AlphaVantageResponse는 Alpha Vantage API의 실시간 환율 응답 구조체다.
type AlphaVantageResponse struct {
    RealtimeCurrencyExchangeRate struct {
        FromCurrencyCode string `json:"1. From_Currency Code"`
        FromCurrencyName string `json:"2. From_Currency Name"`
        ToCurrencyCode   string `json:"3. To_Currency Code"`
        ToCurrencyName   string `json:"4. To_Currency Name"`
        ExchangeRate     string `json:"5. Exchange Rate"`
        LastRefreshed    string `json:"6. Last Refreshed"`
        TimeZone         string `json:"7. Time Zone"`
        BidPrice         string `json:"8. Bid Price"`
        AskPrice         string `json:"9. Ask Price"`
    } `json:"Realtime Currency Exchange Rate"`
}
// getExchangeRateUSDKRW는 Alpha Vantage API에서 USD/KRW 환율을 조회한다.
// 반환값: 환율 문자열, 조회시각(YYYYMMDDHHMiSS KST), error
func getExchangeRateUSDKRW()(string, string, error){
	url := `https://www.alphavantage.co/query?function=CURRENCY_EXCHANGE_RATE&from_currency=USD&to_currency=KRW&apikey=CGKF7AX2H13T4UHN`

    resp, err := http.Get(url)
    if err != nil {
        return "", "", fmt.Errorf("API 요청 실패: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", "", fmt.Errorf("API 응답 오류: %s", resp.Status)
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", "", fmt.Errorf("응답 읽기 실패: %v", err)
    }

    var result AlphaVantageResponse
    if err := json.Unmarshal(body, &result); err != nil {
        return "", "", fmt.Errorf("JSON 파싱 실패: %v", err)
    }

    //fmt.Println(string(body), result)

    // 환율이 비어있는지 확인
    if result.RealtimeCurrencyExchangeRate.ExchangeRate == "" {
        return "", "", fmt.Errorf("환율 정보를 찾을 수 없습니다")
    }
	// 한국시간으로 변경 YYYYMMDDHHMiSS
	loc, _ := time.LoadLocation("Asia/Seoul")
	utcTime, err := time.ParseInLocation("2006-01-02 15:04:05",
		result.RealtimeCurrencyExchangeRate.LastRefreshed, time.UTC)
	if err != nil {
		Log.Error("시간파싱: RESULT %+v", result)
        return "", "", fmt.Errorf("시간 파싱 오류")
	}
	kstTime := utcTime.In(loc)
    return result.RealtimeCurrencyExchangeRate.ExchangeRate,
			kstTime.Format("20060102150405"), nil
}

func main() {
	err := Log.InitNetLogout(os.Args[0], "")
	if err != nil{
		fmt.Printf("InitNetLogout() : %v\n" , err)
		os.Exit(0)
	}
    defer Log.Close()

    isrun, err := api.IsRunning("goExchange60")
    if isrun == true || err != nil {
        Log.Error("goExchange60는 실행중입니다 : %v", err)
        os.Exit(0)
    }
    // PUB SOCKET
    proxyfile := os.Getenv("PROXY_YAML")
        if proxyfile == ""{
        Log.Error("환경 변수에 PROXY_YAML이 없습니다.")
        os.Exit(0)
    }
    pubpoint, err := api.GetProxyInfor(proxyfile, "EXTRAINFOR", "front")
    if err != nil{
        Log.Error("%s file의 EXTRAINFOR Proxy에 front가 없습니다", proxyfile)
        os.Exit(0)
    }
    pubsocket, _ := zmq.NewSocket(zmq.PUB)
    err = pubsocket.Connect(pubpoint)
    if err != nil{
        Log.Error("Fail to connect (%s) : %v", pubpoint, err)
        os.Exit(0)
    }
    defer pubsocket.Close()

    Log.Infor("EXTRAINFOR Proxy front(PUB) end point is : %s", pubpoint)
    time.Sleep(100 * time.Millisecond)

    // Db Connect
    envfile := os.Getenv("PLATFORM_ENV")
    if envfile == ""{
        Log.Error("PLATFORM_ENV 환경변수가 없습니다")
        os.Exit(0)
    }
    Log.Infor("envfile : %s", envfile)
    dbyaml, err := api.GetEnvValue(envfile, "DBCONFIG_YAML")
    if err != nil {
        Log.Error("%s에 DBCONFIG_YAML이 없습니다", envfile)
        os.Exit(0)
    }
    dbContext, err := api.ConnectDb(dbyaml)
    if err != nil{
        Log.Error("ConnectDb(%s) : %v", dbyaml, err)
        os.Exit(0)
    }

	// Create a ticker (TICK_SEC Second)
	ticker := time.NewTicker(TICK_SEC * time.Second)
	defer ticker.Stop()

	minNum := 0
	Log.Infor("Ticker 시각 %d초 단위, Data 단위 %d분단위", TICK_SEC, DATA_MIN)

	for{
		select{
			case t := <-ticker.C:
				// 60분간격
				cmin, _ := strconv.Atoi(
					t.Truncate(DATA_MIN * time.Minute).Format("200601021504"))
				if cmin != minNum {
					exrate, tmstr, err := getExchangeRateUSDKRW()
					if err != nil{
						minNum = cmin
						continue	
					}

					minNum = cmin
					//fmt.Printf("Change [%d] timestr[%s] exrate[%s]\n", 
					// minNum, tmstr, exrate)
					// Db Insert 
					err = DbInsertUsdExrate(dbContext, cmin, tmstr, exrate)
					if err != nil{
						Log.Error("%v\nDB오류 : 기준시각(%d) USD/KRW(%s)",
							err, cmin, exrate)
					}

					// 내부 MESSAGE PUB
					msg := fmt.Sprintf(
						`{"CURRENCY":"USD", "TIME":"%d", "RATE":"%s" }`, 
						cmin, exrate)
            		// PUB : EXRATE + USDT + Base time + Exchange rate
            		_, err = pubsocket.SendMessage("EXRATE", msg)
            		if err != nil{
                		Log.Error("Pub.SendMessage(EXRATE), %v", err)
            		}

				}

		} // Select
	} // Main Loop
}
