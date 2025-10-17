package log

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New(environment string) zerolog.Logger {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    environment == "production",
	}

	logger := zerolog.New(output).With().
		Timestamp().
		Str("env", environment).
		Logger()

	if environment != "production" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return logger
}
