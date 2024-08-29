package utils

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/kTowkA/shortener/internal/model"
	"github.com/kTowkA/shortener/internal/storage"
)

// SaveLink пробует сгенерировать новую сокращенную ссылку для link за attems попыток для пользователя userID и сохранить в store.
func SaveLink(ctx context.Context, store storage.Storager, userID uuid.UUID, link string) (string, error) {
	shortLink, err := GenerateShortStringSHA1(link, defaultLenght)
	if err != nil {
		return "", err
	}

	// создаем короткую ссылка за attems попыток генерации
	for i := 0; i < attempts; i++ {
		savedLink, err := store.SaveURL(ctx, userID, link, shortLink)
		// такая ссылка уже существует
		if errors.Is(err, storage.ErrURLIsExist) {
			shortLink = shortLink + bonus[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(bonus))]
			continue
		} else if errors.Is(err, storage.ErrURLConflict) {
			return savedLink, err
		} else if err != nil {
			return "", err
		}

		// успешно
		return savedLink, nil
	}

	// не уложись в заданное количество попыток для создания короткой ссылки
	return "", fmt.Errorf("не смогли создать короткую ссылку за %d попыток генерации", attempts)
}

// SaveBatch пробует сохранить ссылки из batch в хранилище store
func SaveBatch(ctx context.Context, store storage.Storager, userID uuid.UUID, batch model.BatchRequest) (model.BatchResponse, error) {
	result := make([]model.BatchResponseElement, 0, len(batch))
	for i := 0; i < attempts; i++ {
		err := GenerateShortStringsSHA1ForBatch(batch)
		if err != nil {
			return nil, err
		}
		resp, err := store.Batch(ctx, userID, batch)
		if err != nil {
			return nil, err
		}
		// очищаем batch и будем складывать в него строки с коллизиями
		batch = make(model.BatchRequest, 0)
		for i := range resp {
			if resp[i].ShortURL != "" {
				// resp[i].ShortURL = s.Config.BaseAddress() + resp[i].ShortURL
				result = append(result, resp[i])
				continue
			}
			// если были колизии пробуем сохранить с добавлением подстроки
			if resp[i].Collision {
				batch = append(batch, model.BatchRequestElement{
					CorrelationID: resp[i].CorrelationID,
					OriginalURL:   resp[i].OriginalURL,
					ShortURL:      resp[i].ShortURL,
				})
			}
		}
		if len(batch) == 0 {
			break
		}
	}
	// не смогли уложиться в заданное количество попыток для сохранения всех ссылок
	if len(batch) != 0 {
		return nil, fmt.Errorf("не было сохранено %d записей", len(batch))
	}
	return result, nil
}
