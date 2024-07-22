package app

import (
	"math/rand"
	"testing"
	"time"

	"github.com/kTowkA/shortener/internal/model"
)

func BenchmarkGenerateSHA1(b *testing.B) {
	b.Run("lenght 5", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			str := randomString(5, 30)
			b.ResetTimer()
			_, _ = generateSHA1(str, 5)
		}
	})
	b.Run("lenght 10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			str := randomString(5, 30)
			b.ResetTimer()
			_, _ = generateSHA1(str, 10)
		}
	})
	b.Run("lenght 15", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			str := randomString(5, 30)
			b.ResetTimer()
			_, _ = generateSHA1(str, 10)
		}
	})
}

func randomString(minLen, maxLen int) string {
	var letters = "0123456789abcdefghijklmnopqrstuvwxyz"
	r := rand.New(rand.NewSource(time.Now().Unix()))
	slen := r.Intn(maxLen-minLen) + minLen

	s := make([]byte, 0, slen)
	i := 0
	for len(s) < slen {
		idx := r.Intn(len(letters) - 1)
		char := letters[idx]
		if i == 0 && '0' <= char && char <= '9' {
			continue
		}
		s = append(s, char)
		i++
	}

	return string(s)
}
func generateLink(minLen, maxLen int) string {
	return "http://" + randomString(minLen, maxLen) + ".com"
}

func BenchmarkGenerateLinksBatch(b *testing.B) {
	b.Run("batch 100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			batch := generateBatch(100)
			b.ResetTimer()
			_ = generateLinksBatch(batch)
		}
	})
	b.Run("batch 1000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			batch := generateBatch(1000)
			b.ResetTimer()
			_ = generateLinksBatch(batch)
		}
	})
}

func generateBatch(size int) model.BatchRequest {
	b := make([]model.BatchRequestElement, 0, size)
	for i := 0; i < size; i++ {
		b = append(b, model.BatchRequestElement{OriginalURL: randomString(5, 30)})
	}
	return b
}
