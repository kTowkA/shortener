package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kTowkA/shortener/internal/model"
	"github.com/kTowkA/shortener/internal/storage"
	mocks "github.com/kTowkA/shortener/internal/storage/mocs"
	pb "github.com/kTowkA/shortener/proto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ctxDuration = time.Second * 10

type Test struct {
	name            string
	req             any
	wantError       bool
	wantErrorStatus codes.Code
	wantResponse    any
	mockFunc        func()
}
type GRPCSuite struct {
	suite.Suite
	mockStorage *mocks.Storager
	gs          *ShortenerServer
}

func (suite *GRPCSuite) SetupSuite() {
	// запускаем начальную настройку
	suite.Suite.T().Log("Suite setup")

	suite.mockStorage = new(mocks.Storager)
	// создаем новый экземпляр нашего приложения
	suite.gs = new(ShortenerServer)
	suite.gs.db = suite.mockStorage
}
func (suite *GRPCSuite) TearDownSuite() {
	//заканчиваем работу с тестовым сценарием
	defer suite.Suite.T().Log("Suite test is ended")

	// закрываем мок хранилища
	suite.mockStorage.On("Close").Return(nil).Once()
	err := suite.mockStorage.Close()
	suite.Require().NoError(err, "close mock storage")

	// смотрим чтобы все описанные вызовы были использованы
	suite.mockStorage.AssertExpectations(suite.T())
}
func (suite *GRPCSuite) TestPing() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxDuration)
	defer cancel()

	// fail
	suite.mockStorage.On("Ping", mock.Anything).Return(errors.New("1")).Once()
	_, err := suite.gs.Ping(ctx, nil)
	suite.Error(err)
	if e, ok := status.FromError(err); ok {
		suite.EqualValues(codes.Unavailable, e.Code())
	} else {
		suite.Fail("должна содержаться ошибка")
	}

	// ok
	suite.mockStorage.On("Ping", mock.Anything).Return(nil).Once()
	resp, err := suite.gs.Ping(ctx, nil)
	suite.NoError(err)
	suite.True(resp.Status.Ok)
}
func (suite *GRPCSuite) TestEncodeURL() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxDuration)
	defer cancel()
	tests := []Test{
		{
			name: "плохой URL",
			req: &pb.EncodeURLRequest{
				UserId:      uuid.New().String(),
				OriginalUrl: " http://foo.com",
			},
			wantError:       true,
			wantErrorStatus: codes.InvalidArgument,
		},
		{
			name: "плохой uuid",
			req: &pb.EncodeURLRequest{
				UserId:      "12313123",
				OriginalUrl: "http://foo.com",
			},
			wantError:       true,
			wantErrorStatus: codes.InvalidArgument,
		},
		{
			name: "конфликт",
			req: &pb.EncodeURLRequest{
				UserId:      uuid.New().String(),
				OriginalUrl: "http://foo.com",
			},
			wantError:    false,
			wantResponse: &pb.EncodeURLResponse{Error: storage.ErrURLConflict.Error()},
			mockFunc: func() {
				suite.mockStorage.On("SaveURL", mock.Anything, mock.Anything, "http://foo.com", mock.AnythingOfType("string")).Return("123", storage.ErrURLConflict)
			},
		},
		{
			name: "внутренняя ошибка",
			req: &pb.EncodeURLRequest{
				UserId:      uuid.New().String(),
				OriginalUrl: "http://foo1.com",
			},
			wantError: true,
			mockFunc: func() {
				suite.mockStorage.On("SaveURL", mock.Anything, mock.Anything, "http://foo1.com", mock.AnythingOfType("string")).Return("", errors.New("encode error"))
			},
		},
		{
			name: "все хорошо",
			req: &pb.EncodeURLRequest{
				UserId:      uuid.New().String(),
				OriginalUrl: "http://foo2.com",
			},
			wantError:    false,
			wantResponse: &pb.EncodeURLResponse{},
			mockFunc: func() {
				suite.mockStorage.On("SaveURL", mock.Anything, mock.Anything, "http://foo2.com", mock.AnythingOfType("string")).Return("123", nil)
			},
		},
	}

	for _, t := range tests {
		if t.mockFunc != nil {
			t.mockFunc()
		}

		resp, err := suite.gs.EncodeURL(ctx, (t.req).(*pb.EncodeURLRequest))
		if !t.wantError {
			wr := (t.wantResponse).(*pb.EncodeURLResponse)
			if wr.Error != "" {
				suite.EqualValues(wr.Error, resp.Error)
			} else {
				suite.NotEmpty(resp.SavedLink)
			}
			continue
		}
		suite.Error(err)
		if t.wantErrorStatus > 0 {
			if e, ok := status.FromError(err); ok {
				suite.EqualValues(t.wantErrorStatus, e.Code())
			} else {
				suite.Fail("должна содержаться ошибка")
			}
		}
	}
}
func (suite *GRPCSuite) TestDecode() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxDuration)
	defer cancel()

	tests := []Test{
		{
			name:            "ничего не найдено",
			wantError:       true,
			wantErrorStatus: codes.NotFound,
			req:             &pb.DecodeURLRequest{ShortUrl: "123"},
			mockFunc: func() {
				suite.mockStorage.On("RealURL", mock.Anything, "123").Return(model.StorageJSON{}, storage.ErrURLNotFound)
			},
		},
		{
			name:      "ошибка хранилища",
			wantError: true,
			req:       &pb.DecodeURLRequest{ShortUrl: "234"},
			mockFunc: func() {
				suite.mockStorage.On("RealURL", mock.Anything, "234").Return(model.StorageJSON{}, errors.New("decode error"))
			},
		},
		{
			name:      "было удалено",
			wantError: true,
			req:       &pb.DecodeURLRequest{ShortUrl: "345"},
			mockFunc: func() {
				suite.mockStorage.On("RealURL", mock.Anything, "345").Return(model.StorageJSON{IsDeleted: true}, nil)
			},
		},
		{
			name:      "все хорошо",
			wantError: false,
			req:       &pb.DecodeURLRequest{ShortUrl: "777"},
			mockFunc: func() {
				suite.mockStorage.On("RealURL", mock.Anything, "777").Return(model.StorageJSON{OriginalURL: "666"}, nil)
			},
			wantResponse: &pb.DecodeURLResponse{OriginalUrl: "666"},
		},
	}
	for _, t := range tests {
		if t.mockFunc != nil {
			t.mockFunc()
		}

		resp, err := suite.gs.DecodeURL(ctx, (t.req).(*pb.DecodeURLRequest))
		if !t.wantError {
			suite.EqualValues(t.wantResponse, resp)
			continue
		}
		suite.Error(err)
		if t.wantErrorStatus > 0 {
			if e, ok := status.FromError(err); ok {
				suite.EqualValues(t.wantErrorStatus, e.Code())
			} else {
				suite.Fail("должна содержаться ошибка")
			}
		}
	}
}
func (suite *GRPCSuite) TestStats() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxDuration)
	defer cancel()

	tests := []Test{
		{
			name:      "ошибка хранилища",
			req:       &pb.StatsRequest{},
			wantError: true,
			mockFunc: func() {
				suite.mockStorage.On("Stats", mock.Anything).Return(model.StatsResponse{}, errors.New("stats error")).Once()
			},
		},
		{
			name:      "все хорошо",
			req:       &pb.StatsRequest{},
			wantError: false,
			mockFunc: func() {
				suite.mockStorage.On("Stats", mock.Anything).Return(model.StatsResponse{TotalUsers: 1, TotalURLs: 2}, nil).Once()
			},
		},
	}
	for _, t := range tests {
		if t.mockFunc != nil {
			t.mockFunc()
		}
		resp, err := suite.gs.Stats(ctx, (t.req).(*pb.StatsRequest))
		if !t.wantError {
			suite.EqualValues(&pb.StatsResponse{Users: 1, Urls: 2}, resp)
			continue
		}
		suite.Error(err)
	}
}
func TestAppSuite(t *testing.T) {
	suite.Run(t, new(GRPCSuite))
}
