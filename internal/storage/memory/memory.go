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

	"github.com/google/uuid"
	"github.com/kTowkA/shortener/internal/logger"
	"github.com/kTowkA/shortener/internal/model"
	"github.com/kTowkA/shortener/internal/storage"
	"go.uber.org/zap"
)

type Storage struct {
	pairs map[string]model.StorageJSONWithUserID
	sync.Mutex
	storageFile string
}

func NewStorage(storageFile string) (*Storage, error) {
	var (
		links map[string]model.StorageJSONWithUserID
		err   error
	)
	if storageFile != "" {
		links, err = restoreFromFile(storageFile)
		if err != nil {
			return nil, fmt.Errorf("создание хранилища. %w", err)
		}
	}
	if links == nil {
		links = make(map[string]model.StorageJSONWithUserID)
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

func (s *Storage) SaveURL(ctx context.Context, userID uuid.UUID, real, short string) (string, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	if _, ok := s.pairs[short]; ok {
		return "", storage.ErrURLIsExist
	}
	if oldShort := s.findShortURL(real, userID); oldShort != "" {
		return oldShort, storage.ErrURLConflict
	}

	s.pairs[short] = model.StorageJSONWithUserID{
		UserID: userID.String(),
		StorageJSON: model.StorageJSON{
			UUID:        uuid.New().String(),
			ShortURL:    short,
			OriginalURL: real,
		},
	}
	if s.storageFile == "" {
		return short, nil
	}
	err := savelToFile(s.storageFile, []model.StorageJSONWithUserID{s.pairs[short]}, os.O_WRONLY|os.O_CREATE|os.O_APPEND)
	if err != nil {
		return "", fmt.Errorf("сохранение результатов в файл. %w", err)
	}
	return short, nil
}

func (s *Storage) RealURL(ctx context.Context, short string) (model.StorageJSON, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	if real, ok := s.pairs[short]; ok {
		return model.StorageJSON{
			OriginalURL: real.OriginalURL,
			IsDeleted:   real.IsDeleted,
		}, nil
	}
	return model.StorageJSON{}, storage.ErrURLNotFound
}

func (s *Storage) Ping(ctx context.Context) error {
	return nil
}

func restoreFromFile(filename string) (map[string]model.StorageJSONWithUserID, error) {
	file, err := os.Open(filename)

	if err != nil && strings.Contains(err.Error(), "no such file or directory") {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("восстановление записей из файла. %w", err)
	}
	elements := make(map[string]model.StorageJSONWithUserID)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		raw := scanner.Bytes()
		if len(raw) == 0 {
			continue
		}
		element := model.StorageJSONWithUserID{}
		err = json.Unmarshal(raw, &element)
		if err != nil {
			logger.Log.Error("раскодирование элемента", zap.Error(err))
			continue
		}
		elements[element.ShortURL] = element
	}
	return elements, nil
}

func (s *Storage) Batch(ctx context.Context, userID uuid.UUID, values model.BatchRequest) (model.BatchResponse, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	result := make([]model.BatchResponseElement, 0, len(values))
	valuesForFile := make([]model.StorageJSONWithUserID, 0, len(values))
	for _, v := range values {
		e := model.BatchResponseElement{
			CorrelationID: v.CorrelationID,
			OriginalURL:   v.OriginalURL,
		}
		if _, ok := s.pairs[v.ShortURL]; ok {
			e.Collision = true
			e.Error = storage.ErrURLIsExist
		} else {
			if oldShort := s.findShortURL(v.OriginalURL, userID); oldShort != "" {
				e.Error = storage.ErrURLConflict
				e.ShortURL = oldShort
			} else {
				s.pairs[v.ShortURL] = model.StorageJSONWithUserID{
					UserID: userID.String(),
					StorageJSON: model.StorageJSON{
						UUID:        uuid.New().String(),
						ShortURL:    v.ShortURL,
						OriginalURL: v.OriginalURL,
					},
				}

				valuesForFile = append(valuesForFile, s.pairs[v.ShortURL])
			}
		}

		result = append(result, e)
	}

	if s.storageFile == "" {
		return result, nil
	}
	err := savelToFile(s.storageFile, valuesForFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND)
	if err != nil {
		return nil, fmt.Errorf("сохранение результатов в файл. %w", err)
	}
	return result, nil
}

func (s *Storage) UserURLs(ctx context.Context, userID uuid.UUID) ([]model.StorageJSON, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	results := make([]model.StorageJSON, 0)
	for _, v := range s.pairs {
		if v.UserID != userID.String() {
			continue
		}
		results = append(results, v.StorageJSON)
	}
	if len(results) == 0 {
		return nil, storage.ErrURLNotFound
	}
	return results, nil
}
func (s *Storage) DeleteURLs(ctx context.Context, deleteLinks []model.DeleteURLMessage) error {
	s.Mutex.Lock()
	change := false
	for _, v := range deleteLinks {
		if val, ok := s.pairs[v.ShortURL]; ok && val.UserID == v.UserID {
			val.IsDeleted = true
			s.pairs[v.ShortURL] = val
			change = true
		}
	}
	s.Mutex.Unlock()

	if s.storageFile == "" || !change {
		return nil
	}

	return s.rewriteFile()
}

// findShortURL ищем короткую ссылку (добавили когда ввели функционал с 409 ошибкой)
func (s *Storage) findShortURL(real string, userID uuid.UUID) string {
	// не блокируем mutex так как вызываем только в служебных целях
	for k, v := range s.pairs {
		if v.OriginalURL == real && userID.String() == v.UserID {
			return k
		}
	}
	return ""
}

// savelToFile сохранение в файле
func savelToFile(fileName string, values []model.StorageJSONWithUserID, flag int) error {
	file, err := os.OpenFile(fileName, flag, 0666)
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

// rewriteFile перезаписываем файл
func (s *Storage) rewriteFile() error {
	s.Mutex.Lock()
	values := make([]model.StorageJSONWithUserID, 0, len(s.pairs))
	for _, v := range s.pairs {
		values = append(values, v)
	}
	s.Mutex.Unlock()
	err := savelToFile(s.storageFile, values, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("перезапись файла. %w", err)
	}
	return nil
}
