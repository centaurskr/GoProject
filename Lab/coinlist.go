package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// readCoinList reads a file and returns a map where the key is a category and the value is a list of coins.
func readCoinList(fname string) map[string][]string {
	coinMap := make(map[string][]string)

	// Open the file
	file, err := os.Open(fname)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") { // Ignore empty lines or comments
			continue
		}

		// Example format: "Crypto: BTC,ETH,DOGE"
		parts := strings.SplitN(line, "=", 2)
		if len(parts) < 2 {
			fmt.Println("Skipping invalid line:", line)
			continue
		}

		category := strings.TrimSpace(parts[0])
		coins := strings.Split(strings.TrimSpace(parts[1]), ",")

		for i := range coins {
			coins[i] = strings.TrimSpace(coins[i]) // Remove any extra spaces
		}

		coinMap[category] = coins
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}

	return coinMap
}
func writeCoinList(fname string, coinMap map[string][]string) error {
	// Create or overwrite the file
	file, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for category, coins := range coinMap {
		line := fmt.Sprintf("%s=%s\n", category, strings.Join(coins, ", "))
		_, err := writer.WriteString(line)
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}

	// Flush the buffer to ensure all data is written
	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("error flushing file buffer: %v", err)
	}

	return nil
}
func main() {
	/*************
	coinList := readCoinList("COIN.LIST")

	// Print the coin list
	for category, coins := range coinList {
		fmt.Printf("%s: %v\n", category, coins)
	}
	***********/
	coinList := map[string][]string{
		"ORD_01": {"BTC", "ETH", "DOGE"},
		"ORD_02":  {"USD", "EUR", "JPY"},
		"ORD_03": {"Gold", "Silver"},
	}
	writeCoinList("COIN.LIST", coinList)
}

