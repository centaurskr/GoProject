package main

import (
    "fmt"
    "math/rand"
    "time"
)
func getDir()float64{
    rand.Seed(time.Now().UnixNano())
	idx := rand.Intn(2)
	if idx == 0{ 
		return -1
	}else{
		return 1
	}
}
func getPrice(prc float64)float64{
    rand.Seed(time.Now().UnixNano())
    fluctuation := 0.5
    //fluctuation := 0.0001
	change := 1 + (fluctuation * (rand.Float64()*2 - 1))
	return prc *change
}
func main() {
	price := 5000.0
	//price := .00123
	//price := 2.0
	for i := 0; i < 1000; i++{
		price = getPrice(price)
		fmt.Printf("[%d]  - %f\n", i,  price)
		time.Sleep(1 * time.Second)
	}
}

