package main
import (
    Log "centolib/logout"
    "os"
    "encoding/json"
    "io/ioutil"
)
type Coin struct {
	Uid        int
    Name       string    `json:"name"`
	ItemCode   int       `json:"itemcode"`
	CoinCode   int       `json:"coincode"`
	MkCode     int       `json:"mkcode"`
    Price      float64   `json:"price"`
    Qty        float64   `json:"qty"`
} 

var CoinList []Coin // Ensure CoinList is defined globally if needed

func readCoin() error {
    file, err := os.Open("coins.json")
    if err != nil {
        return err
    }
    defer file.Close()

    value, err := ioutil.ReadAll(file)
    if err != nil {
        return err
    }

    if err := json.Unmarshal(value, &CoinList); err != nil {
        return err
    }

    // Assuming Log.InforX is a valid logging function in your environment
    for _, coin := range CoinList {
        Log.Infor("%+v", coin)
    }
    return nil
}
func main() {
	Log.InitLogger("",Log.CONSOLE_MODE)
	defer Log.Close()

	err := readCoin()
	if err != nil{
		Log.Error("%v", err)
	}
}
