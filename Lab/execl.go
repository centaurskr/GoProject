package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main(){
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return
	}

	// Construct the full path to the executable
	cmdPath := filepath.Join(homeDir, "bin", "goDumpCoin")

	cmd := exec.Command(cmdPath, "XRP.USDT")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Println("Command failed with exit code:", exitErr.ExitCode())
			fmt.Println("Stderr:", string(exitErr.Stderr))
		} else {
			fmt.Println("Error executing command:", err)
		}
		return
	}
	fmt.Println("Output :", string(output))
}
