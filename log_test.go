package log

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestCallerLocation(t *testing.T) {
	// 创建一个测试logger
	testLogger := NewLogger(Config{
		Level:    slog.LevelDebug,
		Format:   "text",
		Filename: "", // 不写文件
		Stdout:   true,
	})

	// 替换标准输出为我们的pipe
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	testLogger.Debug("test contains time filed", "time", 321)

	// 测试完成后恢复标准输出
	defer func() {
		os.Stdout = oldStdout
	}()

	// 在这里调用日志
	testLogger.Info("test message")

	// 刷新输出并读取内容
	w.Close()
	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	outputStr := string(output)

	// 验证输出中包含正确的文件名和行号
	if !strings.Contains(outputStr, "log_test.go:") {
		t.Errorf("Expected log output to contain file name 'log_test.go', got: %s", outputStr)
	}

	// 验证输出中不包含日志库内部的文件名
	if strings.Contains(outputStr, "log.go:") {
		t.Errorf("Log output should not contain internal logger file name 'log.go', got: %s", outputStr)
	}
}

func TestCallerLocationInDifferentPackage(t *testing.T) {
	// 创建临时目录用于测试
	tmpDir := t.TempDir()

	// 配置logger写入临时文件
	tmpLog := NewLogger(Config{
		Level:    slog.LevelDebug,
		Format:   "text",
		Filename: tmpDir + "/test.log",
		Stdout:   false,
	})

	SetDefaultLogger(tmpLog)

	// 在不同的函数中调用日志
	func() {
		Debug("debug from nested function")
	}()

	// 读取日志文件内容
	content, err := os.ReadFile(tmpDir + "/test.log")
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	output := string(content)

	// 验证输出包含正确的调用位置
	if !strings.Contains(output, "log_test.go:") {
		t.Errorf("Expected log output to contain file name 'log_test.go', got: %s", output)
	}

	// 验证行号是否正确（应该是调用 Debug 的行号）
	if !strings.Contains(output, "log_test.go:61") { // 这里的行号应该是 Debug() 调用的实际行号
		t.Errorf("Expected log output to contain the correct line number, got: %s", output)
	}
}
