// собственно тесты на удаление
// работают хорошо. 3 потока (3 разных пользователя)
// вначале генерируются строки и понеслась
// как я утверждаю воркер удаления никак не связан с хендлерами и запускается один раз с сервером
// зато нашел грубую ошибку при генерации строки
package app

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/kTowkA/shortener/internal/config"
	"github.com/kTowkA/shortener/internal/storage"
	"github.com/kTowkA/shortener/internal/storage/memory"
	"github.com/kTowkA/shortener/internal/storage/postgres"
	"github.com/stretchr/testify/suite"
)

type (
	appDeleteTestSuite struct {
		suite.Suite
		server *Server
	}
)

var tmpURL = "https://practicum.yandex.ru/learn/go-advanced/courses/"

func (suite *appDeleteTestSuite) SetupSuite() {
	suite.Suite.T().Log("Suite setup")
	cfg := config.DefaultConfig
	var (
		myStorage storage.Storager
		err       error
	)
	if cfg.DatabaseDSN() != "" {
		myStorage, err = postgres.NewStorage(context.Background(), cfg.DatabaseDSN())
	} else {
		myStorage, err = memory.NewStorage(cfg.FileStoragePath())
	}
	suite.Require().NoError(err)

	s, err := NewServer(cfg, slog.Default())
	suite.Require().NoError(err)
	s.db = myStorage

	suite.server = s
	go s.Run(context.Background())
	time.Sleep(3 * time.Second)
}

func (suite *appDeleteTestSuite) TearDownSuite() {
	suite.server.db.Close()
}
func (suite *appDeleteTestSuite) TestDelete() {
	lenght := 30

	wg := new(sync.WaitGroup)
	for range []int{3, 2, 1} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			links := generatedLinks(lenght)
			shorts := make([]string, 0, lenght)
			pairs := make(map[string]string)
			jar1, err := cookiejar.New(nil)
			suite.Require().NoError(err)
			client := resty.New().SetCookieJar(jar1)
			for _, l := range links {
				resp, err := client.R().SetBody(l).Post("http://" + suite.server.Config.Address())
				suite.Assert().Equal(http.StatusCreated, resp.StatusCode())
				suite.Require().NoError(err)
				sl := strings.Split(string(resp.Body()), "/")
				pairs[string(resp.Body())] = l
				shorts = append(shorts, sl[len(sl)-1])
			}
			body, err := json.Marshal(shorts)
			suite.Require().NoError(err)
			resp, err := client.R().SetHeader("Content-Type", "application/json").SetBody(body).Delete("http://" + suite.server.Config.Address() + "/api/user/urls")
			suite.Require().NoError(err)
			suite.Assert().Equal(http.StatusAccepted, resp.StatusCode())

			ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
			defer cancel()
			for {
				if len(pairs) == 0 {
					break
				}
				for short := range pairs {
					select {
					case <-ctx.Done():
						suite.Assert().Equal(0, len(pairs))
						return
					default:
						resp, _ := client.R().Get(short)
						if resp != nil && resp.StatusCode() == http.StatusGone {
							delete(pairs, short)
						}
					}
				}
			}
		}()
	}
	wg.Wait()

}
func TestAppTestSuiteDelete(t *testing.T) {
	suite.Run(t, new(appDeleteTestSuite))
}

func generatedLinks(count int) []string {
	links := make([]string, 0, count)
	for i := 0; i < count; i++ {
		tmpURL = tmpURL + "b"
		links = append(links, tmpURL)
	}
	return links
}
