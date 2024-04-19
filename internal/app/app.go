package app

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/kTowkA/shortener/internal/config"
	"github.com/kTowkA/shortener/internal/logger"
	"github.com/kTowkA/shortener/internal/storage"
	"github.com/kTowkA/shortener/internal/storage/memory"
	"github.com/sirupsen/logrus"
)

const (
	plainTextContentType       = "text/plain"
	applicationJSONContentType = "application/json"
	textHTMLContentType        = "text/html"
	applicationXGZIP           = "application/x-gzip"
	contentType                = "content-type"
	contentEncoding            = "content-encoding"
	acceptEncoding             = "accept-encoding"

	// attems количество попыток генерации
	attems = 10

	// defaultLenght длина по умолчанию
	defaultLenght = 7

	// avalChars доступные символы для генерации
	avalChars = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM"
)

var (
	generateChars = strings.Split(avalChars, "")
)

type Server struct {
	db     storage.Storager
	Config config.Config
}

func NewServer(cfg config.Config) (*Server, error) {
	logger.Init(os.Stdout, logger.LevelFromString(cfg.LogLevel))
	cfg.BaseAddress = strings.TrimSuffix(cfg.BaseAddress, "/") + "/"
	storage, err := memory.NewStorage(cfg.FileStoragePath)
	if err != nil {
		return nil, fmt.Errorf("запуск сервера. %w", err)
	}
	return &Server{
		Config: cfg,
		db:     storage,
	}, nil
}

func (s *Server) ListenAndServe() error {
	mux := chi.NewRouter()

	mux.Use(withLog, withGZIP)

	mux.Route("/", func(r chi.Router) {
		r.Post("/", s.encodeURL)
		r.Get("/{short}", s.decodeURL)
		r.Route("/api", func(r chi.Router) {
			r.Post("/shorten", s.apiShorten)
		})
	})

	logger.Log.WithFields(logrus.Fields{
		"адрес": s.Config.Address,
	}).Info("запуск сервера")
	return http.ListenAndServe(s.Config.Address, mux)
}
