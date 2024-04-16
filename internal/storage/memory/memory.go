package memory

import (
	"context"
	"sync"

	"github.com/kTowkA/shortener/internal/storage"
)

type Storage struct {
	pairs map[string]string
	sync.Mutex
}

func NewStorage() *Storage {
	return &Storage{
		pairs: make(map[string]string),
		Mutex: sync.Mutex{},
	}
}

func (s *Storage) SaveURL(ctx context.Context, real, short string) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	if _, ok := s.pairs[short]; ok {
		return storage.ErrURLIsExist
	}
	s.pairs[short] = real
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
