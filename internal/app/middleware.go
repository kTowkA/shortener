package app

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/kTowkA/shortener/internal/logger"
	"go.uber.org/zap"
)

const (
	authCookie      = "jwt"
	TokenExp12Hours = 12 * time.Hour
)

type contextKey string

// Claims — структура утверждений, которая включает стандартные утверждения и
// одно пользовательское UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
}

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func withLog(h http.Handler) http.Handler {

	logFn := func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   &responseData{},
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)
		logger.Log.Info(
			"входящий запрос",
			zap.String("uri", r.RequestURI),
			zap.String("http метод", r.Method),
			zap.Duration("длительность запроса", duration),
			zap.Int("статус", lw.responseData.status),
			zap.Int("размер ответа", lw.responseData.size),
		)
	}

	return http.HandlerFunc(logFn)
}

type (
	gzipWriter struct {
		http.ResponseWriter
		gzw *gzip.Writer
	}
	gzipReader struct {
		orig io.ReadCloser
		gzr  *gzip.Reader
	}
)

func (gzw *gzipWriter) Write(p []byte) (int, error) {
	return gzw.gzw.Write(p)
}

func (gzr *gzipReader) Read(p []byte) (n int, err error) {
	return gzr.gzr.Read(p)
}

func (gzr *gzipReader) Close() error {
	if err := gzr.orig.Close(); err != nil {
		return err
	}
	return gzr.gzr.Close()
}

func withGZIP(h http.Handler) http.Handler {
	zfunc := func(w http.ResponseWriter, r *http.Request) {
		newWriter := w

		if gzipValidContenType(r.Header) {
			cw := &gzipWriter{
				ResponseWriter: w,
				gzw:            gzip.NewWriter(w),
			}
			newWriter = cw
			cw.Header().Set("Content-Encoding", "gzip")
			defer cw.gzw.Close()
		}

		if strings.Contains(r.Header.Get("content-encoding"), "gzip") {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			rzip, err := gzip.NewReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			gzr := &gzipReader{
				orig: r.Body,
				gzr:  rzip,
			}
			r.Body = gzr
			defer gzr.Close()
		}

		h.ServeHTTP(newWriter, r)
	}
	return http.HandlerFunc(zfunc)
}

func gzipValidContenType(header http.Header) bool {
	validContentType := []string{
		"text/html",
		"application/json",
	}
	if !strings.Contains(header.Get("accept-encoding"), "gzip") {
		return false
	}
	for _, ct := range validContentType {
		if strings.Contains(header.Get("content-type"), ct) {
			return true
		}
	}
	return false
}

func withToken(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromCookie(r, secretKey)
		if err == nil {
			// все хорошо, токен валиден и есть userID - продолжаем
			h.ServeHTTP(w, r)
			return
		}
		// создаем новый токен
		// в настоящий момент нет системы авторизации/регистрации - мы генерируем новый userID в таких случаях
		userID := uuid.New()
		newTokenString, err := buildJWTString(userID, secretKey)
		if err != nil {
			logger.Log.Error("создание токена", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: authCookie, Value: newTokenString})

		// сохраняем ID пользователя в контекте запроса и передаем дальше
		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey("userID"), userID)))
	})
}

// buildJWTString создаёт токен и возвращает его в виде строки.
func buildJWTString(userID uuid.UUID, secret string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp12Hours)),
		},
		// собственное утверждение
		UserID: userID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

// getUserIDFromToken - получает ID из JWT токена
func getUserIDFromToken(tokenString, secret string) (uuid.UUID, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неожиданный метод подписи: %v", t.Header["alg"])
			}
			return []byte(secret), nil
		})
	if err != nil {
		return uuid.UUID{}, err
	}

	if !token.Valid {
		return uuid.UUID{}, fmt.Errorf("токен не прошел проверку")
	}

	return claims.UserID, nil
}

// getUserIDFromCookie - получает ID пользователя из куки
func getUserIDFromCookie(r *http.Request, secret string) (uuid.UUID, error) {
	token, err := r.Cookie(authCookie)
	if err != nil {
		logger.Log.Error("не смогли получить cookie", zap.String("", err.Error()))
		return uuid.UUID{}, fmt.Errorf("не смогли получить cookie. %w", err)
	}
	userID, err := getUserIDFromToken(token.Value, secret)
	if err != nil {
		logger.Log.Error("получение userID", zap.Error(err))
		return uuid.UUID{}, fmt.Errorf("не смогли получить userID из токена. %w", err)
	}
	if err := uuid.Validate(userID.String()); err != nil {
		return uuid.UUID{}, fmt.Errorf("userID не представляет собой UUID. %w", err)
	}
	return userID, nil
}
