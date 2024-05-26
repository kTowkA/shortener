package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kTowkA/shortener/internal/config"
	"github.com/kTowkA/shortener/internal/logger"
	"github.com/kTowkA/shortener/internal/model"
	"github.com/kTowkA/shortener/internal/storage"
	"github.com/kTowkA/shortener/internal/storage/memory"
	"github.com/kTowkA/shortener/internal/storage/postgres"
	"go.uber.org/zap"
)

const (

	// attems количество попыток генерации
	attems = 10

	// defaultLenght длина по умолчанию
	defaultLenght = 10
)

type Server struct {
	db            storage.Storager
	Config        config.Config
	deleteMessage chan model.DeleteURLMessage
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
		Config:        cfg,
		db:            storage,
		deleteMessage: make(chan model.DeleteURLMessage, 100),
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
	defer close(s.deleteMessage)

	go s.flushDeleteMessages()
	s.db = myStorage

	mux := chi.NewRouter()

	mux.Use(withLog, withGZIP, s.withToken)

	mux.Route("/", func(r chi.Router) {
		r.Post("/", s.encodeURL)
		r.Get("/{short}", s.decodeURL)
		r.Route("/api", func(r chi.Router) {
			r.Route("/shorten", func(r chi.Router) {
				r.Post("/", s.apiShorten)
				r.Post("/batch", s.batch)
			})
			r.Route("/user", func(r chi.Router) {
				r.Get("/urls", s.getUserURLs)
				r.Delete("/urls", s.deleteUserURLs)
			})
		})
		r.Get("/ping", s.ping)
	})

	logger.Log.Info("запуск сервера", zap.String("адрес", s.Config.Address))
	return http.ListenAndServe(s.Config.Address, mux)
}
func (s *Server) flushDeleteMessages() {
	ticker := time.NewTicker(10 * time.Second)

	var messages []model.DeleteURLMessage

	for {
		select {
		case msg := <-s.deleteMessage:
			messages = append(messages, msg)
		case <-ticker.C:
			if len(messages) == 0 {
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			err := s.db.DeleteURLs(ctx, messages)
			cancel()
			if err != nil {
				logger.Log.Error("удаление сообщений", zap.Error(err))
			}
			messages = nil
		}
	}
}
