package main

import (
	"fmt"
	"time"
	"strconv"
)

func main() {
	// Create a ticker that ticks every 1 minute
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	dbMin,_ := strconv.Atoi(time.Now().Truncate(1 * time.Minute).Format("200601021504"))
	fmt.Printf("dbMin Trucate(1min) : %d(yyyymmddhhmm)\n", dbMin)

	// Signal indefinitely every minute
	fmt.Println("Ticker started. Signals every 1 minute:")
	for {
		select {
		case t := <-ticker.C:
			fmt.Printf("Signal at: %s -> %s\n", 
			t.Format("2006-01-02 15:04:05"),
			t.Truncate(2 * time.Minute).Format("2006-01-02 15:04:05"))
			
			cmin, _ := 	strconv.Atoi(t.Truncate(1 * time.Minute).Format("200601021504"))
			if dbMin != cmin {
				fmt.Printf("Change [%d] -> [%d]\n", dbMin, cmin)
				dbMin = cmin
			}
		}
	}
}

