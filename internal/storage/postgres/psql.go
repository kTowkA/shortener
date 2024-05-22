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
			user_id uuid,
			short_url text,
			original_url text,
			is_deleted bool,
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
func (p *PStorage) SaveURL(ctx context.Context, userID uuid.UUID, real, short string) (string, error) {
	resp, err := p.Batch(
		ctx,
		userID,
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
func (p *PStorage) RealURL(ctx context.Context, short string) (model.StorageJSON, error) {
	answ := model.StorageJSON{}
	err := p.QueryRow(
		ctx,
		"SELECT original_url,is_deleted FROM url_list WHERE short_url=$1",
		short,
	).Scan(
		&answ.OriginalURL,
		&answ.IsDeleted,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.StorageJSON{}, storage.ErrURLNotFound
	}
	if err != nil {
		return model.StorageJSON{}, fmt.Errorf("получение записи из БД. %w", err)
	}
	return answ, nil
}
func (p *PStorage) Ping(ctx context.Context) error {
	return p.Pool.Ping(ctx)
}
func (p *PStorage) Close() error {
	p.Pool.Close()
	return nil
}

func (p *PStorage) Batch(ctx context.Context, userID uuid.UUID, values model.BatchRequest) (model.BatchResponse, error) {
	tx, err := p.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("создание транзакции. %w", err)
	}

	// проходим по нашим значениям и создаем batch
	b := pgx.Batch{}
	for _, v := range values {
		b.Queue(
			"INSERT INTO url_list(uuid,user_id,original_url,short_url,is_deleted) VALUES($1,$2,$3,$4,$5)",
			uuid.New(),
			userID,
			v.OriginalURL,
			v.ShortURL,
			false,
		)
	}
	// отправляем весь batch
	br := tx.SendBatch(ctx, &b)

	// результирующая переменная все ли было хорошо
	ok := true
	// заполняем результат,для этого проходим по переданным значениям и вызываем Exec у BatchResult
	result := make([]model.BatchResponseElement, 0, len(values))
	for _, v := range values {
		// e содержит наш результат
		e := model.BatchResponseElement{
			CorrelationID: v.CorrelationID,
			OriginalURL:   v.OriginalURL,
		}

		// смотрим как выполнилось
		// возможные варианты:
		// 1. была ошибка
		// 2. ошибки не было, но строку не вставили (маловероятное, но вдруг и так быват)
		// 3. все прошло успешно
		tc, err := br.Exec()

		switch {
		case err != nil:
			ok = false
			e.Error = err
			var pgErr *pgconn.PgError
			// если это не внутренняя ошибка postgres, то выходим
			if !errors.As(err, &pgErr) {
				break
			}
			// если это не ошибка UniqueViolation то нам неинтересно
			if pgErr.Code != pgerrcode.UniqueViolation {
				break
			}
			if strings.Contains(err.Error(), "original_url") { //оригинальный урл. надо получить уже имеющиеся совпадение
				// ищем ранее сохраненный url
				short, err := p.short(ctx, v.OriginalURL, userID)
				if err != nil {
					e.Error = err
					break
				}
				e.ShortURL = short
				e.Error = storage.ErrURLConflict
			} else { // а в этом случае коллизия, т.е. у нас уже есть такой же ключ в колонке short_url
				e.Collision = true
				e.Error = storage.ErrURLIsExist
			}
		case tc.RowsAffected() != 1:
			ok = false
			e.Error = fmt.Errorf("не было сохранено")
		default:
			e.ShortURL = v.ShortURL
		}

		result = append(result, e)
	}
	// не забываем закрыть
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
func (p *PStorage) DeleteURLs(ctx context.Context, deleteLinks []model.DeleteURLMessage) error {
	tx, err := p.Begin(ctx)
	if err != nil {
		return fmt.Errorf("создание транзакции. %w", err)
	}

	// проходим по нашим значениям и создаем batch
	b := pgx.Batch{}
	for _, v := range deleteLinks {
		b.Queue(
			"UPDATE url_list SET is_deleted=$1 WHERE user_id=$2 AND short_url=$3",
			true,
			v.UserID,
			v.ShortURL,
		)
	}
	// отправляем весь batch
	br := tx.SendBatch(ctx, &b)

	grpErrors := make([]error, 0, len(deleteLinks))

	for _, v := range deleteLinks {
		tc, err := br.Exec()
		switch {
		case err == pgx.ErrNoRows || (err == nil && tc.RowsAffected() == 0):
			grpErrors = append(grpErrors, fmt.Errorf("userID %s, shortURL %s. %w", v.UserID, v.ShortURL, storage.ErrURLNotFound))
		case err != nil:
			grpErrors = append(grpErrors, fmt.Errorf("userID %s, shortURL %s. %w", v.UserID, v.ShortURL, err))
		}
	}
	// не забываем закрыть
	err = br.Close()
	if err != nil {
		err = tx.Rollback(ctx)
		if err != nil {
			return fmt.Errorf("откат изменений транзакции. %w", err)
		}
	} else {
		err = tx.Commit(ctx)
		if err != nil {
			return fmt.Errorf("сохранение изменений транзакции. %w", err)
		}
	}
	return errors.Join(grpErrors...)
}

// получить уже сохраненое значение (для исключения дублирования original_url)
func (p *PStorage) short(ctx context.Context, real string, userID uuid.UUID) (string, error) {

	var short string
	err := p.QueryRow(
		ctx,
		"SELECT short_url FROM url_list WHERE original_url=$1 AND user_id=$2",
		real,
		userID,
	).Scan(&short)
	if err != nil && err != pgx.ErrNoRows {
		return "", err
	}
	if err == pgx.ErrNoRows {
		return "", storage.ErrURLNotFound
	}
	return short, nil
}

func (p *PStorage) UserURLs(ctx context.Context, userID uuid.UUID) ([]model.StorageJSON, error) {
	rows, err := p.Query(
		ctx,
		"SELECT short_url,original_url FROM url_list WHERE user_id=$1",
		userID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, storage.ErrURLNotFound
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]model.StorageJSON, 0)
	for rows.Next() {
		r := model.StorageJSON{}
		err = rows.Scan(&r.ShortURL, &r.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("получение отдельной записи для пользователя. %w", err)
		}
		results = append(results, r)
	}
	if len(results) == 0 {
		return nil, storage.ErrURLNotFound
	}
	return results, nil
}
