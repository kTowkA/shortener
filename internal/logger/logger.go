package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger = zap.NewNop()

func New(level zapcore.Level) error {
	cfg := zap.NewProductionConfig()
	cfg.Level.SetLevel(level)
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}

func LevelFromString(level string) zapcore.Level {
	switch level {
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	case "error":
		return zapcore.ErrorLevel
	case "warn":
		return zapcore.WarnLevel
	case "info":
		return zapcore.InfoLevel
	case "debug":
		return zapcore.DebugLevel
	default:
		return zapcore.InfoLevel
	}
}
