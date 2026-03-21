package helper

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a new zap.Logger that logs to both the CLI and a file.
// The log file is stored in ~/.idx-scraper/<commandName>/<timestamp>.log
func NewLogger(commandName string) (*zap.Logger, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	logDir := filepath.Join(home, ".idx-scraper", commandName)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", logDir, err)
	}

	logFile := filepath.Join(logDir, fmt.Sprintf("%s_%s.log", commandName, time.Now().Format("20060102_150405")))

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("02-01-2006T15:04:05"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// For CLI (console)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	// For File (console format as requested or preferred for readability)
	fileEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	// Open the log file
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", logFile, err)
	}

	// High-priority output (errors and above) to stderr
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zap.ErrorLevel
	})
	// Low-priority output (info and warn) to stdout
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zap.ErrorLevel && lvl >= zap.InfoLevel
	})

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), lowPriority),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stderr), highPriority),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(f), zap.InfoLevel),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	return logger, nil
}
