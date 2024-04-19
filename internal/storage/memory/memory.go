package memory

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/kTowkA/shortener/internal/logger"
	"github.com/kTowkA/shortener/internal/model"
	"github.com/kTowkA/shortener/internal/storage"
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
func (s *Storage) SaveURL(ctx context.Context, real, short string) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	if _, ok := s.pairs[short]; ok {
		return storage.ErrURLIsExist
	}
	s.pairs[short] = real
	if s.storageFile == "" {
		return nil
	}
	file, err := os.OpenFile(s.storageFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("сохранение в файл. %w", err)
	}
	defer file.Close()
	element := model.StorageJSON{
		UUID:        "",
		ShortURL:    short,
		OriginalURL: real,
	}
	body, err := json.Marshal(element)
	if err != nil {
		return fmt.Errorf("сохранение елемента в файле. %w", err)
	}
	_, err = file.Write(body)
	if err != nil {
		return fmt.Errorf("сохранение елемента в файле. %w", err)
	}
	_, err = file.Write([]byte("\n"))
	if err != nil {
		return fmt.Errorf("сохранение елемента в файле. %w", err)
	}
	return nil
}

func (s *Storage) RealURL(ctx context.Context, short string) (string, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	if real, ok := s.pairs[short]; ok {
		return real, nil
	}
	return "", storage.ErrURLNotFound
}

func restoreFromFile(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	logger.Log.Info(err)
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
			logger.Log.Error("раскодирование элемента", "ошибка", err)
			continue
		}
		elements[element.ShortURL] = element.OriginalURL
	}
	return elements, nil
}
