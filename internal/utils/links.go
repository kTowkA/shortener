package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"net/url"
	"strings"
	"time"

	"math/rand"

	"github.com/kTowkA/shortener/internal/model"
)

var (
	bonus = []string{"q", "w", "e", "r", "t", "y", "u", "i", "o", "p", "Q", "W", "E", "R", "T", "Y", "U", "I", "O", "P"}
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

// GenerateShortStringsSHA1ForBatch генерирует новую строку для каждого элемента batch
func GenerateShortStringsSHA1ForBatch(batch model.BatchRequest) error {
	for i := range batch {
		if batch[i].ShortURL != "" {
			batch[i].ShortURL += bonus[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(bonus))]
			continue
		}
		short, err := GenerateShortStringSHA1(batch[i].OriginalURL, defaultLenght)
		if err != nil {
			return err
		}
		batch[i].ShortURL = short
	}
	return nil
}

// ValidateAndGenerateBatch проверяет переданный batch, удаляя пустые значения и невалидные ссылки. Возвращает model.BatchRequest только с валидными ссылками
func ValidateAndGenerateBatch(batch model.BatchRequest) model.BatchRequest {
	newBatch := make([]model.BatchRequestElement, 0, len(batch))
	for _, v := range batch {
		v.OriginalURL = strings.TrimSpace(v.OriginalURL)
		if v.OriginalURL == "" {
			continue
		}
		if _, err := url.ParseRequestURI(v.OriginalURL); err != nil {
			continue
		}
		newBatch = append(newBatch, v)
	}
	return newBatch
}
