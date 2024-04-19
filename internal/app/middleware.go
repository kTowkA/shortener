package app

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
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

type (
	gzipWriter struct {
		orig http.ResponseWriter
		gzw  *gzip.Writer
	}
	gzipReader struct {
		orig io.ReadCloser
		gzr  *gzip.Reader
	}
)

func (gzw *gzipWriter) Header() http.Header {
	return gzw.orig.Header()
}
func (gzw *gzipWriter) WriteHeader(statusCode int) {
	gzw.orig.WriteHeader(statusCode)
}
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

		if strings.Contains(r.Header.Get(acceptEncoding), "gzip") && (strings.Contains(contentType, textHTMLContentType) || strings.Contains(contentType, applicationJSONContentType)) {
			cw := &gzipWriter{
				orig: w,
				gzw:  gzip.NewWriter(w),
			}
			newWriter = cw
			defer cw.gzw.Close()
		}

		if strings.Contains(r.Header.Get(contentEncoding), "gzip") {
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
