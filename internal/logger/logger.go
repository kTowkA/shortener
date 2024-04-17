package logger

import (
	"io"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func Init(output io.Writer, level logrus.Level) {
	Log = logrus.New()
	Log.SetFormatter(&logrus.TextFormatter{})
	Log.SetOutput(output)
	Log.SetLevel(level)
}

func LevelFromString(level string) logrus.Level {
	switch level {
	case "panic":
		return logrus.PanicLevel
	case "fatal":
		return logrus.FatalLevel
	case "error":
		return logrus.ErrorLevel
	case "warn":
		return logrus.WarnLevel
	case "info":
		return logrus.InfoLevel
	case "debug":
		return logrus.DebugLevel
	case "trace":
		return logrus.TraceLevel
	default:
		return logrus.InfoLevel
	}
}
