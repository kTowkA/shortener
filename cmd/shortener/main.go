package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/kTowkA/shortener/internal/app"
	"github.com/kTowkA/shortener/internal/config"
	gapp "github.com/kTowkA/shortener/internal/grpc/app"
	"github.com/kTowkA/shortener/internal/logger"
	"github.com/kTowkA/shortener/internal/storage"
	"github.com/kTowkA/shortener/internal/storage/memory"
	"github.com/kTowkA/shortener/internal/storage/postgres"
	"github.com/kTowkA/shortener/internal/storage/postgres/migrations"
	"golang.org/x/sync/errgroup"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	// При указании флага линковщика -ldflags и указании флага -X можно собрать приложение с определенными значениями buildVersion,buildDate,buildCommit
	// -X main.buildVersion=версия
	fmt.Println(buildVersion)
	// -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')' для получения даты сборки
	fmt.Println(buildDate)
	// -X 'main.buildCommit=$(git show --oneline -s)' для получения коммита
	fmt.Println(buildCommit)

	// логгер
	customLog := initLog()
	defer customLog.Close()

	// конфигурация
	cfg, err := config.ParseConfig(customLog.Logger)
	if err != nil {
		customLog.Error("загрузка конфигурации", slog.String("ошибка", err.Error()))
		return
	}

	// хранилище
	myStorage, err := initStorage(cfg)
	if err != nil {
		customLog.Error("инициализация хранилища", slog.String("ошибка", err.Error()))
	}
	defer myStorage.Close()

	// приложение
	srv, err := app.NewServer(cfg, customLog.Logger)
	if err != nil {
		customLog.Error("создание сервера приложения", slog.String("ошибка", err.Error()))
		return
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()
	gr, _ := errgroup.WithContext(ctx)
	gr.Go(func() error {
		if err = srv.Run(ctx, myStorage); err != nil {
			customLog.Error("запуск сервера приложения", slog.String("ошибка", err.Error()))
			return err
		}
		return nil
	})
	gr.Go(func() error {
		if cfg.GRPC() == "" {
			return nil
		}
		if err = gapp.Run(ctx, myStorage, customLog.Logger, cfg.GRPC()); err != nil {
			customLog.Error("запуск gRPC-сервера приложения", slog.String("ошибка", err.Error()))
			return err
		}
		return nil
	})
	err = gr.Wait()
	if err != nil {
		customLog.Error("запуск группы", slog.String("ошибка", err.Error()))
	}
}

// инициализация логера
func initLog() *logger.Log {
	customLog, err := logger.NewLogger(slog.LevelDebug)
	if err != nil {
		log.Fatalf("создание пользовательского логгера: %v", err)
	}
	return customLog
}

// инициализация хранилища
func initStorage(cfg config.Config) (storage.Storager, error) {
	if cfg.DatabaseDSN() != "" {
		err := migrations.MigrationsUP(cfg.DatabaseDSN())
		if err != nil {
			return nil, fmt.Errorf("проведение миграций. %w", err)
		}
		return postgres.NewStorage(context.Background(), cfg.DatabaseDSN())
	}
	return memory.NewStorage(cfg.FileStoragePath())
}
