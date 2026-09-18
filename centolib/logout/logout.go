//
// ┌────────────────────────────────────────────────────┐
// │           ZAP를 이용한 Logout Module               │
// │ file name : logout.go                              │
// │ Date      : 22024. 11. 15. (금) 14:41:09 KST       │
// │ Author    : Cento                                  │
// └────────────────────────────────────────────────────┘
//
//-. 최초 InitLogger() 호출해야한다
//   환경변수 LOG_HOME을 사용한다. 만일 LOG_HOME가 설정되니 않은 경우
//   Logfile은 "/tmp/Log"아래에 {Processname}.log로생성된다.
//  
//-.  InitLogger(ext string, mode int)
//  ext : Logfile의 추가 문자. 같은 이름의 Process가 여러개 실행되는 경우 
//        구문 문자를 주어서 Logfile을 생성하기 위한 string이다.
//        ext 가 ""이 아닌경우 Logfile 이름은 {Processname}_{ext}.log로 생성된다
//  mode : 출력 대상을 알려 주는 것으로 FILE_MODE, CONSOLE_MODE, ALL_MODE가 있다
//        FILE_MODE    : Log message를 File에 기록한다
//        CONTOLE_MODE : Log message를 STDOUT으로 출력한다
//        ALL_MODE     : Log Mesasge를 File과 STDOUT으로 모두 출력한다
//
//-. Level별 Log function은 Debug(), Infor(), Warn(), Error()가 있다
//   사용법은 Printf() 용법과 동일하다
//-. console mode에서는 Level문자에 색이 지원된다.
//-. 각각의 색은 INFOR([34m), DEBUG([35m), WARN([33m) ERROR([31m)이다
//-. Default log level은 DebugLevel이다
// 
package logout

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strings"
	"time"
	"fmt"
)

var (
	log          *zap.Logger       // Zap logger
	fileLevel    *zap.AtomicLevel   // File Log Level
	consoleLevel *zap.AtomicLevel   // Consol Log Level
)
const (
	DEBUG = zapcore.DebugLevel
	INFOR = zapcore.InfoLevel
	WARN  = zapcore.WarnLevel
	ERROR = zapcore.ErrorLevel

	CONSOLE_MODE = int(0x01)
	FILE_MODE    = int(0x10)
	ALL_MODE     = int(CONSOLE_MODE | FILE_MODE)
)

func InitLogger(ext string, mode int) {

	path := os.Getenv("LOG_HOME")
	if path == "" {
		path = "/tmp/Log"	
	}
	strs := strings.Split(os.Args[0], "/")	
	name := strs[len(strs) -1]
	if ext != "" {
		path = path + "/" + name + "_" + ext + ".log"
	}else{
		path = path + "/" + name + ".log"
	}
	
	///////////////////////// FILE Logger ///////////////////////////////
	fileCoreFunc := func()zapcore.Core{
		// Configure lumberjack for log file rotation
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   path,
			MaxSize:    100,  // Max size in MB before rotation
			MaxBackups: 5,    // Max number of backups to keep
			MaxAge:     30,   // Max age in days to keep logs
			Compress:   true, // Compress log files after rotation
		})
		// JSON encoder configuration for the log file
		fileEncoderConfig := zapcore.EncoderConfig{
			TimeKey:       "T",
			LevelKey:      "L",
			CallerKey:     "C",
			MessageKey:    "M",
			EncodeLevel:   zapcore.CapitalLevelEncoder,
			EncodeTime:    func (t time.Time, enc zapcore.PrimitiveArrayEncoder){
                tstr := t.Format("2006/01/02-15:04:05")
                timestr := tstr + "." + fmt.Sprintf("%09d", t.Nanosecond())
                enc.AppendString(timestr)
        	},
			EncodeCaller:  zapcore.ShortCallerEncoder,
			LineEnding:    zapcore.DefaultLineEnding,
		}
		// Default loglevel = DebugLevel
		lv := zap.NewAtomicLevelAt(zap.DebugLevel)
		fileLevel = &lv
		// Create separate encoders for JSON (file) and console (stdout)
		fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)
		// Create cores with different encoders and output destinations
		return zapcore.NewCore(fileEncoder, fileWriter, *fileLevel)
	}

	///////////////////////// Console Logger ////////////////////////////
	consoleCoreFunc := func()zapcore.Core{
		// Console encoder configuration for stdout
		consoleEncoderConfig := zapcore.EncoderConfig{
			TimeKey:       "T",
			LevelKey:      "L",
			CallerKey:     "C",
			MessageKey:    "M",
			EncodeTime:    func (t time.Time, enc zapcore.PrimitiveArrayEncoder){
               	tstr := t.Format("15:04:05")
               	timestr := tstr + "." + fmt.Sprintf("%09d", t.Nanosecond())
               	enc.AppendString(timestr)
        	},
			EncodeLevel:   zapcore.CapitalColorLevelEncoder,
			EncodeCaller:  zapcore.ShortCallerEncoder,
			LineEnding:    zapcore.DefaultLineEnding,
		}

		// Default loglevel = DebugLevel
		lv := zap.NewAtomicLevelAt(zap.DebugLevel)
		consoleLevel = &lv
		// Create separate encoders for JSON (file) and console (stdout)
		consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
		// Create cores with different encoders and output destinations
		return zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), *consoleLevel)
	}

	var core zapcore.Core
	// Combine both cores
	switch mode {
		case CONSOLE_MODE : 
			core = zapcore.NewTee(consoleCoreFunc()) // Console 만 - Devmode
		case FILE_MODE    : 
			core = zapcore.NewTee(fileCoreFunc()) // File 만-Production mode
		case ALL_MODE     : 
			core = zapcore.NewTee(fileCoreFunc(), consoleCoreFunc()) // 모두
		default           : 
			core = zapcore.NewTee(consoleCoreFunc()) // Console 만 - Devmode
	}
	// Create the global logger with caller info
	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	Infor("🔵===== START ===== FileLevel[%s] ConsoleLevel[%s]", 
		GetFileLogLevel(),
		GetConsoleLogLevel())
	for i:= 1; i < len(os.Args); i++ {
		Infor("  = ✓ Arg[%d] : %s", i, os.Args[i])
	}
	Infor("  =====================================================")
}
// Debug Message를 출력한다.
// 표시 형태는 "[D] Message ..."의 형태로 
// Mode 출력 색은 Yellow이다
func Debug(format string, a ...interface{}){
	if log == nil {return}
	msg := fmt.Sprintf(format, a...)
	log.Debug(msg)
}
func Infor(format string, a ...interface{}){
	if log == nil {return}
	msg := fmt.Sprintf(format, a...)
	log.Info(msg)
}
func Error(format string, a ...interface{}){
	if log == nil {return}
	msg := fmt.Sprintf(format, a...)
	log.Error(msg)
}
func Warn(format string, a ...interface{}){
	if log == nil {return}
	msg := fmt.Sprintf(format, a...)
	log.Warn(msg)
}
func Close() {
	log.Sync()
}
func SetFileLogLevel(loglv zapcore.Level){
	var save    zapcore.Level   // Save console Log Level
	if fileLevel == nil {return}
	if consoleLevel != nil{
		save = consoleLevel.Level()
		consoleLevel.SetLevel(zap.DebugLevel)
	}
	fileLevel.SetLevel(zap.DebugLevel)

	Infor("--------File LogLevel Change : OLD[%v] New[%v]. Console[%s]", 
		fileLevel, loglv, save.String())

	fileLevel.SetLevel(loglv)
	if consoleLevel != nil{
		consoleLevel.SetLevel(save)
	}
}
func SetConsoleLogLevel(loglv zapcore.Level){
	var save    zapcore.Level   // Save file Log Level
	if consoleLevel == nil {return}
	if fileLevel != nil{
		save = fileLevel.Level()
		fileLevel.SetLevel(zap.DebugLevel)
	}
	consoleLevel.SetLevel(zap.DebugLevel)

	Infor("----Console LogLevel Change : OLD[%v] New[%v]. File[%s]", 
		consoleLevel, loglv, save.String())

	consoleLevel.SetLevel(loglv)
	if fileLevel != nil{
		fileLevel.SetLevel(save)
	}
}
func GetFileLogLevel()string{
	if fileLevel != nil{
		return fileLevel.String()
	}else{
		return "NULL"
	}
}
func GetConsoleLogLevel()string{
	if consoleLevel != nil {
		return consoleLevel.String()
	}else{
		return "NULL"
	}
}

