package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/kTowkA/shortener/internal/config"
	"github.com/kTowkA/shortener/internal/logger"
	"github.com/kTowkA/shortener/internal/storage"
	"github.com/kTowkA/shortener/internal/storage/memory"
	"github.com/kTowkA/shortener/internal/storage/postgres"
	_ "github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

const (

	// attems количество попыток генерации
	attems = 10

	// defaultLenght длина по умолчанию
	defaultLenght = 7
)

type Server struct {
	db     storage.Storager
	Config config.Config
}

func NewServer(cfg config.Config) (*Server, error) {
	err := logger.New(logger.LevelFromString(cfg.LogLevel))
	if err != nil {
		return nil, fmt.Errorf("создание сервера. %w", err)
	}
	cfg.BaseAddress = strings.TrimSuffix(cfg.BaseAddress, "/") + "/"
	// не должно так быть, хранилище инициализируется при запуске ниже, но без этого не проходят тесты
	storage, err := memory.NewStorage("")
	if err != nil {
		return nil, fmt.Errorf("создание сервера. %w", err)
	}
	return &Server{
		Config: cfg,
		db:     storage,
	}, nil
}

func (s *Server) ListenAndServe() error {
	var (
		myStorage storage.Storager
		err       error
	)
	if s.Config.DatabaseDSN != "" {
		myStorage, err = postgres.NewStorage(context.Background(), s.Config.DatabaseDSN)
	} else {
		myStorage, err = memory.NewStorage(s.Config.FileStoragePath)
	}
	if err != nil {
		return fmt.Errorf("запуск сервера. %w", err)
	}
	defer myStorage.Close()

	s.db = myStorage

	mux := chi.NewRouter()

	mux.Use(withLog, withGZIP)

	mux.Route("/", func(r chi.Router) {
		r.Post("/", s.encodeURL)
		r.Get("/{short}", s.decodeURL)
		r.Route("/api", func(r chi.Router) {
			r.Route("/shorten", func(r chi.Router) {
				r.Post("/", s.apiShorten)
				r.Post("/batch", s.batch)
			})
		})
		r.Get("/ping", s.ping)
	})

	logger.Log.Info("запуск сервера", zap.String("адрес", s.Config.Address))
	return http.ListenAndServe(s.Config.Address, mux)
}
