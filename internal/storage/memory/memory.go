package memory

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/kTowkA/shortener/internal/logger"
	"github.com/kTowkA/shortener/internal/model"
	"github.com/kTowkA/shortener/internal/storage"
	"go.uber.org/zap"
)

type Storage struct {
	pairs map[string]string
	sync.Mutex
	storageFile string
}

func NewStorage(storageFile string) (*Storage, error) {
	var (
		links map[string]string
		err   error
	)
	if storageFile != "" {
		links, err = restoreFromFile(storageFile)
		if err != nil {
			return nil, fmt.Errorf("создание хранилища. %w", err)
		}
	}
	if links == nil {
		links = make(map[string]string)
	}
	return &Storage{
		pairs:       links,
		Mutex:       sync.Mutex{},
		storageFile: storageFile,
	}, nil
}
func (s *Storage) Close() error {
	return nil
}

func (s *Storage) SaveURL(ctx context.Context, real, short string) (string, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	if _, ok := s.pairs[short]; ok {
		return "", storage.ErrURLIsExist
	}
	if oldShort := s.findShortUrl(real); oldShort != "" {
		return oldShort, storage.ErrURLConflict
	}

	s.pairs[short] = real
	if s.storageFile == "" {
		return short, nil
	}
	err := savelToFile(s.storageFile, []model.StorageJSON{
		{
			UUID:        "",
			ShortURL:    short,
			OriginalURL: real,
		},
	},
	)
	if err != nil {
		return "", fmt.Errorf("сохранение результатов в файл. %w", err)
	}
	return short, nil
}

func (s *Storage) RealURL(ctx context.Context, short string) (string, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	if real, ok := s.pairs[short]; ok {
		return real, nil
	}
	return "", storage.ErrURLNotFound
}

func (s *Storage) Ping(ctx context.Context) error {
	return nil
}

func restoreFromFile(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	// if errors.Is(err, os.ErrNotExist) {
	if err != nil && strings.Contains(err.Error(), "no such file or directory") {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("восстановление записей из файла. %w", err)
	}
	elements := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		raw := scanner.Bytes()
		if len(raw) == 0 {
			continue
		}
		element := model.StorageJSON{}
		err = json.Unmarshal(raw, &element)
		if err != nil {
			logger.Log.Error("раскодирование элемента", zap.Error(err))
			continue
		}
		elements[element.ShortURL] = element.OriginalURL
	}
	return elements, nil
}

func (s *Storage) Batch(ctx context.Context, values model.BatchRequest) (model.BatchResponse, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	result := make([]model.BatchResponseElement, 0, len(values))
	valuesForFile := make([]model.StorageJSON, 0, len(values))
	for _, v := range values {
		e := model.BatchResponseElement{
			CorrelationID: v.CorrelationID,
			OriginalURL:   v.OriginalURL,
		}
		if _, ok := s.pairs[v.ShortURL]; ok {
			e.Collision = true
			e.Error = storage.ErrURLIsExist
		} else {
			if oldShort := s.findShortUrl(v.OriginalURL); oldShort != "" {
				e.Error = storage.ErrURLConflict
				e.ShortURL = oldShort
			} else {
				s.pairs[v.ShortURL] = v.OriginalURL

				valuesForFile = append(valuesForFile, model.StorageJSON{
					UUID:        "",
					ShortURL:    v.ShortURL,
					OriginalURL: v.OriginalURL,
				})
			}
		}

		result = append(result, e)
	}

	if s.storageFile == "" {
		return result, nil
	}
	err := savelToFile(s.storageFile, valuesForFile)
	if err != nil {
		return nil, fmt.Errorf("сохранение результатов в файл. %w", err)
	}
	return result, nil
}

// findShortUrl ищем короткую ссылку (добавили когда ввели функционал с 409 ошибкой)
func (s *Storage) findShortUrl(real string) string {
	// не блокируем mutex так как вызываем только в служебных целях
	for k, v := range s.pairs {
		if v == real {
			return k
		}
	}
	return ""
}

// savelToFile сохранение в файле
func savelToFile(fileName string, values []model.StorageJSON) error {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("открытие файла %s. %w", fileName, err)
	}
	defer file.Close()
	errs := make([]error, 0)
	for _, v := range values {
		body, err := json.Marshal(v)
		if err != nil {
			errs = append(errs, fmt.Errorf("кодирование в JSON %v. %w", v, err))
			continue
		}
		_, err = file.Write(body)
		if err != nil {
			errs = append(errs, fmt.Errorf("сохранение элемента в файле %v. %w", v, err))
			continue
		}
		_, err = file.Write([]byte("\n"))
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Join(errs...)
}
