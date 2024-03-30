package app

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// var (
// 	defaultAddress     = "localhost:8080"
// 	defaultBaseAddress = "http://localhost:8080"

// 	cfg = config.Config{
// 		Address:     defaultAddress,
// 		BaseAddress: defaultBaseAddress,
// 	}
// )

func TestPost(t *testing.T) {

	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name    string
		request http.Request
		want    want
	}{
		{
			name:    "неправильный метод",
			request: *httptest.NewRequest(http.MethodGet, "http://localhost", nil),
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name:    "неправильный content-type",
			request: *httptest.NewRequest(http.MethodPost, "http://localhost", nil),
			want: want{
				code:        http.StatusBadRequest,
				contentType: plainTextContentType,
			},
		},
		{
			name: "пустое тело запроса 1",
			request: func() http.Request {
				request := *httptest.NewRequest(http.MethodPost, "http://localhost", nil)
				request.Header.Set(contentType, plainTextContentType)
				return request
			}(),
			want: want{
				code:        http.StatusBadRequest,
				contentType: plainTextContentType,
			},
		},
		{
			name: "пустое тело запроса 2",
			request: func() http.Request {
				request := *httptest.NewRequest(http.MethodPost, "http://localhost", strings.NewReader(""))
				request.Header.Set(contentType, plainTextContentType)
				return request
			}(),
			want: want{
				code:        http.StatusBadRequest,
				contentType: plainTextContentType,
			},
		},
		{
			name: "валидный запрос",
			request: func() http.Request {
				request := *httptest.NewRequest(http.MethodPost, "http://localhost", strings.NewReader("http://ya.ru"))
				request.Header.Set(contentType, plainTextContentType)
				return request
			}(),
			want: want{
				code:        http.StatusCreated,
				contentType: plainTextContentType,
			},
		},
	}

	s, err := NewServer()

	if err != nil {
		t.Fatalf("Создание сервера. %v", err)
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.encodeURL(w, &test.request)
			response := w.Result()
			assert.Equal(t, test.want.code, response.StatusCode)
			link, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			if string(link) != "" {
				t.Logf("переданная ссылка: %s. короткая ссылка: %s", "", string(link))
			}
			err = response.Body.Close()
			require.NoError(t, err)
		})
	}
}

func TestGet(t *testing.T) {
	s, err := NewServer()

	if err != nil {
		t.Fatalf("Создание сервера. %v", err)
	}

	// создаем короткую ссылку
	r := httptest.NewRequest(http.MethodPost, "http://localhost", strings.NewReader("http://ya.ru"))
	r.Header.Set(contentType, plainTextContentType)
	w := httptest.NewRecorder()
	s.encodeURL(w, r)
	response := w.Result()
	require.Equal(t, http.StatusCreated, response.StatusCode)
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	err = response.Body.Close()
	require.NoError(t, err)
	fullLinkShort := string(body)
	sl := strings.Split(fullLinkShort, "/")
	linkShort := sl[len(sl)-1]
	log.Println("full", fullLinkShort, "short", linkShort)
	type want struct {
		code        int
		contentType string
		location    string
	}
	tests := []struct {
		name    string
		request http.Request
		want    want
	}{
		{
			name:    "неправильный метод",
			request: *httptest.NewRequest(http.MethodPost, "http://localhost/", nil),
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "нет части path",
			request: func() http.Request {
				request := *httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
				request.Header.Set(contentType, plainTextContentType)
				return request
			}(),
			want: want{
				code:        http.StatusBadRequest,
				contentType: "",
			},
		},
		{
			name: "404",
			request: func() http.Request {
				request := *httptest.NewRequest(http.MethodGet, "http://localhost/404", nil)
				request.Header.Set(contentType, plainTextContentType)
				return request
			}(),
			want: want{
				code:        http.StatusNotFound,
				contentType: "",
			},
		},
		{
			name: "успех 1",
			request: func() http.Request {
				request := *httptest.NewRequest(http.MethodGet, fullLinkShort, nil)
				request.Header.Set(contentType, plainTextContentType)
				return request
			}(),
			want: want{
				code:        http.StatusTemporaryRedirect,
				contentType: "",
				location:    "http://ya.ru",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.decodeURL(w, &test.request)
			response := w.Result()
			assert.Equal(t, test.want.code, response.StatusCode)
			assert.Equal(t, test.want.location, response.Header.Get("Location"))
			err = response.Body.Close()
			require.NoError(t, err)
			// assert.Equal(t, test.want.contentType, response.Header.Get("content-type"))
		})
	}
}
