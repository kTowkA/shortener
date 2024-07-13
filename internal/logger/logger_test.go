package logger

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLog(t *testing.T) {
	l, err := NewLogger(slog.LevelError)
	assert.NoError(t, err)
	l.Close()
	l, err = NewLogger(slog.LevelInfo)
	assert.NoError(t, err)
	l.Close()
	l, err = NewLogger(slog.LevelWarn)
	assert.NoError(t, err)
	l.Close()
	l, err = NewLogger(slog.LevelDebug)
	assert.NoError(t, err)
	l.Close()
}
