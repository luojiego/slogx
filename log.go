package log

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"gopkg.in/natefinch/lumberjack.v2"
)

var defaultLogger *Logger

const (
	DefaultMaxSize    = 50  // 默认50MB
	DefaultMaxBackups = 100 // 默认保留100个备份
	DefaultMaxAge     = 30  // 默认保留30天
	DefaultLevel      = "debug"
)

// getLogFileName 获取日志文件名，去除可能的.exe后缀
func getLogFileName() string {
	execName := filepath.Base(os.Args[0])
	// 去除可能的.exe后缀
	execName = strings.TrimSuffix(execName, ".exe")
	return execName + ".log"
}

// getEnvOrDefault 获取环境变量值，如果不存在则返回默认值
func getEnvOrDefault(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

// isProduction 判断是否为生产环境
func isProduction() bool {
	env := strings.ToLower(os.Getenv("GO_ENV"))
	return env == "prod" || env == "production"
}

func init() {
	// 确保logs目录存在
	if err := os.MkdirAll("logs", 0755); err != nil {
		panic("failed to create logs directory: " + err.Error())
	}

	// 获取环境变量配置
	maxSize := getEnvOrDefault("LOG_MAX_SIZE", DefaultMaxSize)
	maxBackups := getEnvOrDefault("LOG_MAX_BACKUPS", DefaultMaxBackups)
	maxAge := getEnvOrDefault("LOG_MAX_AGE", DefaultMaxAge)
	logLevel := getEnvOrDefault("LOG_LEVEL", int(slog.LevelDebug))

	// 根据环境设置压缩和标准输出
	isProd := isProduction()
	compress := isProd
	stdout := !isProd

	// 使用默认配置初始化全局logger
	defaultLogger = NewLogger(Config{
		Level:      slog.Level(logLevel),
		Format:     "text",
		Filename:   filepath.Join("logs", getLogFileName()),
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   compress,
		Stdout:     stdout,
	})
}

// 提供包级别的日志函数
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

func Fatal(msg string, args ...any) {
	defaultLogger.Fatal(msg, args...)
}

// With returns a new Logger with the given attributes added to the global logger
func With(args ...any) *Logger {
	return defaultLogger.With(args...)
}

// SetDefaultLogger allows users to replace the default logger with a custom one
func SetDefaultLogger(l *Logger) {
	defaultLogger = l
}

// GetDefaultLogger returns the current default logger
func GetDefaultLogger() *Logger {
	return defaultLogger
}

// Config 定义日志库的配置
type Config struct {
	Level      slog.Level // 日志级别: debug, info, warn, error
	Format     string     // 输出格式: json, text
	Filename   string     // 日志文件路径
	MaxSize    int        // 每个日志文件的最大兆字节数 (MB)
	MaxBackups int        // 保留的旧日志文件的最大数量
	MaxAge     int        // 保留旧日志文件的最大天数
	Compress   bool       // 是否压缩旧日志文件
	Stdout     bool       // 是否同时输出到标准输出
}

// Logger 是我们封装的日志器
type Logger struct {
	*slog.Logger
	handler    slog.Handler
	level      *slog.LevelVar
	callerSkip int // 添加 callerSkip 字段来控制调用栈跳过的层数
}

// getCallerLocation returns the file name and line number of the caller
func getCallerLocation(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if ok {
		// funcName := runtime.FuncForPC(pc).Name()
		fileName := path.Base(file)
		// funcNames := strings.Split(funcName, ".")
		// funcName = funcNames[len(funcNames)-1]
		var buffer bytes.Buffer
		buffer.WriteString("[")
		buffer.WriteString(fileName)
		// buffer.WriteString(":")
		// buffer.WriteString(funcName)
		buffer.WriteString(":")
		buffer.WriteString(strconv.Itoa(line))
		buffer.WriteString("]")
		return buffer.String()
	}
	return ""
}

// 以下是封装的日志方法，可以直接调用 slog.Logger 的方法
func (l *Logger) Debug(msg string, args ...any) {
	caller := getCallerLocation(3 + l.callerSkip)
	args = append(args, "source", caller)
	l.Logger.Debug(msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	caller := getCallerLocation(3 + l.callerSkip)
	args = append(args, "source", caller)
	l.Logger.Info(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	caller := getCallerLocation(3 + l.callerSkip)
	args = append(args, "source", caller)
	l.Logger.Warn(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	caller := getCallerLocation(3 + l.callerSkip)
	args = append(args, "source", caller)
	l.Logger.Error(msg, args...)
}

// Fatal 级别，通常在记录后退出程序
func (l *Logger) Fatal(msg string, args ...any) {
	caller := getCallerLocation(3 + l.callerSkip)
	// 将 caller 信息添加到 args 中
	args = append(args, "source", caller)
	l.Logger.Error(msg, args...) // slog 没有内置 fatal 级别，通常用 Error 记录后 os.Exit
	os.Exit(1)
}

// With 为 Logger 添加额外的属性
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger:     l.Logger.With(args...),
		handler:    l.handler,
		level:      l.level,
		callerSkip: l.callerSkip + 1, // 增加 callerSkip，因为多了一层调用
	}
}

// WithCallerSkip returns a new Logger with custom caller skip level
func (l *Logger) WithCallerSkip(skip int, args ...any) *Logger {
	newLogger := &Logger{
		Logger:  l.Logger.With(args...),
		handler: l.handler,
		level:   l.level,
	}
	return newLogger
}

// wrappedHandler 包装原有的 handler，添加文件行号
type wrappedHandler struct {
	handler slog.Handler
}

func (h *wrappedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *wrappedHandler) Handle(ctx context.Context, r slog.Record) error {
	// 创建一个新的 Record，先不设置消息
	newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)

	// 先添加调用位置
	newRecord.AddAttrs(slog.String("source", getCallerLocation(4)))

	// 添加原有的其他属性
	r.Attrs(func(a slog.Attr) bool {
		newRecord.AddAttrs(a)
		return true
	})

	return h.handler.Handle(ctx, newRecord)
}

func (h *wrappedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &wrappedHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *wrappedHandler) WithGroup(name string) slog.Handler {
	return &wrappedHandler{handler: h.handler.WithGroup(name)}
}

// WithField creates a logger with a field
func WithField(key string, value any) *slog.Logger {
	// 创建一个新的 handler 来包装原有的 handler
	origLogger := defaultLogger.With(key, value)

	// 创建一个新的 handler，在每次记录日志时添加文件行号
	newHandler := &wrappedHandler{
		handler: origLogger.Handler(),
	}

	return slog.New(newHandler)
}

// NewLogger 初始化并返回一个 Logger 实例
func NewLogger(cfg Config) *Logger {
	var writers []io.Writer

	// 配置 lumberjack
	if cfg.Filename != "" {
		lumberjackLogger := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		writers = append(writers, lumberjackLogger)
	}

	// 是否同时输出到标准输出
	if cfg.Stdout {
		writers = append(writers, os.Stdout)
	}

	// 如果没有配置任何输出，则默认输出到标准输出
	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}

	// 创建一个 MultiWriter 来同时写入多个目标
	multiWriter := io.MultiWriter(writers...)

	// 设置日志级别
	level := &slog.LevelVar{}
	level.Set(cfg.Level)

	var handler slog.Handler
	// 配置 slog Handler
	handlerOptions := &slog.HandlerOptions{
		AddSource: false,
		Level:     level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && a.Value.Kind() == slog.KindTime {
				return slog.Attr{
					Key:   "time",
					Value: slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05.000000")),
				}
			}
			return a
		},
	}

	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(multiWriter, handlerOptions)
	} else {
		handler = slog.NewTextHandler(multiWriter, handlerOptions)
	}

	logger := &Logger{
		Logger:     slog.New(handler),
		handler:    handler,
		level:      level,
		callerSkip: 0, // 初始化时设置为0
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2)
		for sig := range c {
			switch sig {
			case syscall.SIGHUP:
				logger.level.Set(slog.LevelDebug)
				logger.Warn("Log level changed to DEBUG")
			case syscall.SIGUSR1:
				logger.level.Set(slog.LevelInfo)
				logger.Warn("Log level changed to INFO")
			case syscall.SIGUSR2:
				logger.level.Set(slog.LevelWarn)
				logger.Warn("Log level changed to WARN")
			}
		}
	}()

	return logger
}
