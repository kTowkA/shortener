package logger

import (
	"fmt"
	"log/slog"

	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
)

// Log структура логгера с логером типа slog и бэкендом в виде zap
type Log struct {
	*slog.Logger
	zl *zap.Logger
}

// NewLogger создает новый логер с уровнем level
func NewLogger(level slog.Level) (*Log, error) {
	zc := zap.NewProductionConfig()
	zc.OutputPaths = []string{
		"stdout",
	}
	zc.Encoding = "json"

	if level != slog.LevelInfo {
		switch level {
		case slog.LevelWarn:
			zc.Level.SetLevel(zap.WarnLevel)
		case slog.LevelDebug:
			zc.Level.SetLevel(zap.DebugLevel)
		case slog.LevelError:
			zc.Level.SetLevel(zap.ErrorLevel)
		}
	}

	l, err := zc.Build()
	if err != nil {
		return nil, fmt.Errorf("создание логера. сборка zap логера. %w", err)
	}
	return &Log{
		Logger: slog.New(zapslog.NewHandler(l.Core(), nil)),
		zl:     l,
	}, nil
}

// Close закрытие логера
func (l *Log) Close() {
	// На ошибку не проверяем, там было открыто issue и разработчики советовали пропустить, на unix-системах возвращает не nil
	_ = l.zl.Sync()
}
