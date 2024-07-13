package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kTowkA/shortener/internal/app"
	"github.com/kTowkA/shortener/internal/config"
	"github.com/kTowkA/shortener/internal/logger"
)

func main() {
	// логгер
	customLog, err := logger.NewLogger(slog.LevelDebug)
	if err != nil {
		log.Fatalf("создание пользовательского логгера: %v", err)
	}
	defer customLog.Close()

	// конфигурация
	cfg, err := config.ParseConfig(customLog.WithGroup("конфигуратор"))
	if err != nil {
		customLog.Error("загрузка конфигурации", slog.String("ошибка", err.Error()))
		return
	}

	// приложение
	srv, err := app.NewServer(cfg, customLog.WithGroup("приложение"))
	if err != nil {
		customLog.Error("создание сервера приложения", slog.String("ошибка", err.Error()))
		return
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err = srv.Run(ctx); err != nil {
		customLog.Error("запуск сервера приложения", slog.String("ошибка", err.Error()))
	}
}
