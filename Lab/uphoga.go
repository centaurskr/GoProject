package main

import (
	"fmt"
	"net/http"
	"io"
)

func main() {

	url := "https://api.bithumb.com/v1/orderbook?markets=KRW-BTC"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(string(body))

}
