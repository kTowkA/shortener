package app

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/kTowkA/shortener/internal/config"
	"github.com/kTowkA/shortener/internal/model"
	"github.com/kTowkA/shortener/internal/storage"
	mocks "github.com/kTowkA/shortener/internal/storage/mocs"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type serverSuite struct {
	suite.Suite
	server      *Server
	mockStorage *mocks.Storager
}

func (suite *serverSuite) SetupSuite() {
	suite.Suite.T().Log("Suite setup")

	suite.server = &Server{
		logger: slog.Default(),
		Config: config.DefaultConfig,
	}
	suite.mockStorage = new(mocks.Storager)
	suite.server.db = suite.mockStorage

}

// завершение работы нашего приложения
func (suite *serverSuite) TearDownSuite() {
	suite.mockStorage.On("Close", mock.Anything).Return(nil)
	err := suite.mockStorage.Close()
	suite.Require().NoError(err)
	suite.mockStorage.AssertExpectations(suite.T())
}

func (suite *serverSuite) TestMainRoute() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.server.encodeURL(w, r)
	}))
	defer ts.Close()

	req := resty.New().SetHeader("content-type", "text/plain").SetBaseURL(ts.URL).R()
	resp, err := req.SetBody(nil).Post("")
	suite.Require().NoError(err)
	suite.EqualValues(http.StatusBadRequest, resp.StatusCode())

	resp, err = req.SetBody(strings.NewReader("")).Post("")
	suite.Require().NoError(err)
	suite.EqualValues(http.StatusBadRequest, resp.StatusCode())

	resp, err = req.SetBody(strings.NewReader("sadasdasdasdf")).Post("")
	suite.Require().NoError(err)
	suite.EqualValues(http.StatusBadRequest, resp.StatusCode())

	link := generateLink(5, 30)
	suite.mockStorage.On("SaveURL", mock.Anything, mock.Anything, link, mock.AnythingOfType("string")).Return("test", nil)
	resp, err = req.SetBody(strings.NewReader(link)).Post("")
	suite.Require().NoError(err)
	suite.EqualValues(http.StatusCreated, resp.StatusCode())

	link = generateLink(4, 30)
	suite.mockStorage.On("SaveURL", mock.Anything, mock.Anything, link, mock.AnythingOfType("string")).Return("test", storage.ErrURLConflict)
	resp, err = req.SetBody(strings.NewReader(link)).Post("")
	suite.Require().NoError(err)
	suite.EqualValues(http.StatusConflict, resp.StatusCode())
}

func (suite *serverSuite) TestDecodeURL() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.server.decodeURL(w, r)
	}))
	defer ts.Close()

	req := resty.New().SetBaseURL(ts.URL).R()

	resp, err := req.SetBody(nil).Get("")
	suite.Require().NoError(err)
	suite.EqualValues(http.StatusBadRequest, resp.StatusCode())

	suite.mockStorage.On("RealURL", mock.Anything, "ErrURLNotFound").Return(model.StorageJSON{}, storage.ErrURLNotFound)
	resp, err = req.SetBody(nil).Get("/ErrURLNotFound")
	suite.Require().NoError(err)
	suite.EqualValues(http.StatusNotFound, resp.StatusCode())

	suite.mockStorage.On("RealURL", mock.Anything, "StatusInternalServerError").Return(model.StorageJSON{}, errors.New("StatusInternalServerError"))
	resp, err = req.SetBody(nil).Get("/StatusInternalServerError")
	suite.Require().NoError(err)
	suite.EqualValues(http.StatusInternalServerError, resp.StatusCode())

	suite.mockStorage.On("RealURL", mock.Anything, "IsDeleted").Return(model.StorageJSON{IsDeleted: true}, nil)
	resp, err = req.SetBody(nil).Get("/IsDeleted")
	suite.Require().NoError(err)
	suite.EqualValues(http.StatusGone, resp.StatusCode())

	suite.mockStorage.On("RealURL", mock.Anything, "StatusTemporaryRedirect").Return(model.StorageJSON{}, nil)
	resp, err = req.SetBody(nil).Get("/StatusTemporaryRedirect")
	suite.Require().NoError(err)
	suite.EqualValues(http.StatusTemporaryRedirect, resp.StatusCode())
}

func (suite *serverSuite) TestAPIShorten() {
	// ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	suite.server.apiShorten(w, r)
	// }))
	// defer ts.Close()

	// req := resty.New().SetBaseURL(ts.URL).SetHeader("content-type", "application/json").R()

	// resp, err := req.SetBody(nil).Post("")
	// suite.NoError(err)
	// suite.EqualValues(http.StatusBadRequest, resp.StatusCode())

	// resp, err = req.SetBody(strings.NewReader(`{"url":http://ddd.com"}`)).Post("")
	// suite.NoError(err)
	// suite.EqualValues(http.StatusBadRequest, resp.StatusCode())

	// resp, err = req.SetBody(model.RequestShortURL{URL: "http://ddd"}).Post("")
	// suite.NoError(err)
	// suite.EqualValues(http.StatusBadRequest, resp.StatusCode())

	// suite.mockStorage.On("SaveURL", mock.Anything, mock.Anything, "http://ddd.com", mock.AnythingOfType("string")).Return("http://ddd.com", nil)
	// body = model.RequestShortURL{
	// 	URL: "http://ddd.com",
	// }
	// resp, err = req.SetBody(body).Post("")
	// suite.NoError(err)
	// suite.EqualValues(http.StatusCreated, resp.StatusCode())
}

func (suite *serverSuite) TestPing() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.server.ping(w, r)
	}))
	defer ts.Close()

	req := resty.New().SetBaseURL(ts.URL).R()

	suite.mockStorage.On("Ping", mock.Anything).Return(nil).Once()
	resp, err := req.Get("")
	suite.Require().NoError(err)
	suite.EqualValues(http.StatusOK, resp.StatusCode())

	suite.mockStorage.On("Ping", mock.Anything).Return(errors.New(""))
	resp, err = req.Get("")
	suite.Require().NoError(err)
	suite.EqualValues(http.StatusInternalServerError, resp.StatusCode())
}

func (suite *serverSuite) TestBatch() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.server.batch(w, r)
	}))
	defer ts.Close()

	req := resty.New().SetBaseURL(ts.URL).R()

	resp, err := req.SetBody(`[{"1":"2"},{"3":"4"}]`).Post("")
	suite.Require().NoError(err)
	suite.EqualValues(http.StatusBadRequest, resp.StatusCode())

}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(serverSuite))
}

func BenchmarkEncode(b *testing.B) {
	s := new(serverSuite)
	s.SetT(&testing.T{})
	s.SetupSuite()
	b.StartTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.TestMainRoute()
	}
	b.StopTimer()
	s.TearDownSuite()
}
func BenchmarkDecode(b *testing.B) {
	s := new(serverSuite)
	s.SetT(&testing.T{})
	s.SetupSuite()
	b.StartTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.TestDecodeURL()
	}
	b.StopTimer()
	s.TearDownSuite()
}
func BenchmarkPing(b *testing.B) {
	s := new(serverSuite)
	s.SetT(&testing.T{})
	s.SetupSuite()
	b.StartTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.TestPing()
	}
	b.StopTimer()
	s.TearDownSuite()
}
