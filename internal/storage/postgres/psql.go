package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kTowkA/shortener/internal/model"
	"github.com/kTowkA/shortener/internal/storage"
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
	err = storage.bootstrap(ctx)
	if err != nil {
		return nil, fmt.Errorf("инициализация БД таблицами. %w", err)
	}
	return storage, nil
}

func (p *PStorage) bootstrap(ctx context.Context) error {
	tx, err := p.Begin(ctx)
	if err != nil {
		return err
	}
	_, err = tx.Exec(
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
	if err != nil {
		return tx.Rollback(ctx)
	}
	return tx.Commit(ctx)
}

// реализация интерфейса Storager
func (p *PStorage) SaveURL(ctx context.Context, real, short string) (string, error) {
	resp, err := p.Batch(
		ctx,
		model.BatchRequest{
			model.BatchRequestElement{
				OriginalURL: real,
				ShortURL:    short,
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("добавление новой записи в БД. %w", err)
	}
	if resp[0].Collision {
		return "", storage.ErrURLIsExist
	}
	if resp[0].Error != nil {
		return resp[0].ShortURL, resp[0].Error
	}
	return resp[0].ShortURL, nil
}
func (p *PStorage) RealURL(ctx context.Context, short string) (string, error) {
	var real string
	err := p.QueryRow(
		ctx,
		"SELECT original_url FROM url_list WHERE short_url=$1",
		short,
	).Scan(&real)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", storage.ErrURLNotFound
	}
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

func (p *PStorage) Batch(ctx context.Context, values model.BatchRequest) (model.BatchResponse, error) {
	tx, err := p.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("создание транзакции. %w", err)
	}

	// проходим по нашим значениям и создаем batch
	b := pgx.Batch{}
	for _, v := range values {
		b.Queue(
			"INSERT INTO url_list(uuid,original_url,short_url) VALUES($1,$2,$3)",
			uuid.New(),
			v.OriginalURL,
			v.ShortURL,
		)
	}
	// отправляем весь batch и не забываем закрыть
	br := tx.SendBatch(ctx, &b)

	ok := false
	// заполняем результат,для этого проходим по переданным значениям и вызываем Exec у BatchResult
	result := make([]model.BatchResponseElement, 0, len(values))
	for _, v := range values {
		// e содержит наш результат
		e := model.BatchResponseElement{
			CorrelationID: v.CorrelationID,
			OriginalURL:   v.OriginalURL,
		}
		tc, err := br.Exec()
		if err != nil {
			// записываем ошибку
			e.Error = err
			// получаем код ошибки. мы может пытаемся записать уникальное значение в одно из нашим полей с индексом UNIQUE(original_url ИЛИ short_url)
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {

				// была попытка записать уникальное значение.и здесь у нас 2 варианта
				if pgErr.Code == pgerrcode.UniqueViolation {
					// мы не считаем это ошибкой и сбрасываем
					e.Error = nil
					if strings.Contains(err.Error(), "original_url") { //оригинальный урл. надо получить уже имеющиеся совпадение
						// ищем ранее сохраненный url
						short, err := p.short(ctx, v.OriginalURL)
						if err != nil {
							e.Error = err
						} else {
							e.ShortURL = short
							e.Error = storage.ErrURLConflict
						}
					} else { // а в этом случае коллизия, т.е. у нас уже есть такой же ключ в колонке short_url
						e.Collision = true
						e.Error = storage.ErrURLIsExist
					}
				}
			}
		} else {
			ok = true
			// сюда конечно не должны попадать, но на всякий случай
			if tc.RowsAffected() != 1 {
				e.Error = fmt.Errorf("не было сохранено")
			} else {
				e.ShortURL = v.ShortURL
			}
		}
		result = append(result, e)
	}
	br.Close()
	if ok {
		err = tx.Commit(ctx)
		if err != nil {
			return nil, fmt.Errorf("сохранение изменений транзакции. %w", err)
		}
	} else {
		err = tx.Rollback(ctx)
		if err != nil {
			return nil, fmt.Errorf("откат изменений транзакции. %w", err)
		}
	}

	return result, nil
}

// получить уже сохраненое значение (для исключения дублирования original_url)
func (p *PStorage) short(ctx context.Context, real string) (string, error) {

	var short string
	err := p.QueryRow(
		ctx,
		"SELECT short_url FROM url_list WHERE original_url=$1",
		real,
	).Scan(&short)
	if err != nil && err != pgx.ErrNoRows {
		return "", err
	}
	if err == pgx.ErrNoRows {
		return "", storage.ErrURLNotFound
	}
	return short, nil
}
