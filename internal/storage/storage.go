package storage

import (
	"context"
	"errors"
)

var (
	ErrURLNotFound = errors.New("URL  не найден")
	ErrURLIsExist  = errors.New("такой ключ занят")
)

type Storager interface {
	SaveURL(ctx context.Context, real, short string) error
	RealURL(ctx context.Context, short string) (string, error)
	Ping(ctx context.Context) error
	Close() error
}
