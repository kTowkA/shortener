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

	"github.com/kTowkA/shortener/internal/logger"
	"github.com/kTowkA/shortener/internal/model"
	"github.com/kTowkA/shortener/internal/storage"
	"go.uber.org/zap"
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
	w.Header().Set("content-type", "text/plain")

	if err != nil {
		if errors.Is(err, storage.ErrURLConflict) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(newLink))
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
	conflict := false
	newLink, err := s.saveLink(r.Context(), req.URL, attems)
	if errors.Is(err, storage.ErrURLConflict) {
		conflict = true
	}
	if err != nil && !conflict {
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
	if conflict {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	w.Write(resp)
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request) {
	err := s.db.Ping(r.Context())
	if err != nil {
		logger.Log.Error("проверка на доступность БД", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) batch(w http.ResponseWriter, r *http.Request) {
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

	req := model.BatchRequest{}
	err = json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req = validateBatch(req)
	// проверяем, что есть запросы
	if len(req) == 0 {
		http.Error(w, fmt.Errorf("передали пустой batch").Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.saveBatch(r.Context(), req, attems)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if w.Header().Get("content-type") == "" {
		w.Header().Set("content-type", "application/json")
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(result)
}

func (s *Server) saveLink(ctx context.Context, link string, attems int) (string, error) {
	const forUnique = "X"
	genLink, err := generateSHA1(link, defaultLenght)
	if err != nil {
		return "", err
	}

	// создаем короткую ссылка за attems попыток генерации
	for i := 0; i < attems; i++ {
		genLink, err = s.db.SaveURL(ctx, link, genLink)
		// такая ссылка уже существует
		if errors.Is(err, storage.ErrURLIsExist) {
			genLink = genLink + forUnique
			continue
		} else if errors.Is(err, storage.ErrURLConflict) {
			return s.Config.BaseAddress + genLink, err
		} else if err != nil {
			return "", err
		}

		// успешно
		return s.Config.BaseAddress + genLink, nil
	}

	// не уложись в заданное количество попыток для создания короткой ссылки
	return "", errors.New("не смогли создать короткую ссылку")
}
func (s *Server) saveBatch(ctx context.Context, batch model.BatchRequest, attems int) (model.BatchResponse, error) {
	result := make([]model.BatchResponseElement, 0, len(batch))
	for i := 0; i < attems; i++ {
		err := generateLinksBatch(batch)
		if err != nil {
			return nil, err
		}
		resp, err := s.db.Batch(ctx, batch)
		if err != nil {
			return nil, err
		}
		// очищаем batch и будем складывать в него строки с коллизиями
		batch = make(model.BatchRequest, 0)
		for i := range resp {
			if resp[i].ShortURL != "" {
				resp[i].ShortURL = s.Config.BaseAddress + resp[i].ShortURL
				result = append(result, resp[i])
				continue
			}
			if resp[i].Error != nil {
				logger.Log.Error("ошибка в batch", zap.String("original_url", resp[i].OriginalURL), zap.Error(err))
			}
			// если были колизии пробуем сохранить с добавлением подстроки
			if resp[i].Collision {
				batch = append(batch, model.BatchRequestElement{
					CorrelationID: resp[i].CorrelationID,
					OriginalURL:   resp[i].OriginalURL,
					ShortURL:      resp[i].ShortURL,
				})
			}
		}
		if len(batch) == 0 {
			break
		}
	}
	// не смогли уложиться в заданное количество попыток для сохранения всех ссылок
	if len(batch) != 0 {
		logger.Log.Debug("не смогли уложиться в заданное количество попыток для сохранения всех ссылок", zap.Int("сохранено", len(result)), zap.Int("не сохранено", len(batch)))
		return nil, fmt.Errorf("не было сохранено %d записей", len(batch))
	}
	return result, nil
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

func validateBatch(batch model.BatchRequest) model.BatchRequest {
	newBatch := make([]model.BatchRequestElement, 0, len(batch))
	for _, v := range batch {
		v.OriginalURL = strings.TrimSpace(v.OriginalURL)
		if v.OriginalURL == "" {
			continue
		}
		if _, err := url.ParseRequestURI(v.OriginalURL); err != nil {
			continue
		}
		newBatch = append(newBatch, v)
	}
	return newBatch
}

func generateLinksBatch(batch model.BatchRequest) error {
	bonus := "X"
	for i := range batch {
		if batch[i].ShortURL != "" {
			batch[i].ShortURL += bonus
			continue
		}
		short, err := generateSHA1(batch[i].OriginalURL, defaultLenght)
		if err != nil {
			return err
		}
		batch[i].ShortURL = short
	}
	return nil
}
