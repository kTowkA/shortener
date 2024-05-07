package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
	storage := &PStorage{pool}
	err = storage.initDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("инициализация БД таблицами. %w", err)
	}
	return storage, nil
}

func (p *PStorage) initDatabase(ctx context.Context) error {
	p.Exec(
		ctx,
		`
		CREATE TABLE IF NOT EXISTS url_list (
			uuid uuid,
			short_url text,
			original_url text,
			PRIMARY KEY(uuid),
			UNIQUE(short_url),
			UNIQUE(original_url)
		);		
		`,
	)
	return nil
}

// реализация интерфейса Storager
func (p *PStorage) SaveURL(ctx context.Context, real, short string) error {
	_, err := p.Exec(
		ctx,
		"INSERT INTO url_list(uuid,original_url,short_url) VALUES($1,$2,$3)",
		uuid.New(),
		real,
		short,
	)
	if err != nil {
		return fmt.Errorf("добавление новой записи в БД. %w", err)
	}
	return nil
}
func (p *PStorage) RealURL(ctx context.Context, short string) (string, error) {
	var real string
	err := p.QueryRow(
		ctx,
		"SELECT original_url FROM url_list WHERE short_url=$1",
		short,
	).Scan(&real)
	if err != nil {
		return "", fmt.Errorf("получение записи из БД. %w", err)
	}
	return real, nil
}
func (p *PStorage) Ping(ctx context.Context) error {
	return p.Pool.Ping(ctx)
}
func (p *PStorage) Close() error {
	p.Pool.Close()
	return nil
}
