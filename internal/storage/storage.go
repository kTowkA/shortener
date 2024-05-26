package storage

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/kTowkA/shortener/internal/model"
)

var (
	ErrURLNotFound = errors.New("URL не найден")
	ErrURLConflict = errors.New("оригинальный URL уже был добавлен")
	ErrURLIsExist  = errors.New("такой ключ занят")
)

type Storager interface {
	// SaveURL сохраняет пару real-short url
	SaveURL(ctx context.Context, userID uuid.UUID, real, short string) (string, error)

	// Batch пакетное сохранение всех значений values
	Batch(ctx context.Context, userID uuid.UUID, values model.BatchRequest) (model.BatchResponse, error)

	// RealURL получение оригинального url
	RealURL(ctx context.Context, short string) (model.StorageJSON, error)

	// UserURLs получает все записи сохраненные пользователем
	UserURLs(ctx context.Context, userID uuid.UUID) ([]model.StorageJSON, error)

	// DeleteURLs удаляет записи сохраненные пользователями
	DeleteURLs(ctx context.Context, deleteLinks []model.DeleteURLMessage) error

	// Ping проверка доступности хранилища
	Ping(ctx context.Context) error

	// Close закрытие хранилища
	Close() error
}
