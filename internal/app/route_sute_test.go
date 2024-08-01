// Suite Test версия 2
//
//	после выполнения курсового проекта было желание улучшить ттекущий проект и в первую очередь это касалось тестов
//
// результаты приведены в этом файле
//
// в данном тесте используется мок и запускается не наше приложение, а httptest.server с роутом на основе нашего приложения
//
// используются почти везде табличные тесты.
//
// покрытие > 70 % на текущий момент
package app

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/kTowkA/shortener/internal/config"
	"github.com/kTowkA/shortener/internal/model"
	"github.com/kTowkA/shortener/internal/storage"
	mocks "github.com/kTowkA/shortener/internal/storage/mocs"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	waitCtxTest = time.Duration(10 * time.Second)
)

type AppSuite struct {
	suite.Suite
	ts          *httptest.Server
	mockStorage *mocks.Storager
}

type Test struct {
	name        string
	call        func() (*resty.Response, error)
	callStorage func() *mock.Call
	wantStatus  int
	wantBody    any
}

func (suite *AppSuite) SetupSuite() {
	// запускаем начальную настройку
	suite.Suite.T().Log("Suite setup")

	suite.mockStorage = new(mocks.Storager)
	// создаем новый экземпляр нашего приложения
	srv, err := NewServer(config.DefaultConfig, slog.Default())
	suite.Require().NoError(err, "create app")

	srv.db = suite.mockStorage

	// устанавливаем роут в сервере приложения
	srv.setRoute()
	// и используем этот роут из приложения для тестового сервера
	suite.ts = httptest.NewServer(srv.server.Handler)
}
func (suite *AppSuite) TearDownSuite() {
	//заканчиваем работу с тестовым сценарием
	defer suite.Suite.T().Log("Suite test is ended")

	// закрываем мок хранилища
	suite.mockStorage.On("Close").Return(nil).Once()
	err := suite.mockStorage.Close()
	suite.Require().NoError(err, "close mock storage")

	// закрываем тестовый сервер
	suite.ts.Close()

	// смотрим чтобы все описанные вызовы были использованы
	suite.mockStorage.AssertExpectations(suite.T())
}

func (suite *AppSuite) TestPing() {
	ctx, cancel := context.WithTimeout(context.Background(), waitCtxTest)
	defer cancel()

	const path = "/ping"
	tests := []Test{
		{
			name: "все хорошо",
			callStorage: func() *mock.Call {
				return suite.mockStorage.On("Ping", mock.Anything).Return(nil).Once()
			},
			call: func() (*resty.Response, error) {
				return resty.New().R().SetContext(ctx).Get(suite.ts.URL + path)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "пинг не проходит",
			callStorage: func() *mock.Call {
				return suite.mockStorage.On("Ping", mock.Anything).Return(errors.New("ping error")).Once()
			},
			call: func() (*resty.Response, error) {
				return resty.New().R().SetContext(ctx).Get(suite.ts.URL + path)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, t := range tests {
		if t.callStorage != nil {
			t.callStorage()
		}
		resp, err := t.call()
		suite.Require().NoError(err, t.name)
		suite.EqualValues(t.wantStatus, resp.StatusCode(), t.name)
		if t.wantBody != nil {
			result := model.ResponseShortURL{}
			err = json.Unmarshal(resp.Body(), &result)
			suite.Require().NoError(err, t.name)
			suite.EqualValues(t.wantBody, result, t.name)
		}
	}
}
func (suite *AppSuite) TestEncode() {
	ctx, cancel := context.WithTimeout(context.Background(), waitCtxTest)
	defer cancel()

	const path = "/"

	link := "https://go.dev"

	tests := []Test{
		{
			name: "неправильный метод",
			call: func() (*resty.Response, error) {
				return resty.New().R().SetContext(ctx).SetHeader("Content-type", "text/plain").Get(suite.ts.URL + path)
			},
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name: "неправильный контент-тайп",
			call: func() (*resty.Response, error) {
				return resty.New().R().SetContext(ctx).SetHeader("Content-type", "application/json").Post(suite.ts.URL + path)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "пустой запрос",
			call: func() (*resty.Response, error) {
				return resty.New().R().SetContext(ctx).SetHeader("Content-type", "text/plain").SetBody("").Post(suite.ts.URL + path)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "невалидная ссылка",
			call: func() (*resty.Response, error) {
				return resty.New().R().SetContext(ctx).SetHeader("Content-type", "text/plain").SetBody("ht//go.dev").Post(suite.ts.URL + path)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "валидная ссылка / все создается",
			call: func() (*resty.Response, error) {
				return resty.New().R().SetContext(ctx).SetHeader("Content-type", "text/plain").SetBody(link).Post(suite.ts.URL + path)
			},
			callStorage: func() *mock.Call {
				return suite.mockStorage.On("SaveURL", mock.Anything, mock.Anything, link, mock.Anything).Return("", nil).Once()
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "валидная ссылка / конфликт",
			call: func() (*resty.Response, error) {
				return resty.New().R().SetContext(ctx).SetHeader("Content-type", "text/plain").SetBody(link).Post(suite.ts.URL + path)
			},
			callStorage: func() *mock.Call {
				return suite.mockStorage.On("SaveURL", mock.Anything, mock.Anything, link, mock.Anything).Return("", storage.ErrURLConflict).Once()
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "валидная ссылка / внутренняя ошибка",
			call: func() (*resty.Response, error) {
				return resty.New().R().SetContext(ctx).SetHeader("Content-type", "text/plain").SetBody(link).Post(suite.ts.URL + path)
			},
			callStorage: func() *mock.Call {
				return suite.mockStorage.On("SaveURL", mock.Anything, mock.Anything, link, mock.Anything).Return("", errors.New("save error")).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "валидная ссылка / не уложились в заданое количество попыток",
			call: func() (*resty.Response, error) {
				return resty.New().R().SetContext(ctx).SetHeader("Content-type", "text/plain").SetBody(link).Post(suite.ts.URL + path)
			},
			callStorage: func() *mock.Call {
				return suite.mockStorage.On("SaveURL", mock.Anything, mock.Anything, link, mock.Anything).Return("", storage.ErrURLIsExist)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, t := range tests {
		if t.callStorage != nil {
			t.callStorage()
		}
		resp, err := t.call()
		suite.Require().NoError(err, t.name)
		suite.EqualValues(t.wantStatus, resp.StatusCode(), t.name)
		if t.wantBody != nil {
			result := model.ResponseShortURL{}
			err = json.Unmarshal(resp.Body(), &result)
			suite.Require().NoError(err, t.name)
			suite.EqualValues(t.wantBody, result, t.name)
		}
	}
}
func (suite *AppSuite) TestDecode() {
	ctx, cancel := context.WithTimeout(context.Background(), waitCtxTest)
	defer cancel()

	const path = "/"

	short := "shorterurl"
	originalURL := "https://practicum.yandex.ru"

	tests := []Test{
		{
			name: "ничего не нашли",
			call: func() (*resty.Response, error) {
				return resty.New().R().SetContext(ctx).Get(suite.ts.URL + path + short)
			},
			callStorage: func() *mock.Call {
				return suite.mockStorage.On("RealURL", mock.Anything, short).Return(model.StorageJSON{}, storage.ErrURLNotFound).Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "была ошибка",
			callStorage: func() *mock.Call {
				return suite.mockStorage.On("RealURL", mock.Anything, short).Return(model.StorageJSON{}, errors.New("real error")).Once()
			},
			call: func() (*resty.Response, error) {
				return resty.New().R().SetContext(ctx).Get(suite.ts.URL + path + short)
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "был удален",
			call: func() (*resty.Response, error) {
				return resty.New().R().SetContext(ctx).Get(suite.ts.URL + path + short)
			},
			callStorage: func() *mock.Call {
				return suite.mockStorage.On("RealURL", mock.Anything, short).Return(model.StorageJSON{IsDeleted: true}, nil).Once()
			},
			wantStatus: http.StatusGone,
		},
	}
	for _, t := range tests {
		if t.callStorage != nil {
			t.callStorage()
		}
		resp, err := t.call()
		suite.Require().NoError(err, t.name)
		suite.EqualValues(t.wantStatus, resp.StatusCode(), t.name)
		if t.wantBody != nil {
			result := model.ResponseShortURL{}
			err = json.Unmarshal(resp.Body(), &result)
			suite.Require().NoError(err, t.name)
			suite.EqualValues(t.wantBody, result, t.name)
		}
	}

	// ок. тут выбивается из табличного теста, решил его не усложнять и сделать это отдельно
	suite.mockStorage.On("RealURL", mock.Anything, short).Return(model.StorageJSON{OriginalURL: originalURL}, nil).Once()
	resp, err := resty.New().SetRedirectPolicy(resty.NoRedirectPolicy()).R().SetContext(ctx).Get(suite.ts.URL + path + short)
	suite.ErrorIs(err, resty.ErrAutoRedirectDisabled)
	suite.EqualValues(http.StatusTemporaryRedirect, resp.StatusCode())
	suite.EqualValues(originalURL, resp.Header().Get("Location"))
}
func (suite *AppSuite) TestUserURLs() {
	ctx, cancel := context.WithTimeout(context.Background(), waitCtxTest)
	defer cancel()

	const path = "/api/user/urls"

	// для чистоты теста создаем клиента и ищем ID пользователя
	cl := resty.New()
	// делаем запрос
	short := "short"
	suite.mockStorage.On("RealURL", mock.Anything, short).Return(model.StorageJSON{}, storage.ErrURLNotFound).Once()
	resp, err := cl.R().SetContext(ctx).Get(suite.ts.URL + "/" + short)
	suite.Require().NoError(err)
	suite.EqualValues(http.StatusNotFound, resp.StatusCode())
	jwtC := ""
	for _, c := range resp.Cookies() {
		if c.Name == authCookie {
			jwtC = c.Value
			break
		}
	}
	userID, err := getUserIDFromToken(jwtC, config.DefaultConfig.SecretKey())
	suite.Require().NoError(err)

	// want когда есть ссылки, их ожидаем
	want := []model.StorageJSON{
		{
			ShortURL: "short_1",
		},
		{
			ShortURL: "short_2",
		},
	}
	tests := []Test{
		{
			name: "не авторизован (новый клиент)",
			call: func() (*resty.Response, error) {
				return resty.New().R().SetContext(ctx).Get(suite.ts.URL + path)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "нет ссылок",
			call: func() (*resty.Response, error) {
				return cl.R().SetContext(ctx).Get(suite.ts.URL + path)
			},
			callStorage: func() *mock.Call {
				return suite.mockStorage.On("UserURLs", mock.Anything, userID).Return(nil, storage.ErrURLNotFound).Once()
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "есть ссылки",
			call: func() (*resty.Response, error) {
				return cl.R().SetContext(ctx).Get(suite.ts.URL + path)
			},
			callStorage: func() *mock.Call {
				return suite.mockStorage.On("UserURLs", mock.Anything, userID).Return(want, nil).Once()
			},
			wantStatus: http.StatusOK,
			wantBody:   want,
		},
	}

	for _, t := range tests {
		if t.callStorage != nil {
			t.callStorage()
		}
		resp, err := t.call()
		suite.Require().NoError(err, t.name)
		suite.EqualValues(t.wantStatus, resp.StatusCode(), t.name)
		if t.wantBody != nil {
			result := []model.StorageJSON{}
			err = json.Unmarshal(resp.Body(), &result)
			suite.Require().NoError(err, t.name)
			suite.EqualValues(t.wantBody, result, t.name)
		}
	}
}
func (suite *AppSuite) TestUserURLsDelete() {
	ctx, cancel := context.WithTimeout(context.Background(), waitCtxTest)
	defer cancel()

	const path = "/api/user/urls"

	// создаем клиента и делаем запрос чтобы пользователь уже был сохранен
	cl := resty.New()
	short := "short"
	suite.mockStorage.On("RealURL", mock.Anything, short).Return(model.StorageJSON{}, storage.ErrURLNotFound).Once()
	resp, err := cl.R().SetContext(ctx).Get(suite.ts.URL + "/" + short)
	suite.Require().NoError(err)
	suite.EqualValues(http.StatusNotFound, resp.StatusCode())

	tests := []Test{
		{
			name: "не авторизован",
			call: func() (*resty.Response, error) {
				return resty.New().R().SetContext(ctx).Delete(suite.ts.URL + path)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "плохой контент-тайп",
			call: func() (*resty.Response, error) {
				return cl.R().SetContext(ctx).SetHeader("Content-type", "text/plain").Delete(suite.ts.URL + path)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "плохое тело запроса",
			call: func() (*resty.Response, error) {
				return cl.R().SetContext(ctx).SetBody([]any{1, "2", "3"}).Delete(suite.ts.URL + path)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "принято на удаление",
			call: func() (*resty.Response, error) {
				return cl.R().SetContext(ctx).SetBody([]string{"1", "2", "3"}).Delete(suite.ts.URL + path)
			},
			wantStatus: http.StatusAccepted,
		},
	}

	for _, t := range tests {
		if t.callStorage != nil {
			t.callStorage()
		}
		resp, err := t.call()
		suite.Require().NoError(err, t.name)
		suite.EqualValues(t.wantStatus, resp.StatusCode(), t.name)
	}
}
func (suite *AppSuite) TestAPIShorten() {
	const path = "/api/shorten"

	link := "http://one.com"
	req := model.RequestShortURL{
		URL: link,
	}
	saved := "one"

	ctx, cancel := context.WithTimeout(context.Background(), waitCtxTest)
	defer cancel()
	// создаем клиента
	cl := resty.New()

	tests := []struct {
		name        string
		call        func() (*resty.Response, error)
		callStorage func() *mock.Call
		wantStatus  int
		wantBody    any
	}{
		{
			name: "плохой контент-тайп",
			call: func() (*resty.Response, error) {
				return cl.R().SetContext(ctx).SetHeader("Content-type", "text/plain").Post(suite.ts.URL + path)
			},
			callStorage: nil,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "плохое тело запроса",
			call: func() (*resty.Response, error) {
				return cl.R().SetContext(ctx).SetBody(nil).SetHeader("Content-type", "application/json").Post(suite.ts.URL + path)
			},
			callStorage: nil,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "плохое тело запроса. невалидный json тег",
			call: func() (*resty.Response, error) {
				return cl.R().SetContext(ctx).SetBody(`{"not_valid":"http://one.com"}`).SetHeader("Content-type", "application/json").Post(suite.ts.URL + path)
			},
			callStorage: nil,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "плохое тело запроса. невалидная ссылка",
			call: func() (*resty.Response, error) {
				return cl.R().SetContext(ctx).SetBody(`{"url":"one.com"}`).SetHeader("Content-type", "application/json").Post(suite.ts.URL + path)
			},
			callStorage: nil,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "конфликт",
			call: func() (*resty.Response, error) {
				return cl.R().SetContext(ctx).SetBody(req).SetHeader("Content-type", "application/json").Post(suite.ts.URL + path)
			},
			callStorage: func() *mock.Call {
				return suite.mockStorage.On("SaveURL", mock.Anything, mock.Anything, link, mock.Anything).Return(saved, storage.ErrURLConflict).Once()
			},
			wantStatus: http.StatusConflict,
			wantBody: model.ResponseShortURL{
				Result: config.DefaultConfig.BaseAddress() + saved,
			},
		},
		{
			name: "внутренняя ошибка",
			call: func() (*resty.Response, error) {
				return cl.R().SetContext(ctx).SetBody(req).SetHeader("Content-type", "application/json").Post(suite.ts.URL + path)
			},
			callStorage: func() *mock.Call {
				return suite.mockStorage.On("SaveURL", mock.Anything, mock.Anything, link, mock.Anything).Return("", errors.New("shorten error")).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "все хорошо",
			call: func() (*resty.Response, error) {
				return cl.R().SetContext(ctx).SetBody(req).SetHeader("Content-type", "application/json").Post(suite.ts.URL + path)
			},
			callStorage: func() *mock.Call {
				return suite.mockStorage.On("SaveURL", mock.Anything, mock.Anything, link, mock.Anything).Return(saved, nil).Once()
			},
			wantStatus: http.StatusCreated,
			wantBody: model.ResponseShortURL{
				Result: config.DefaultConfig.BaseAddress() + saved,
			},
		},
		{
			name: "не уложились в заданное количество попыток",
			call: func() (*resty.Response, error) {
				return cl.R().SetContext(ctx).SetBody(req).SetHeader("Content-type", "application/json").Post(suite.ts.URL + path)
			},
			callStorage: func() *mock.Call {
				return suite.mockStorage.On("SaveURL", mock.Anything, mock.Anything, link, mock.Anything).Return("", storage.ErrURLIsExist)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, t := range tests {
		if t.callStorage != nil {
			t.callStorage()
		}
		resp, err := t.call()
		suite.Require().NoError(err, t.name)
		suite.EqualValues(t.wantStatus, resp.StatusCode(), t.name)
		if t.wantBody != nil {
			result := model.ResponseShortURL{}
			err = json.Unmarshal(resp.Body(), &result)
			suite.Require().NoError(err, t.name)
			suite.EqualValues(t.wantBody, result, t.name)
		}
	}
}
func (suite *AppSuite) TestAPIBatch() {
	ctx, cancel := context.WithTimeout(context.Background(), waitCtxTest)
	defer cancel()

	const path = "/api/shorten/batch"

	// создаем клиента
	cl := resty.New()

	// want для теста ок
	want := model.BatchResponse{
		{
			ShortURL: "one",
		},
		{
			ShortURL: "two",
		},
	}
	tests := []Test{
		{
			name: "плохой контент-тайп",
			call: func() (*resty.Response, error) {
				return cl.R().SetContext(ctx).SetHeader("Content-type", "text/plain").Post(suite.ts.URL + path)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "плохое тело запроса",
			call: func() (*resty.Response, error) {
				return cl.R().SetContext(ctx).SetBody(nil).SetHeader("Content-type", "application/json").Post(suite.ts.URL + path)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "невалидный json",
			call: func() (*resty.Response, error) {
				return cl.R().SetContext(ctx).SetBody(`[{"original_url_2":"http://one.com"}]`).SetHeader("Content-type", "application/json").Post(suite.ts.URL + path)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "нет запросов",
			call: func() (*resty.Response, error) {
				return cl.R().SetContext(ctx).SetBody(model.BatchRequest{}).SetHeader("Content-type", "application/json").Post(suite.ts.URL + path)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "ок",
			call: func() (*resty.Response, error) {
				return cl.R().SetContext(ctx).SetBody(model.BatchRequest{{OriginalURL: "http://one.com"}, {OriginalURL: "http://two.com"}}).SetHeader("Content-type", "application/json").Post(suite.ts.URL + path)
			},
			callStorage: func() *mock.Call {
				return suite.mockStorage.On("Batch", mock.Anything, mock.Anything, mock.AnythingOfType("model.BatchRequest")).Return(want, nil).Once()
			},
			wantStatus: http.StatusCreated,
			wantBody:   want,
		},
	}
	for _, t := range tests {
		if t.callStorage != nil {
			t.callStorage()
		}
		resp, err := t.call()
		suite.Require().NoError(err, t.name)
		suite.EqualValues(t.wantStatus, resp.StatusCode(), t.name)
		if t.wantBody != nil {
			result := model.BatchResponse{}
			err = json.Unmarshal(resp.Body(), &result)
			suite.Require().NoError(err, t.name)
			suite.EqualValues(t.wantBody, result, t.name)
		}
	}
}
func TestAppSuite(t *testing.T) {
	suite.Run(t, new(AppSuite))
}
