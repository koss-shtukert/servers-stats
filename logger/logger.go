package logger

import (
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

	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.Logger{}, err
	}

	zerolog.SetGlobalLevel(logLevel)

	logDir := "/app/logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return zerolog.Logger{}, err
	}

	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "speedtest.log"),
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
