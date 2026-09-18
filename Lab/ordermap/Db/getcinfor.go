// COIN정보 가져오기
package main
import (
    "fmt"
    "gopkg.in/yaml.v3"
    "io/ioutil"
	Mb "centolib/matchingboard"
    "database/sql"
    _ "github.com/lib/pq"
)
type DbConfig struct{
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    User     string `yaml:"user"`
    PassWord string `yaml:"password"`
    DataBase string `yaml:"database"`
}
func ConnectDB(conf DbConfig)(*sql.DB, error){
    var dbstr string = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        conf.Host, conf.Port, conf.User, conf.PassWord, conf.DataBase)
    db, err := sql.Open("postgres", dbstr)
	fmt.Printf("CONNECT STR: %s\n", dbstr)
    return db, err
}
func GetDbEnv(filename string) (DbConfig, error){
    var config DbConfig
    yamlFile, err := ioutil.ReadFile(filename)
    if err != nil {
        return config, err
    }

    err = yaml.Unmarshal(yamlFile, &config)
    if err != nil{
        return config, err
    }
    return config, nil
}
func GetCoinInfor(dbctx *sql.DB, name string)(Mb.CoinInfor, error){
	cinfor := Mb.CoinInfor{}
	sqlstr := `
		SELECT symbol, itemcode, coincode, mkcode 
		FROM coininfor
		WHERE symbol = $1;`

	fmt.Println(sqlstr)

	var symbol string
	err := dbctx.QueryRow(sqlstr,name).Scan(
		&symbol, &cinfor.ItemCode, &cinfor.CoinCode, &cinfor.MkCode)
	if err != nil{return cinfor, err;}

	copy(cinfor.Symbol[:], []byte(symbol))
	return cinfor, nil
}
func main(){
    dbinfor , err := GetDbEnv("/home/cento/Project/env/dbconf.yaml")
    if err != nil{
        fmt.Printf("%v\n", err)
        return
    }
    fmt.Printf("%+v\n", dbinfor)

    dbcontext , err := ConnectDB(dbinfor)
    if err != nil{
        fmt.Printf("%v\n", err)
        return
    }
    defer dbcontext.Close()
    fmt.Printf("Db Connect Ok\n");

	cinfor, err := GetCoinInfor(dbcontext, "BTC.USDT")
	if err != nil{
		fmt.Println(err)
		return
	}
	cinfor.Dump(0)
}

