package log2

import (
	"os"
	"testing"
)

func TestGetLogger(t *testing.T) {
	// 测试获取日志记录器实例
	logger1 := GetLogger()
	logger2 := GetLogger()
	
	// 验证返回的不是 nil
	if logger1 == nil {
		t.Errorf("Expected logger to be non-nil")
	}
	if logger2 == nil {
		t.Errorf("Expected logger to be non-nil")
	}
}

func TestInitLogger(t *testing.T) {
	// 重置内部日志记录器
	inner = nil
	
	// 测试初始化日志记录器
	InitLogger()
	
	// 验证内部日志记录器被初始化
	if inner == nil {
		t.Errorf("Expected inner logger to be initialized, got nil")
	}
	
	// 验证日志目录是否创建
	logDir := "./logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Errorf("Expected log directory to be created, but it doesn't exist")
	}
}

func TestSync(t *testing.T) {
	// 初始化日志记录器
	InitLogger()
	
	// 测试同步日志缓冲区
	Sync()
	// 没有返回值，只要不报错就通过
}

func TestClose(t *testing.T) {
	// 初始化日志记录器
	InitLogger()
	
	// 测试关闭日志记录器
	Close()
	// 没有返回值，只要不报错就通过
}

func TestLogMethods(t *testing.T) {
	// 初始化日志记录器
	InitLogger()
	
	// 测试各种日志方法（除了 Fatal 和 Fatalf，因为它们会导致进程退出）
	// 这些方法应该能够正常执行而不会导致错误
	Errorf("Test errorf: %s", "test")
	Warnf("Test warnf: %s", "test")
	Infof("Test infof: %s", "test")
	Debugf("Test debugf: %s", "test")
	
	Error("Test error")
	Warn("Test warn")
	Info("Test info")
	Debug("Test debug: %s", "test")
	
	// 测试未初始化的情况
	inner = nil
	
	// 这些方法在未初始化时应该能够正常执行而不会导致错误
	Errorf("Test errorf uninitialized: %s", "test")
	Warnf("Test warnf uninitialized: %s", "test")
	Infof("Test infof uninitialized: %s", "test")
	Debugf("Test debugf uninitialized: %s", "test")
	
	Error("Test error uninitialized")
	Warn("Test warn uninitialized")
	Info("Test info uninitialized")
	Debug("Test debug uninitialized: %s", "test")
}

func TestLoggerStruct(t *testing.T) {
	// 测试 Logger 结构体
	logger := GetLogger()
	if logger == nil {
		t.Errorf("Expected logger to be non-nil")
	}
	
	// 测试 Logger 结构体的方法
	logger.Errorf("Test logger errorf: %s", "test")
	logger.Warnf("Test logger warnf: %s", "test")
	logger.Infof("Test logger infof: %s", "test")
	logger.Debugf("Test logger debugf: %s", "test")
}
