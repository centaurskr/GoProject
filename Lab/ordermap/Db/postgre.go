// Postgres Test
package main
import (
    "fmt"
    "gopkg.in/yaml.v3"
    "io/ioutil"

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
    var dbstr string = fmt.Sprintf("host=%s prot=%d user=%s password=%s dbname=%s sslmode=disable",
        conf.Host, conf.Port, conf.User, conf.PassWord, conf.DataBase)
    db, err := sql.Open("postgres", dbstr)
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
}

