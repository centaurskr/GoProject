package main

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"time"
)

func main() {
	// Open or create the lock file
	lockFile := "/tmp/mylockfile.lock"
	file, err := os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("Failed to open lock file: %v", err)
	}
	defer file.Close()

	// Try to acquire an exclusive lock
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		log.Println("Another process is already running.")
		return
	}
	//defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN) // Release the lock

	fmt.Println("Lock acquired, doing work...")
	time.Sleep(10 * time.Second) // Simulate work
	fmt.Println("Work done, exiting.")
}

