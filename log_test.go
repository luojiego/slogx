package log

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestCallerLocation(t *testing.T) {
	// 替换标准输出为我们的pipe，必须先重定向再创建logger，
	// 因为NewLogger内部通过io.MultiWriter在创建时捕获os.Stdout引用
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// 在重定向后创建测试logger，这样logger会写入pipe
	testLogger := NewLogger(Config{
		Level:    slog.LevelDebug,
		Format:   "text",
		Filename: "", // 不写文件
		Stdout:   true,
	})

	// 通过包级函数调用，这样getCallerLocation(3)的调用栈深度正确
	oldLogger := GetDefaultLogger()
	SetDefaultLogger(testLogger)

	Debug("test contains time filed", "time", 321)
	Info("test message")

	SetDefaultLogger(oldLogger)

	// 关闭写入端并读取输出
	w.Close()
	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	// 恢复标准输出
	os.Stdout = oldStdout

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

	// 验证输出包含正确的调用位置（文件名）
	if !strings.Contains(output, "log_test.go:") {
		t.Errorf("Expected log output to contain file name 'log_test.go', got: %s", output)
	}
}
