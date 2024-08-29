package app

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/kTowkA/shortener/internal/model"
	"github.com/kTowkA/shortener/internal/storage"
	"github.com/kTowkA/shortener/internal/utils"
)

// encodeURL обработчик для кодирования входящего урла
func (s *Server) encodeURL(w http.ResponseWriter, r *http.Request) {
	// проверяем, что контент тайп нужный
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") && !strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-gzip") {
		http.Error(w, fmt.Sprintf("разрешенные типы контента: %v", []string{"text/plain", "application/x-gzip"}), http.StatusBadRequest)
		return
	}

	// проверяем, что тело существует
	if r.Body == nil {
		http.Error(w, "пустой запрос", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// читаем тело
	sc := bufio.NewScanner(r.Body)
	link := ""
	for sc.Scan() {
		link = strings.TrimSpace(sc.Text())
		if link != "" {
			break
		}
	}

	// проверяем, что запрос не пуст
	if link == "" {
		http.Error(w, "пустой запрос", http.StatusBadRequest)
		return
	}

	// проверяем, что это ссылка
	_, err := url.ParseRequestURI(link)
	if err != nil {
		http.Error(w, "невалидная ссылка", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(contextKey("userID")).(uuid.UUID)
	if !ok {
		userID = uuid.New()
	}

	w.Header().Set("Content-Type", "text/plain")

	newLink, err := utils.SaveLink(r.Context(), s.db, userID, link)
	if err != nil {
		if errors.Is(err, storage.ErrURLConflict) {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(s.Config.BaseAddress() + newLink))
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(s.Config.BaseAddress() + newLink))
}

// decodeURL обработчик для декодирования короткой ссылки
func (s *Server) decodeURL(w http.ResponseWriter, r *http.Request) {

	// проверяем что есть подзапрос
	short := r.URL.Path
	if short == "/" {
		http.Error(w, "пустой подзапрос", http.StatusBadRequest)
		return
	}

	short = strings.Trim(short, "/")
	real, err := s.db.RealURL(r.Context(), short)
	// ничего не нашли
	if errors.Is(err, storage.ErrURLNotFound) {
		http.Error(w, storage.ErrURLNotFound.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if real.IsDeleted {
		w.WriteHeader(http.StatusGone)
		return
	}
	// успешно
	w.Header().Set("Location", real.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// apiShorten обработчик для API
func (s *Server) apiShorten(w http.ResponseWriter, r *http.Request) {

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
	conflict := false

	userID, ok := r.Context().Value(contextKey("userID")).(uuid.UUID)
	if !ok {
		userID = uuid.New()
	}
	// newLink, err := s.saveLink(r.Context(), userID, req.URL, attems)
	newLink, err := utils.SaveLink(r.Context(), s.db, userID, req.URL)
	if errors.Is(err, storage.ErrURLConflict) {
		conflict = true
	}
	if err != nil && !conflict {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := model.ResponseShortURL{
		Result: s.Config.BaseAddress() + newLink,
	}
	resp, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	if conflict {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write(resp)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(resp)
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request) {
	err := s.db.Ping(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) batch(w http.ResponseWriter, r *http.Request) {

	// проверяем, что тело существует
	if r.Body == nil {
		http.Error(w, "пустой запрос", http.StatusBadRequest)
		s.logger.Error("пустой запрос")
		return
	}

	// работаем с телом ответа
	buf := bytes.Buffer{}
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req := model.BatchRequest{}
	err = json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		s.logger.Error("Unmarshal")
		return
	}
	req = utils.ValidateAndGenerateBatch(req)
	// проверяем, что есть запросы
	if len(req) == 0 {
		http.Error(w, fmt.Errorf("передали пустой batch").Error(), http.StatusBadRequest)
		s.logger.Error("пустой batch")
		return
	}
	userID, ok := r.Context().Value(contextKey("userID")).(uuid.UUID)
	if !ok {
		userID = uuid.New()
	}
	resp, err := utils.SaveBatch(r.Context(), s.db, userID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for i := range resp {
		if resp[i].ShortURL != "" {
			resp[i].ShortURL = s.Config.BaseAddress() + resp[i].ShortURL
		}
	}

	result, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(result)
}
func (s *Server) getUserURLs(w http.ResponseWriter, r *http.Request) {

	// проверяем, что userID записан в cookie
	token, err := r.Cookie(authCookie)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if errors.Is(err, http.ErrNoCookie) {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userID, err := getUserIDFromToken(token.Value, s.Config.SecretKey())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	urls, err := s.db.UserURLs(r.Context(), userID)
	if errors.Is(err, storage.ErrURLNotFound) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	for i := range urls {
		urls[i].ShortURL = s.Config.BaseAddress() + urls[i].ShortURL
	}

	result, err := json.MarshalIndent(urls, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result)
}
func (s *Server) deleteUserURLs(w http.ResponseWriter, r *http.Request) {

	// проверяем, что userID записан в cookie
	token, err := r.Cookie(authCookie)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if errors.Is(err, http.ErrNoCookie) {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userID, err := getUserIDFromToken(token.Value, s.Config.SecretKey())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// проверяем, что тело существует
	if r.Body == nil {
		http.Error(w, "пустой запрос", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// работаем с телом ответа
	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req := []string{}
	err = json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for i := range req {
		s.deleteMessage <- model.DeleteURLMessage{
			UserID:   userID.String(),
			ShortURL: req[i],
		}
	}
	w.WriteHeader(http.StatusAccepted)
}
func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.db.Stats(r.Context())
	if err != nil {
		s.logger.Error("запрос статистики сервиса", slog.String("ошибка", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	result, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		s.logger.Error("конвертация в json", slog.String("ошибка", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result)
}
