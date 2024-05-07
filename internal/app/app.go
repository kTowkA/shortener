package app

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/kTowkA/shortener/internal/config"
	"github.com/kTowkA/shortener/internal/logger"
	"github.com/kTowkA/shortener/internal/storage"
	"github.com/kTowkA/shortener/internal/storage/memory"
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
	storage, err := memory.NewStorage(s.Config.FileStoragePath)
	if err != nil {
		return fmt.Errorf("запуск сервера. %w", err)
	}
	defer storage.Close()
	s.db = storage

	mux := chi.NewRouter()

	mux.Use(withLog, withGZIP)

	mux.Route("/", func(r chi.Router) {
		r.Post("/", s.encodeURL)
		r.Get("/{short}", s.decodeURL)
		r.Route("/api", func(r chi.Router) {
			r.Post("/shorten", s.apiShorten)
		})
		r.Get("/ping", s.ping)
	})

	logger.Log.Info("запуск сервера", zap.String("адрес", s.Config.Address))
	return http.ListenAndServe(s.Config.Address, mux)
}
