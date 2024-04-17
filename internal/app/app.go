package app

import (
	"bufio"
	"context"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kTowkA/shortener/internal/config"
	"github.com/kTowkA/shortener/internal/logger"
	"github.com/kTowkA/shortener/internal/storage"
	"github.com/kTowkA/shortener/internal/storage/memory"
	"github.com/sirupsen/logrus"
)

const (
	plainTextContentType = "text/plain"
	contentType          = "content-type"

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
	return &Server{
		Config: cfg,
		db:     memory.NewStorage(),
	}, nil
}

func (s *Server) ListenAndServe() error {
	mux := chi.NewRouter()

	mux.Use(withLog)

	mux.Route("/", func(r chi.Router) {
		r.Post("/", s.encodeURL)
		r.Route("/{short}", func(r chi.Router) {
			r.Get("/", s.decodeURL)
		})
	})
	logger.Log.WithFields(logrus.Fields{
		"адрес": s.Config.Address,
	}).Info("запуск сервера")
	return http.ListenAndServe(s.Config.Address, mux)
}

// encodeURL обработчик для кодирования входящего урла
func (s *Server) encodeURL(w http.ResponseWriter, r *http.Request) {
	// проверяем что метод валиден
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// проверяем, что контент тайп нужный
	if !strings.HasPrefix(r.Header.Get(contentType), plainTextContentType) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// проверяем, что тело существует
	if r.Body == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// читаем тело
	link, err := bufio.NewReader(r.Body).ReadString('\n')
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	link = strings.TrimSpace(link)

	// проверяем, что запрос не пуст
	if link == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// проверяем, что это ссылка
	_, err = url.ParseRequestURI(link)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// создаем короткую ссылка за attems попыток генерации
	for i := 0; i < attems; i++ {
		genLink, err := generate(defaultLenght)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = s.db.SaveURL(context.Background(), link, genLink)
		// такая ссылка уже существует
		if errors.Is(err, storage.ErrURLIsExist) {
			continue
		} else if err != nil {
			// сюда попасть мы не можем, других ошибок не возвращаем пока, это на будущее
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// успешно
		w.Header().Set(contentType, plainTextContentType)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(s.Config.BaseAddress + genLink))

		return
	}
	// не уложись в заданное количество попыток для создания короткой ссылки
	http.Error(w, errors.New("не смогли создать короткую ссылку").Error(), http.StatusInternalServerError)
}

// decodeURL обработчик для декодирования короткой ссылки
func (s *Server) decodeURL(w http.ResponseWriter, r *http.Request) {
	// проверяем что метод валиден
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// проверяем что есть подзапрос
	short := strings.Trim(r.URL.Path, "/")
	if short == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	real, err := s.db.RealURL(context.Background(), short)
	// ничего не нашли
	if errors.Is(err, storage.ErrURLNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		// сюда попасть мы не можем, других ошибок не возвращаем пока, это на будущее
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// успешно
	w.Header().Set("Location", real)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// generate генерируем случайную строку
func generate(lenght int) (string, error) {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	b := strings.Builder{}
	b.Grow(lenght)

	for i := 0; i < lenght; i++ {
		_, err := b.WriteString(generateChars[r.Intn(len(generateChars))])
		if err != nil {
			return "", err
		}
	}
	return b.String(), nil
}
