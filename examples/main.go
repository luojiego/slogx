package main

import (
	"log/slog"

	log "github.com/luojiedev/slogx"
)

func logSomething() {
	log.Debug("This is a debug message")
	log.Info("This is an info message")

	// 测试带字段的日志
	log.Info("User logged in", "userId", 123, "ip", "192.168.1.1")
}

func main() {
	// 直接调用包级别的函数
	log.Info("Application started")

	// 在不同的函数中调用
	logSomething()

	// 测试 With 功能
	logger := log.With("module", "auth")
	logger.Error("Authentication failed", "reason", "invalid_token")

	// 测试不同级别的日志
	log.Debug("Debug level message")
	log.Info("Info level message")
	log.Warn("Warning level message")
	log.Error("Error level message")

	// 演示 GetLevel / SetLevel 动态调整日志等级
	currentLevel := log.GetLevel()
	log.Info("Current log level", "level", currentLevel)

	// 将日志等级提升为 INFO，DEBUG 级别日志将不再输出
	log.SetLevel(slog.LevelInfo)
	log.Info("Log level changed to INFO")
	log.Debug("This debug message should NOT appear")
	log.Info("This info message should still appear")

	// 通过 Logger 实例也能获取和设置等级
	logger2 := log.NewLogger(log.Config{
		Level:    slog.LevelWarn,
		Format:   "text",
		Filename: "",
		Stdout:   true,
	})
	log.Info("New logger level", "level", logger2.GetLevel())
	logger2.SetLevel(slog.LevelError)
	logger2.Warn("This warn message should NOT appear")
	logger2.Error("This error message should still appear")
}
