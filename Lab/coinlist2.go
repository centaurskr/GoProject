package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	// Read the file
	fileName := "COIN.LIST"
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Parse the file content
	var lines []string
	var descriptionLines []string
	var coinLines []string
	coinPattern := regexp.MustCompile(`^ORD_(\d{2})=(.*)`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") { // Description line
			descriptionLines = append(descriptionLines, line)
		} else if match := coinPattern.FindStringSubmatch(line); match != nil {
			coinLines = append(coinLines, match[0]) // Add the coin line
		} else {
			lines = append(lines, line)
		}
	}

	// Find the smallest order number
	minOrder := 99
	for _, coinLine := range coinLines {
		orderMatch := regexp.MustCompile(`ORD_(\d{2})=`).FindStringSubmatch(coinLine)
		if len(orderMatch) > 1 {
			orderNum := orderMatch[1]
			// Convert to integer
			var orderInt int
			fmt.Sscanf(orderNum, "%d", &orderInt)
			if orderInt < minOrder {
				minOrder = orderInt
			}
		}
	}

	// Increment to the next available order number
	newOrderNum := fmt.Sprintf("%02d", minOrder+1)
	newCoin := "Litecoin" // Modify this to the coin you want to add
	newOrderLine := fmt.Sprintf("ORD_%s=%s", newOrderNum, newCoin)

	// Reconstruct the file content
	// First, add the description lines
	lines = append(lines, descriptionLines...)
	// Then, add the new coin line
	lines = append(lines, newOrderLine)

	// Write the updated content back to the file
	err = os.WriteFile(fileName, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("New coin added successfully.")
}

