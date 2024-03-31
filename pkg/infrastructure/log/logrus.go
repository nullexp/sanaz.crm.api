package log

import (
	"errors"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

const LogLevel = "TRACE"

var LoggerInstance *Logger

type Logger struct {
	*logrus.Logger
}

func NewLog(logLevel string) *Logger {
	logger, err := stdoutInit(logLevel)
	logger.SetReportCaller(true)
	if err != nil {
		logrus.Panic(err)
	}

	return &Logger{logger}
}

func stdoutInit(lvl string) (*logrus.Logger, error) {
	var err error
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	level, err := logrus.ParseLevel(lvl)
	if err != nil {
		err = errors.New("failed to parse level")
		return nil, err
	}
	logger.Level = level
	var logWriter io.Writer = os.Stdout
	logger.SetOutput(logWriter)

	return logger, err
}
