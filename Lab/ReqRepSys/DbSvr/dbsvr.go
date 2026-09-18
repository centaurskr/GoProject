//
// REQ/REP Db Process Simultion
// 
package main
import (
	zmq "github.com/pebbe/zmq4"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
	"os"
	"fmt"
	//"time"
)

var (
	context *zmq.Context
	myDB    *sql.DB
)
func InsertDb(msg string)error{
	sqlstr := `INSERT INTO logtable VALUES(NULL, ?)`
	if _, err:= myDB.Exec(sqlstr, msg); err != nil{
		return err
	}
	return nil
}

func main(){
	var err error
	if len(os.Args) != 2{
		fmt.Println("Using Check  simulation Program")
		fmt.Printf("Using) %s <REQ endpoint>\n", os.Args[0])
		os.Exit(0)
	}

	// Connect DB (sqlite3)
	db, err := sql.Open("sqlite3", "/home/cento/Order/Mydb.db")
    if err != nil {
        fmt.Println(err)
        os.Exit(0)
    }
	myDB = db;
	
	// Create Zmq Context
	context, err = zmq.NewContext()
	if context == nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}
	defer context.Term()


	// Create Zmq socket
	sock, err := context.NewSocket(zmq.REP)
	if err != nil{
		fmt.Printf("NewSocket(REP)\n%v\n", err)
		os.Exit(0)
	}
	defer sock.Close()
	
	// Bind socket
	err = sock.Bind(os.Args[1])
	if err != nil{
		fmt.Printf("Bind(%s)\n^v", os.Args[1], err)
		os.Exit(0)
	}

	// Data I/O
	for {
		msg, err:= sock.RecvBytes(0)
		if err != nil{
			fmt.Printf("RecvBytes(0) %v", err)
			os.Exit(0)
		}
		err = InsertDb(string(msg[:]))
		if err != nil{
			fmt.Printf("Error : InsertDb(%s)\n", msg)
			os.Exit(0)
		}
		_, err = sock.SendBytes(msg, 0)
		if err !=nil{
			fmt.Printf("SendBytes (%s)\n\t%v\n", string(msg[:]), err)
			break
		}

	}

}
