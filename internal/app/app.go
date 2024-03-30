package app

import (
	"bufio"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kTowkA/shortener/internal/config"
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
	db     DB
	Config config.Config
}

type DB struct {
	pairs map[string]string
	sync.Mutex
}

func NewServer(cfg config.Config) (*Server, error) {
	cfg.BaseAddress = strings.TrimPrefix(cfg.BaseAddress, "/") + "/"
	return &Server{
		Config: cfg,
		db: DB{
			pairs: make(map[string]string),
			Mutex: sync.Mutex{},
		},
	}, nil
}

func (s *Server) ListenAndServe() error {
	mux := chi.NewRouter()
	mux.Route("/", func(r chi.Router) {
		r.Post("/", s.encodeURL)
		r.Route("/{short}", func(r chi.Router) {
			r.Get("/", s.decodeURL)
		})
	})
	// mux.HandleFunc("/", s.rootHandle)

	return http.ListenAndServe(s.Config.Address, mux)
}

// rootHandle стандартный обработчик
// func (s *Server) rootHandle(w http.ResponseWriter, r *http.Request) {
// 	switch r.Method {
// 	case http.MethodPost:
// 		s.encodeURL(w, r)
// 		return
// 	case http.MethodGet:
// 		s.decodeURL(w, r)
// 	}
// 	w.WriteHeader(http.StatusMethodNotAllowed)
// }

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

		genLink = "http://" + r.Host + "/" + genLink

		s.db.Mutex.Lock()
		_, ok := s.db.pairs[genLink]
		s.db.Mutex.Unlock()

		// такая ссылка уже существует
		if ok {
			continue
		}

		s.db.Mutex.Lock()
		s.db.pairs[genLink] = link
		s.db.Mutex.Unlock()

		// успешно
		w.Header().Set(contentType, plainTextContentType)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(genLink))

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
	path := strings.Trim(r.URL.Path, "/")
	// path := chi.URLParam(r, "short")
	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// http://localhost/771249
	// http://localhost/http://localhost/771249
	// получаем короткую ссылку (если нет в урле /, то считаем, что передали короткую ссылку)
	sl := strings.Split(path, "/")
	short := "http://" + r.Host + "/" + path
	if len(sl) > 1 {
		short = path
	}
	s.db.Mutex.Lock()
	real, ok := s.db.pairs[short]
	s.db.Mutex.Unlock()

	// ничего не нашли
	if !ok {
		http.NotFound(w, r)
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
