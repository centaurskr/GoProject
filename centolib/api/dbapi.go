package api
import (
    "gopkg.in/yaml.v3"
    "io/ioutil"
    "database/sql"
    _ "github.com/lib/pq"

    "fmt"
)
type DbConfig struct{
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    User     string `yaml:"user"`
    PassWord string `yaml:"password"`
    DataBase string `yaml:"database"`
}  
// DB 환경을 얻는 Function
// name : DB 접속 Yaml file
func ConnectDb(name string)(*sql.DB, error){
	var conf DbConfig
	yamlf, err := ioutil.ReadFile(name)
	if err != nil{
		return nil, fmt.Errorf("ioutil.ReadFile(%s) %w", name,err)
	}

	err = yaml.Unmarshal(yamlf, &conf)
	if err != nil{
		return nil, fmt.Errorf("yaml.Unmarshal() %w", err)
	}
	var dbstr string = fmt.Sprintf(`
	host=%s port=%d user=%s password=%s dbname=%s sslmode=disable`,	
	conf.Host, conf.Port, conf.User, conf.PassWord, conf.DataBase)

	ctx, err := sql.Open("postgres", dbstr)
	if err != nil{
		return nil, fmt.Errorf("sql.Open(%s)\n\t%w", dbstr, err)
	}
	return ctx, err
}
