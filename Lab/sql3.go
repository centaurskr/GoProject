package main
import (
	"fmt"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
)
var (
	myDB *sql.DB
)

func CreateTable()error{
	sqlstr := `
		CREATE TABLE IF NOT EXISTS logtable(
			id INTEGER NOT NULL PRIMARY KEY,
			msg TEXT
		);
	`

	if _, err := myDB.Exec(sqlstr); err != nil{
		if err != nil{
			return err
		}
	}
	return nil
}
func main(){
	db, err := sql.Open("sqlite3", "/home/cento/Order/Mydb.db")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	myDB = db
	err = CreateTable()
	if err != nil{
		fmt.Println(err)
		os.Exit(0)
	}
	fmt.Println("Create Ok /home/cento/Order/Mydb.db")
}
