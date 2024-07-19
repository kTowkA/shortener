package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	_ "net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kTowkA/shortener/internal/config"
	"github.com/kTowkA/shortener/internal/model"
	"github.com/kTowkA/shortener/internal/storage"
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
	return &Server{
		Config: cfg,
		logger: logger,
		server: &http.Server{
			Addr: cfg.Address(),
		},
		deleteMessage: make(chan model.DeleteURLMessage, 100),
	}, nil
}

func (s *Server) Run(ctx context.Context, storage storage.Storager) error {
	s.db = storage

	s.setRoute()

	// создание группы
	gr, grCtx := errgroup.WithContext(ctx)

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

	go s.flushDeleteMessages()

	return gr.Wait()
}

func (s *Server) setRoute() {
	mux := chi.NewRouter()

	mux.Use(s.withLog, withGZIP, s.withToken)

	mux.Route("/", func(r chi.Router) {
		r.Post("/", s.encodeURL)
		r.Get("/{short}", s.decodeURL)
		r.Route("/api", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(middleware.AllowContentType("application/json", "application/x-gzip"))
				r.Route("/shorten", func(r chi.Router) {
					r.Post("/", s.apiShorten)
					r.Post("/batch", s.batch)
				})
				r.Delete("/user/urls", s.deleteUserURLs)
			})
			r.Get("/user/urls", s.getUserURLs)
		})
		r.Get("/ping", s.ping)
	})
	mux.Mount("/debug", middleware.Profiler())
	s.server.Handler = mux
}

func (s *Server) flushDeleteMessages() {
	ticker := time.NewTicker(5 * time.Second)

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
