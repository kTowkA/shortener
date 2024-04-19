package app

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kTowkA/shortener/internal/model"
	"github.com/kTowkA/shortener/internal/storage"
)

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
	newLink, err := s.saveLink(link, attems)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set(contentType, plainTextContentType)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(newLink))
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

// apiShorten обработчик для API
func (s *Server) apiShorten(w http.ResponseWriter, r *http.Request) {
	// проверяем что метод валиден
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// проверяем, что контент тайп нужный
	if !strings.HasPrefix(r.Header.Get(contentType), applicationJsonContentType) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// проверяем, что тело существует
	if r.Body == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// работаем с телом ответа
	buf := bytes.Buffer{}
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req := model.RequestShortURL{}
	err = json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// if req.URL != "" {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	newLink, err := s.saveLink(req.URL, attems)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := model.ResponseShortURL{
		Result: newLink,
	}
	resp, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set(contentType, applicationJsonContentType)
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

func (s *Server) saveLink(link string, attems int) (string, error) {
	// создаем короткую ссылка за attems попыток генерации
	for i := 0; i < attems; i++ {
		genLink, err := generate(defaultLenght)
		if err != nil {
			return "", err
		}

		err = s.db.SaveURL(context.Background(), link, genLink)
		// такая ссылка уже существует
		if errors.Is(err, storage.ErrURLIsExist) {
			continue
		} else if err != nil {
			// сюда попасть мы не можем, других ошибок не возвращаем пока, это на будущее
			return "", err
		}

		// успешно
		return s.Config.BaseAddress + genLink, nil
	}

	// не уложись в заданное количество попыток для создания короткой ссылки
	return "", errors.New("не смогли создать короткую ссылку")
}
