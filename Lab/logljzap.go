package main

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
		Infor("= ✓ Arg[%d] : %s", i, os.Args[i])
	}
	Infor("=======================================================")
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
	save := consoleLevel.Level()
	consoleLevel.SetLevel(zap.DebugLevel)
	fileLevel.SetLevel(zap.DebugLevel)

	Infor("--------File LogLevel Change : OLD[%v] New[%v]. Console[%s]", 
		fileLevel, loglv, save.String())

	fileLevel.SetLevel(loglv)
	consoleLevel.SetLevel(save)
}
func SetConsoleLogLevel(loglv zapcore.Level){
	save := fileLevel.Level()
	fileLevel.SetLevel(zap.DebugLevel)
	consoleLevel.SetLevel(zap.DebugLevel)

	Infor("----Console LogLevel Change : OLD[%v] New[%v]. File[%s]", 
		consoleLevel, loglv, save.String())

	consoleLevel.SetLevel(loglv)
	fileLevel.SetLevel(save)
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
func main() {
	InitLogger("", CONSOLE_MODE)
	Close()

	// Example usage of the logger
	Infor("This is an info message")
	Debug("This is a Debug message")
	Warn("This is a Debug message")
	Error("This is an error message")
}

