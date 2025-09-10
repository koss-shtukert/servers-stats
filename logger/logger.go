package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

func New(level string) (zerolog.Logger, error) {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = time.RFC3339Nano

	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.Logger{}, err
	}

	zerolog.SetGlobalLevel(logLevel)

	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger()

	zerolog.DefaultContextLogger = &logger

	return logger, nil
}
