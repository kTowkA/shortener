package app

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kTowkA/shortener/internal/model"
	"github.com/kTowkA/shortener/internal/storage"
)

// encodeURL обработчик для кодирования входящего урла
func (s *Server) encodeURL(w http.ResponseWriter, r *http.Request) {

	// проверяем, что контент тайп нужный
	if !strings.HasPrefix(r.Header.Get("content-type"), "text/plain") && !strings.HasPrefix(r.Header.Get("content-type"), "application/x-gzip") {
		http.Error(w, fmt.Sprintf("разрешенные типы контента: %v", []string{"text/plain", "application/x-gzip"}), http.StatusBadRequest)
		return
	}

	// проверяем, что тело существует
	if r.Body == nil {
		http.Error(w, "пустой запрос", http.StatusBadRequest)
		return
	}

	// читаем тело
	link, err := bufio.NewReader(r.Body).ReadString('\n')
	if err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	link = strings.TrimSpace(link)

	// проверяем, что запрос не пуст
	if link == "" {
		http.Error(w, "пустой запрос", http.StatusBadRequest)
		return
	}
	// проверяем, что это ссылка
	_, err = url.ParseRequestURI(link)
	if err != nil {
		http.Error(w, "невалидная ссылка", http.StatusBadRequest)
		return
	}
	newLink, err := s.saveLink(r.Context(), link, attems)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(newLink))
}

// decodeURL обработчик для декодирования короткой ссылки
func (s *Server) decodeURL(w http.ResponseWriter, r *http.Request) {

	// проверяем что есть подзапрос
	short := strings.Trim(r.URL.Path, "/")
	if short == "" {
		http.Error(w, "пустой подзапрос", http.StatusBadRequest)
		return
	}

	real, err := s.db.RealURL(r.Context(), short)
	// ничего не нашли
	if errors.Is(err, storage.ErrURLNotFound) {
		http.Error(w, storage.ErrURLNotFound.Error(), http.StatusNotFound)
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

	// проверяем, что контент тайп нужный
	if !strings.HasPrefix(r.Header.Get("content-type"), "application/json") && !strings.HasPrefix(r.Header.Get("content-type"), "application/x-gzip") {
		http.Error(w, fmt.Sprintf("разрешенные типы контента: %v", []string{"application/json", "application/x-gzip"}), http.StatusBadRequest)
		return
	}

	// проверяем, что тело существует
	if r.Body == nil {
		http.Error(w, "пустой запрос", http.StatusBadRequest)
		return
	}

	// работаем с телом ответа
	buf := bytes.Buffer{}
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req := model.RequestShortURL{}
	err = json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// проверяем, что это ссылка
	_, err = url.ParseRequestURI(req.URL)
	if err != nil {
		http.Error(w, "невалидная ссылка", http.StatusBadRequest)
		return
	}
	newLink, err := s.saveLink(r.Context(), req.URL, attems)
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
	if w.Header().Get("content-type") == "" {
		w.Header().Set("content-type", "application/json")
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

func (s *Server) saveLink(ctx context.Context, link string, attems int) (string, error) {
	const forUnique = "X"
	genLink, err := generateSHA1(link, defaultLenght)
	if err != nil {
		return "", err
	}

	// создаем короткую ссылка за attems попыток генерации
	for i := 0; i < attems; i++ {
		err = s.db.SaveURL(ctx, link, genLink)
		// такая ссылка уже существует
		if errors.Is(err, storage.ErrURLIsExist) {
			genLink = genLink + forUnique
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

func generateSHA1(original string, lenght int) (string, error) {
	// получаем sha1
	sum := sha1.Sum([]byte(original))

	// получаем массив символов длиной lenght
	symbols := strings.Split(hex.EncodeToString(sum[:]), "")[:lenght]

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// тут случайный символ переводим в верхней регистр (чтобы избежать совпадений)
	result := strings.Builder{}
	for _, s := range symbols {
		change := r.Intn(2)
		if change == 1 {
			s = strings.ToUpper(s)
		}
		_, err := result.WriteString(s)
		if err != nil {
			return "", err
		}
	}
	return result.String(), nil
}
