package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Log zerolog.Logger
var level zerolog.Level

func InitLogger() {

	switch os.Getenv("LEVEL") {
	case "trace":
		level = zerolog.TraceLevel
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	case "fatal":
		level = zerolog.FatalLevel
	case "panic":
		level = zerolog.PanicLevel
	default:
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if level == zerolog.DebugLevel {
		Log = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		Log.Info().Msg("running locally in debug mode")
	} else {
		Log = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}
}
