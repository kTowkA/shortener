package utils

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"math/rand"

	"github.com/google/uuid"
	"github.com/kTowkA/shortener/internal/storage"
)

const (
	defaultLenght = 7
	attems        = 10
)

// GenerateShortStringSHA1 генерирует новую строку с применение алгоритма sha1 из original, обрезая ее длину до lenght и переводя случайные символы в верхний регистр
func GenerateShortStringSHA1(original string, lenght int) (string, error) {
	// получаем sha1
	sum := sha1.Sum([]byte(original))
	symbols := hex.EncodeToString(sum[:])
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	result := strings.Builder{}
	result.Grow(20)
	change := 0
	for _, rs := range symbols {
		s := string(rs)
		change = r.Intn(2)
		if change == 1 {
			s = strings.ToUpper(s)
		}
		_, err := result.WriteString(s)
		if err != nil {
			return "", err
		}
		if result.Len() == lenght {
			break
		}
	}
	return result.String(), nil
}

// SaveLink пробует сгенерировать новую сокращеннуб ссылку для link за attems попыток для пользователя userID и сохранить в store.
func SaveLink(ctx context.Context, store storage.Storager, userID uuid.UUID, link string) (string, error) {
	bonus := []string{"q", "w", "e", "r", "t", "y", "u", "i", "o", "p", "Q", "W", "E", "R", "T", "Y", "U", "I", "O", "P"}
	genLink, err := GenerateShortStringSHA1(link, defaultLenght)
	if err != nil {
		return "", err
	}

	// создаем короткую ссылка за attems попыток генерации
	for i := 0; i < attems; i++ {
		savedLink, err := store.SaveURL(ctx, userID, link, genLink)
		// такая ссылка уже существует
		if errors.Is(err, storage.ErrURLIsExist) {
			genLink = genLink + bonus[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(bonus))]
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
	return "", fmt.Errorf("не смогли создать короткую ссылку за %d попыток генерации", attems)
}
