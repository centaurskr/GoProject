///
/// go api server
///
package main

import (
	"log"
	"centolib/fulcrum"
    "encoding/json"

    "net/http"
    "github.com/gorilla/websocket"
)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Balnace struct {
        CONFIRMED   string `json:"Confirmed"`
        UNCONFIRMED string `json:"Unconfirmed"`
}

// Structure Wss json
type WssJson struct {
        METHOD string `json:"Method"`
        PARAM  string `json:"Param"`
}

/// /get_balance
func get_balance_handler(w http.ResponseWriter, r *http.Request){

	fulcrumServer := "192.168.0.102:50001" 

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil{
		log.Fatalf("Error upgrading to websocket : %v", err)
	}
	defer ws.Close()

	for {
		msgtype, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Error reading message : %v", err)
			break;
		}
		injson := WssJson{}
		err = json.Unmarshal(msg, &injson)
		if err != nil {
			log.Printf("Recv msg : %s\nerr : %v\n", string(msg), err)
		}else{
			log.Printf("IN - Method : %s, Param : %s\n", injson.METHOD, injson.PARAM)
			outstr, err := fulcrum.GetBalance(injson.PARAM, fulcrumServer)
			if err != nil {
				log.Println("Error from fulcrum : %v", err)
			}
			outbyte, err := json.Marshal(outstr)
			if err != nil {
				log.Printf("Error json.Marshal(%s) %v", outstr, err)
				continue
			}
			log.Printf("OUT - %s\n", outstr)
			err = ws.WriteMessage(msgtype, outbyte)
			if err != nil {
				log.Printf("Error writing message : %v", err)
				break;
			}
		}
	}
}
func main() {
	http.HandleFunc("/get_balance", get_balance_handler)
	log.Println("START Api server 50005")
	err := http.ListenAndServe(":50005", nil)
	if err != nil {
		log.Fatalf("Server failed : %v", err)
	}
}

