package log2

import (
	"fmt"
	"io"
	"log"
	"orderease/config"
	"os"
	"time"
)

var Logger *log.Logger

func InitLogger() {
	// 创建日志目录
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal("创建日志目录失败:", err)
	}

	// 创建或打开日志文件
	logFile := fmt.Sprintf("logs/app_%s.log", time.Now().Format("2006-01-02"))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("打开日志文件失败:", err)
	}

	// 设置日志输出到文件和控制台
	Logger = log.New(io.MultiWriter(file, os.Stdout), "", log.Ldate|log.Ltime|log.Lshortfile)

}

// 定义日志级别
const (
	LogLevelSilent = iota
	LogLevelError
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

// 获取日志级别
func getLogLevel() int {
	logLevel := config.AppConfig.Database.LogLevel
	switch logLevel {
	case 1:
		return LogLevelSilent
	case 2:
		return LogLevelError
	case 3:
		return LogLevelWarn
	case 4:
		return LogLevelInfo
	case 5:
		return LogLevelDebug
	default:
		return LogLevelInfo
	}
}

// 不同级别的日志输出函数
func Errorf(format string, v ...interface{}) {
	if getLogLevel() >= LogLevelError {
		Logger.Printf("[ERROR] "+format, v...)
	}
}

func Warnf(format string, v ...interface{}) {
	if getLogLevel() >= LogLevelWarn {
		Logger.Printf("[WARN] "+format, v...)
	}
}

func Infof(format string, v ...interface{}) {
	if getLogLevel() >= LogLevelInfo {
		Logger.Printf("[INFO] "+format, v...)
	}
}

func Debugf(format string, v ...interface{}) {
	if getLogLevel() >= LogLevelDebug {
		Logger.Printf("[DEBUG] "+format, v...)
	}
}
