package util

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var log zerolog.Logger

type Level uint8

const (
	DebugLevle Level = iota

	InfoLevel

	WarnLevel

	ErrorLevel

	FatalLEvel

	PanicLevel

	NoLevel

	Disabled
)

func init() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	log = zerolog.New(output).With().Timestamp().Logger()
	SetLevel(DebugLevle)
}

func SetLevel(l Level) {
	zerolog.SetGlobalLevel(zerolog.Level(l))
}

func Infof(format string, v ...interface{}) {
	log.Info().Msgf(format, v...)
}

func Debugf(format string, v ...interface{}) {
	log.Debug().Msgf(format, v...)
}

func Warnf(format string, v ...interface{}) {
	log.Warn().Msgf(format, v...)
}

func Errorf(format string, v ...interface{}) {
	log.Error().Msgf(format, v...)
}

func Panicf(format string, v ...interface{}) {
	log.Panic().Msgf(format, v...)
}
