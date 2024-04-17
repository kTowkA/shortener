package app

import (
	"net/http"
	"time"

	"github.com/kTowkA/shortener/internal/logger"
	"github.com/sirupsen/logrus"
)

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

		logger.Log.WithFields(logrus.Fields{
			"uri":        r.RequestURI,
			"http метод": r.Method,
			"длительность запроса": duration,
			"статус":        lw.responseData.status,
			"размер ответа": lw.responseData.size,
		}).Info("входящий запрос")
	}

	return http.HandlerFunc(logFn)
}
