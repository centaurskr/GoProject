package main

import (
    "fmt"
    "os"
    "path/filepath"
)

func main() {
    // Destination directory (e.g., "data_folder")
    destDir := "data_folder"
/*
    // Ensure the destination directory exists
    err := os.MkdirAll(destDir, 0755)
    if err != nil {
        fmt.Println("Error creating destination directory:", err)
        return
    }
*/
	err := os.Chdir("./BTC.USDT")
    if err != nil {
        fmt.Println("Error finding files:", err)
        return
    }
	

    // Find all files matching "*.data" in the current directory
    files, err := filepath.Glob("MEDO.*")
    if err != nil {
        fmt.Println("Error finding files:", err)
        return
    }

    // Check if any files were found
    if len(files) == 0 {
        fmt.Println("No files matching '*.data' found")
        return
    }

    // Move each file to the destination directory
    for _, file := range files {
        // Construct the new path (destination + filename)
        newPath := filepath.Join(destDir, filepath.Base(file))
/*
        // Move the file
        err = os.Rename(file, newPath)
        if err != nil {
            fmt.Println("Error moving", file, ":", err)
            continue
        }
*/
        fmt.Printf("Moved %s to %s\n", file, newPath)
    }
}
