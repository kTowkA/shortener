package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kTowkA/shortener/internal/config"
	"github.com/kTowkA/shortener/internal/model"
	"github.com/kTowkA/shortener/internal/storage"
	"github.com/kTowkA/shortener/internal/storage/memory"
	"github.com/kTowkA/shortener/internal/storage/postgres"
	"golang.org/x/sync/errgroup"
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
	logger        *slog.Logger
	server        *http.Server
}

func NewServer(cfg config.Config, logger *slog.Logger) (*Server, error) {

	// не должно так быть, хранилище инициализируется при запуске ниже, но без этого не проходят тесты
	// storage, err := memory.NewStorage("")
	// if err != nil {
	// 	return nil, fmt.Errorf("создание сервера. %w", err)
	// }
	return &Server{
		Config: cfg,
		logger: logger,
		server: &http.Server{
			Addr: cfg.Address(),
		},
		// db:            storage,
		deleteMessage: make(chan model.DeleteURLMessage, 100),
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	err := s.setStorage()
	if err != nil {
		s.logger.Error("установка хранилища данных", slog.String("ошибка", err.Error()))
		return err
	}

	s.setRoute()

	// создание группы
	gr, grCtx := errgroup.WithContext(ctx)

	// обработка правильного завершения работы с хранилищем
	gr.Go(func() error {
		defer s.logger.Info("хранилище закрыто")
		<-grCtx.Done()
		return s.db.Close()
	})

	// запуск приложения
	gr.Go(func() error {

		defer s.logger.Info("приложение остановлено")

		s.logger.Info("запуск приложения", slog.String("адрес", s.Config.Address()))

		if err := s.server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				s.logger.Error("запуск сервера", slog.String("ошибка", err.Error()))
			}
		}

		return nil
	})

	// ожидание завершения работы приложения
	gr.Go(func() error {
		// закроем канал с сообщениями на удаление
		defer close(s.deleteMessage)

		// ожидаем отмены
		<-grCtx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// завершаем работу сервера
		if err := s.server.Shutdown(ctx); err != nil {
			s.logger.Error("завершение работы приложения", slog.String("ошибка", err.Error()))
			return err
		}

		return nil
	})

	// обработка сообщений на удаление
	gr.Go(func() error {
		s.flushDeleteMessages()
		return nil
	})

	return gr.Wait()
}

func (s *Server) setStorage() error {
	var (
		myStorage storage.Storager
		err       error
	)
	if s.Config.DatabaseDSN() != "" {
		myStorage, err = postgres.NewStorage(context.Background(), s.Config.DatabaseDSN())
	} else {
		myStorage, err = memory.NewStorage(s.Config.FileStoragePath())
	}
	if err != nil {
		return fmt.Errorf("создание хранилища данных. %w", err)
	}
	s.db = myStorage
	return nil
}

func (s *Server) setRoute() {
	mux := chi.NewRouter()

	mux.Use(s.withLog, withGZIP, s.withToken)

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
	s.server.Handler = mux
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
				s.logger.Error("удаление сообщений", slog.String("ошибка", err.Error()))
			}
			messages = nil
		}
	}
}
