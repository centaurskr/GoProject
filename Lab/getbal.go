package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"github.com/btcsuite/btcutil/bech32"
)

// Struct to parse JSON-RPC response from Fulcrum
type BalanceResponse struct {
	Result struct {
		Confirmed   int64 `json:"confirmed"`
		Unconfirmed int64 `json:"unconfirmed"`
	} `json:"result"`
	ID      int    `json:"id"`
	JSONRPC string `json:"jsonrpc"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Function to decode Bech32 (SegWit) address
func decodeBech32Address(address string) ([]byte, error) {
	hrp, data, err := bech32.Decode(address)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Bech32 address: %v", err)
	}

	if hrp != "bc" {
		return nil, fmt.Errorf("unsupported human-readable part: %s", hrp)
	}

	// Convert 5-bit Bech32 data to 8-bit data
	convertedData, err := bech32.ConvertBits(data[1:], 5, 8, false)
	if err != nil {
		return nil, fmt.Errorf("failed to convert bits: %v", err)
	}

	return convertedData, nil
}

// Function to generate the scripthash from a Bech32 address
func generateScripthash(address string) (string, error) {
	witnessProgram, err := decodeBech32Address(address)
	if err != nil {
		return "", err
	}

	// Create the script: OP_0 <witness program>
	script := append([]byte{0x00, byte(len(witnessProgram))}, witnessProgram...)

	// Hash the script using SHA256
	hash := sha256.Sum256(script)

	// Reverse the hash bytes to match Electrum protocol format
	for i, j := 0, len(hash)-1; i < j; i, j = i+1, j-1 {
		hash[i], hash[j] = hash[j], hash[i]
	}

	// Return the scripthash as a hex string
	return hex.EncodeToString(hash[:]), nil
}

// Function to send a JSON-RPC request to the Fulcrum server
func queryFulcrumServer(scripthash, fulcrumAddress string) (*BalanceResponse, error) {
	// Create the JSON-RPC request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "blockchain.scripthash.get_balance",
		"params":  []string{scripthash},
		"id":      1,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Connect to the Fulcrum server via TCP (port 50001)
	conn, err := net.Dial("tcp", fulcrumAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Fulcrum server: %v", err)
	}
	defer conn.Close()

	// Send the JSON-RPC request
	_, err = conn.Write(append(requestBytes, '\n'))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	// Read the server response
	reader := bufio.NewReader(conn)
	responseBytes, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
fmt.Println(string(requestBytes), string(responseBytes))
	// Parse the JSON response
	var response BalanceResponse
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// Check if there is an error in the response
	if response.Error != nil {
		return nil, fmt.Errorf("error from Fulcrum server: %v", response.Error.Message)
	}

	return &response, nil
}

func main() {
	// Example Bech32 address (replace with your actual Bitcoin address)
	address := "bc1qmnnq9vv4nhd35px2j7dkks5smvgu96a5sakf3v"

	// Generate the scripthash
	scripthash, err := generateScripthash(address)
	if err != nil {
		log.Fatalf("Error generating scripthash: %v", err)
	}

	// Query the Fulcrum server for the balance (replace with your actual Fulcrum server address)
	fulcrumServer := "192.168.0.102:50001" // Replace with your actual Fulcrum server's address
	balanceResponse, err := queryFulcrumServer(scripthash, fulcrumServer)
	if err != nil {
		log.Fatalf("Error querying Fulcrum server: %v", err)
	}

	// Print the balance information
	//fmt.Printf("Confirmed balance: %d sats\n", balanceResponse.Result.Confirmed)
	//fmt.Printf("Unconfirmed balance: %d sats\n", balanceResponse.Result.Unconfirmed)
	rtnbyte, _ := json.Marshal(balanceResponse.Result)
	fmt.Println(string(rtnbyte))
}

