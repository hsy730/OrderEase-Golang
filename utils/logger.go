package utils

import (
	"fmt"
	"io"
	"log"
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
