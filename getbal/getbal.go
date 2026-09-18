package main

import (
	"fmt"
	"log"
	"centolib/fulcrum"
	//"libs/fulcrum"
)
func main() {
	address := "bc1qmnnq9vv4nhd35px2j7dkks5smvgu96a5sakf3v"
	fulcrumServer := "192.168.0.102:50001" 
	rtnstr, err := fulcrum.GetBalance(address, fulcrumServer)
	if err != nil {
		log.Fatalf("Error querying Fulcrum server: %v", err)
	}
	fmt.Printf("Addr : %s\nfulcrum Svr : %s\n", address, fulcrumServer)
	fmt.Printf("Return : %s\n", rtnstr)
}

