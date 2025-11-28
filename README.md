# Slogx

A structured logging library for Go, built on top of `slog` with automatic log rotation, environment awareness, and flexible configuration.

[中文文档](README_zh.md)

## Features

- Built on Go 1.21+ `slog` package
- Automatic log rotation (powered by lumberjack)
- JSON and text output formats
- Auto-naming log files based on program name
- Environment-aware configuration (test/production)
- Environment variables support
- Console and file output support
- Dynamic log level adjustment via signals
- Structured logging with field support

## Installation

```bash
go get github.com/luojiego/slogx
```

## Quick Start

```go
package main

import (
    "fmt"
    slogx "github.com/luojiego/slogx"
)

func main() {
    // Use package-level functions
    slogx.Info("Application started")
    slogx.Debug("Debug information")
    slogx.Error("Error occurred", "error", fmt.Errorf("321"))

    // Use With to add extra fields
    logger := slogx.With("module", "user-service")
    logger.Info("User logged in", "userId", 123)
}
```

## Configuration

### Default Settings

- Log file location: `./logs/<program-name>.log`
- Log level: Debug
- Output format: Text
- Single log file size: 50MB
- Number of backup files: 100
- Log retention days: 30 days

### Environment Variables

Configure logging behavior through environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| LOG_MAX_SIZE | Maximum size of each log file (MB) | 50 |
| LOG_MAX_BACKUPS | Maximum number of old log files | 100 |
| LOG_MAX_AGE | Days to retain old log files | 30 |
| GO_ENV | Runtime environment (production/prod for production) | - |

### Environment-Specific Behavior

Test Environment (Default):
- Outputs to both console and file
- No compression for old log files

Production Environment (GO_ENV=production/prod):
- File output only
- Automatic compression for old log files

## Custom Configuration

Use `NewLogger` function for custom configuration:

```go
logger := slogx.NewLogger(slogx.Config{
    Level:      "debug",
    Format:     "json",
    Filename:   "custom.log",
    MaxSize:    100,    // MB
    MaxBackups: 10,     // number of files
    MaxAge:     7,      // days
    Compress:   true,   // compress old files
    Stdout:     true,   // console output
})

// Set as default logger (optional)
slogx.SetDefaultLogger(logger)
```

## Dynamic Log Level Adjustment

Adjust log levels at runtime using system signals:

- `SIGHUP`: Set to Debug level
- `SIGUSR1`: Set to Info level
- `SIGUSR2`: Set to Warn level

Example (Unix/Linux):
```bash
# Switch to Debug level
kill -HUP <pid>

# Switch to Info level
kill -USR1 <pid>

# Switch to Warn level
kill -USR2 <pid>
```

## Dependencies

- Go 1.21+
- gopkg.in/natefinch/lumberjack.v2

## License

MIT License

## Contributing

Issues and Pull Requests are welcome!

## Author

[luojiego](https://github.com/luojiego)