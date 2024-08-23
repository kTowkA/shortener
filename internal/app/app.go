// модуль app отвечает за настройку и запуск сервера сокращателя ссылок.
// он поднимает сервер необходимые роуты, объявленные в задании
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
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sync/errgroup"
)

const (

	// attems количество попыток генерации
	attems = 10

	// defaultLenght длина по умолчанию
	defaultLenght = 10
)

// Server структура сервер служит для создания экземпляра запускаемого сервера.
// содержит только неэкспортируемые поля, включающие в себя хранилище данных, логгер и собственно сам веб-сервер
type Server struct {
	db            storage.Storager
	Config        config.Config
	deleteMessage chan model.DeleteURLMessage
	logger        *slog.Logger
	server        *http.Server
}

// NewServer создает новый экземпляр сервера с конфигурацией cfg и логером logger.
// Возвращает сервер и ошибку
func NewServer(cfg config.Config, logger *slog.Logger) (*Server, error) {
	s := &Server{
		Config: cfg,
		logger: logger,
		server: &http.Server{
			Addr: cfg.Address(),
		},
		deleteMessage: make(chan model.DeleteURLMessage, 100),
	}

	// включен HTTPS
	if s.Config.HTTPS() {
		// установлено доменное имя
		if s.Config.Domain() != "" {
			s.server.TLSConfig = (&autocert.Manager{
				HostPolicy: autocert.HostWhitelist(s.Config.Domain()),
				Prompt:     autocert.AcceptTOS,
			}).TLSConfig()
		}
	}
	return s, nil
}

// Run запуск сервера с указанием контекста для отмены ctx и хранилища storage
func (s *Server) Run(ctx context.Context, storage storage.Storager) error {
	s.db = storage

	s.setRoute()

	// создание группы
	gr, grCtx := errgroup.WithContext(ctx)

	// запуск приложения
	gr.Go(func() error {

		defer s.logger.Info("приложение остановлено")

		s.logger.Info("запуск приложения", slog.String("адрес", s.Config.Address()))

		var err error
		if s.Config.HTTPS() {
			err = s.server.ListenAndServeTLS("", "")
		} else {
			err = s.server.ListenAndServe()
		}

		if err != nil {
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

			r.Route("/internal", func(r chi.Router) {
				r.Use(s.trustedSubnet)
				r.Get("/stats", s.stats)
			})

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
