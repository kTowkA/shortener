package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PStorage struct {
	*pgxpool.Pool
}

func NewStorage(ctx context.Context, dsn string) (*PStorage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("создание клиента postgres. %w", err)
	}
	return &PStorage{pool}, nil
}

// реализация интерфейса Storager
func (p *PStorage) SaveURL(ctx context.Context, real, short string) error {
	return nil
}
func (p *PStorage) RealURL(ctx context.Context, short string) (string, error) {
	return "", nil
}
func (p *PStorage) Ping(ctx context.Context) error {
	return p.Pool.Ping(ctx)
}
func (p *PStorage) Close() error {
	p.Pool.Close()
	return nil
}
