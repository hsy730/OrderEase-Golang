package log2

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	*zap.SugaredLogger
}

var inner *zap.Logger

func GetLogger() *Logger {
	if inner == nil {
		InitLogger()
	}
	return &Logger{inner.Sugar()}
}

// InitLogger 初始化zap日志记录器
func InitLogger() {
	// 确保日志目录存在
	if err := os.MkdirAll("./logs", 0755); err != nil {
		panic(fmt.Sprintf("创建日志目录失败: %v", err))
	}

	// 配置日志轮转
	logFileName := fmt.Sprintf("./logs/app_%s.log", time.Now().Format("2006-01-02"))
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFileName,
		MaxSize:    100, // megabytes
		MaxAge:     7,   // days
		MaxBackups: 3,
		LocalTime:  true,
		Compress:   true,
	}

	// 创建zap编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建不同输出的core
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// 创建不同级别的日志输出
	consoleDebugging := zapcore.Lock(os.Stdout)
	fileWriter := zapcore.AddSync(lumberjackLogger)

	// 创建core
	consoleCore := zapcore.NewCore(consoleEncoder, consoleDebugging, zapcore.DebugLevel)
	fileCore := zapcore.NewCore(fileEncoder, fileWriter, zapcore.DebugLevel)

	// 创建tee core，同时写入控制台和文件
	core := zapcore.NewTee(consoleCore, fileCore)

	// 创建logger
	inner = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// Sync 同步日志缓冲区
func Sync() {
	if inner != nil {
		inner.Sync()
	}
}

// Close 关闭日志记录器
func Close() {
	if inner != nil {
		inner.Sync()
	}
}

func Fatalf(format string, v ...interface{}) {
	if inner != nil {
		inner.Sugar().Fatalf(format, v...)
	}
}

// Errorf 记录错误级别日志
func Errorf(format string, v ...interface{}) {
	if inner != nil {
		inner.Sugar().Errorf(format, v...)
	}
}

// Warnf 记录警告级别日志
func Warnf(format string, v ...interface{}) {
	if inner != nil {
		inner.Sugar().Warnf(format, v...)
	}
}

// Infof 记录信息级别日志
func Infof(format string, v ...interface{}) {
	if inner != nil {
		inner.Sugar().Infof(format, v...)
	}
}

// Debugf 记录调试级别日志
func Debugf(format string, v ...interface{}) {
	if inner != nil {
		inner.Sugar().Debugf(format, v...)
	}
}

func Fatal(v ...interface{}) {
	if inner != nil {
		inner.Sugar().Fatal(v...)
	}
}

func Error(v ...interface{}) {
	if inner != nil {
		inner.Sugar().Error(v...)
	}
}

func Warn(v ...interface{}) {
	if inner != nil {
		inner.Sugar().Warn(v...)
	}
}

func Info(v ...interface{}) {
	if inner != nil {
		inner.Sugar().Info(v...)
	}
}

func Debug(format string, v ...interface{}) {
	if inner != nil {
		inner.Sugar().Debug(v...)
	}
}
