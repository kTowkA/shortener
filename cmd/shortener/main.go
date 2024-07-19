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
	"github.com/kTowkA/shortener/internal/storage"
	"github.com/kTowkA/shortener/internal/storage/memory"
	"github.com/kTowkA/shortener/internal/storage/postgres"
	"github.com/kTowkA/shortener/internal/storage/postgres/migrations"
)

func main() {
	// логгер
	customLog, err := logger.NewLogger(slog.LevelDebug)
	if err != nil {
		log.Fatalf("создание пользовательского логгера: %v", err)
	}
	defer customLog.Close()

	// конфигурация
	cfg, err := config.ParseConfig(customLog.Logger)
	if err != nil {
		customLog.Error("загрузка конфигурации", slog.String("ошибка", err.Error()))
		return
	}

	var myStorage storage.Storager

	// хранилище
	if cfg.DatabaseDSN() != "" {
		err = migrations.MigrationsUP(cfg.DatabaseDSN())
		if err != nil {
			customLog.Error("проведение миграций", slog.String("ошибка", err.Error()))
			return
		}
		myStorage, err = postgres.NewStorage(context.Background(), cfg.DatabaseDSN())
	} else {
		myStorage, err = memory.NewStorage(cfg.FileStoragePath())
	}
	if err != nil {
		customLog.Error("создание хранилища данных", slog.String("ошибка", err.Error()))
		return
	}
	defer myStorage.Close()

	// приложение
	srv, err := app.NewServer(cfg, customLog.Logger)
	if err != nil {
		customLog.Error("создание сервера приложения", slog.String("ошибка", err.Error()))
		return
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err = srv.Run(ctx, myStorage); err != nil {
		customLog.Error("запуск сервера приложения", slog.String("ошибка", err.Error()))
	}
}
