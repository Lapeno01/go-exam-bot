package logger

import (
	"github.com/rs/zerolog"
	"os"
)

var log zerolog.Logger

func Init(logPath string) error {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	multi := zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{Out: os.Stdout},
		logFile,
	)
	log = zerolog.New(multi).With().Timestamp().Logger()
	return nil
}

func Info() *zerolog.Event {
	return log.Info()
}

func Error() *zerolog.Event {
	return log.Error()
}

func Warn() *zerolog.Event {
	return log.Warn()
}
