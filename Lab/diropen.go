/// DIR Open Read
package main

import (
    Log "centolib/logout"
    "os"
	"path/filepath"
)

func main() {
	Log.InitLogger("")
	defer Log.Close()

    dirPath := "." // Replace with the path to your directory

    // Open the directory
    dir, err := os.Open(dirPath)
    if err != nil {
        Log.Error("%v", err)
		return
    }
    defer dir.Close()

    // Read directory entries
    files, err := dir.Readdir(-1)
    if err != nil {
		Log.Error("%v", err)
		return
    }
	
    // Loop through and print file names
	pattern := "*.go"
    for _, file := range files {
		if file.IsDir() { continue}
		match, err := filepath.Match(pattern, file.Name())
		if err != nil {
			Log.Error("filepath.Match : %v", err)
			break;
		}
		if match {
        	Log.Infor(file.Name())
		}
    }
}
