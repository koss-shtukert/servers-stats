package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

func New(level string) (zerolog.Logger, error) {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// Disable console writer to ensure JSON format
	os.Setenv("NO_COLOR", "1")

	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.Logger{}, err
	}

	zerolog.SetGlobalLevel(logLevel)

	// Try multiple log directories in order of preference
	logDirs := []string{"/app/logs", "./logs", "/var/log/servers-stats"}
	var logDir string
	var dirErr error

	for _, dir := range logDirs {
		if err := os.MkdirAll(dir, 0755); err == nil {
			logDir = dir
			break
		} else {
			dirErr = err
		}
	}

	if logDir == "" {
		return zerolog.Logger{}, fmt.Errorf("failed to create log directory: %w", dirErr)
	}

	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "servers-stats.log"),
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	}

	multiWriter := io.MultiWriter(os.Stdout, fileWriter)

	logger := zerolog.New(multiWriter).
		With().
		Timestamp().
		Logger()

	zerolog.DefaultContextLogger = &logger

	return logger, nil
}
