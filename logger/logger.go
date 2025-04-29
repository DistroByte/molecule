package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Log zerolog.Logger

func InitLogger(level zerolog.Level) {
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if level == zerolog.DebugLevel {
		Log = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		Log.Info().Msg("running locally in debug mode")
	} else {
		Log = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}
}
