package main
import (
	Log "centolib/logout"
	"fmt"
	"os"
	"time"
)

func main(){
	Log.InitLogger("", Log.ALL_MODE)
	defer Log.Close()

	Log.Debug("START ------")
	str := fmt.Sprintf("TEST") + fmt.Sprintf("%d", os.Getpid())
	
	Log.Debug("TEST---- %s", str)
	Log.Error("Error TEST---- %s", str)
	Log.Infor("infor TEST---- %s", str)

	time.Sleep(1*time.Second)

	fmt.Printf("============= File ERROR 변경 ======C[%s]\n", Log.GetConsoleLogLevel())
	Log.SetFileLogLevel(Log.ERROR) // Log-level zap.ErrorLevel
	Log.Infor("File Level ERROR로 변경 후 Infor F(X), C(O)")
	Log.Debug("File Level ERROR로 변경 후 Debug F(X), C(O)")
	Log.Error("File Level ERROR로 변경 후 Error F(O), C(O)")
	Log.Warn("File Level ERROR로 변경 후 Warn F(O), C(O)")

	fmt.Printf("\n============= Console ERROR 변경 ======F[%s]\n", Log.GetFileLogLevel())
	Log.SetConsoleLogLevel(Log.ERROR) // Log-level zap.ErrorLevel
	Log.Infor("Console Level ERROR로 변경 후 Infor F(X), C(X)")
	Log.Debug("Console Level ERROR로 변경 후 Debug F(X), C(X)")
	Log.Error("Console Level ERROR로 변경 후 Error F(O), C(O)")
	Log.Warn("Console Level ERROR로 변경 후 Warn F(O), C(O)")

	fmt.Printf("\n============= File DEBUG 변경 ======C[%s]\n", Log.GetConsoleLogLevel())
	Log.SetFileLogLevel(Log.DEBUG) // Log-level zap.DebugLevel
	Log.Infor("File Level DEBUG로 변경 후 Infor F(O), C(X)")
	Log.Debug("File Level DEBUG로 변경 후 Debug F(O), C(X)")
	Log.Error("File Level DEBUG로 변경 후 Error F(O), C(O)")
	Log.Warn("File Level DEBUG로 변경 후 Warn F(O), C(O)")

	fmt.Printf("\n============= Console DEBUG 변경 ======F[%s]\n", Log.GetFileLogLevel())
	Log.SetConsoleLogLevel(Log.DEBUG) // Log-level zap.DebugLevel
	Log.Infor("Console Level DEBUG로 변경 후 Infor F(O), C(O)")
	Log.Debug("Console Level DEBUG로 변경 후 Debug F(O), C(O)")
	Log.Error("Console Level DEBUG로 변경 후 Error F(O), C(O)")
	Log.Warn("Console Level DEBUG로 변경 후 Warn F(O), C(O)")

	fmt.Printf("\n============= Console WARN 변경 ======F[%s]\n", Log.GetFileLogLevel())
	Log.SetConsoleLogLevel(Log.WARN) // Log-level zap.DebugLevel
	Log.Infor("Console Level WARN로 변경 후 Infor F(O), C(X)")
	Log.Debug("Console Level WARN로 변경 후 Debug F(O), C(X)")
	Log.Error("Console Level WARN로 변경 후 Error F(O), C(O)")
	Log.Warn("Console Level WARN로 변경 후 Warn F(O), C(O)")
}

