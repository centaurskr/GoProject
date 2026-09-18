package main

import (
	"bytes"
	"fmt"
	"log"
	Log "centolib/logout"
)

func main() {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "INFO: ", log.Lshortfile)

		infof = func(info string) {
			logger.Output(2, info)
		}
	)

	infof("Hello world")

	fmt.Print(&buf)

	Log.InitLogger("")
	defer Log.Close()

	Log.Infor("Information....")
	Log.Debug("Debugger\n")
	Log.Error("Error Mode")
}
