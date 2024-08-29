package server

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	pb "github.com/kTowkA/shortener/internal/grpc/proto"
	"github.com/kTowkA/shortener/internal/model"
	"github.com/kTowkA/shortener/internal/storage"
	mocks "github.com/kTowkA/shortener/internal/storage/mocs"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var ctxDuration = time.Second * 10

type Test struct {
	name            string
	req             any
	ctxReq          context.Context
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
	suite.gs.logger = slog.Default()
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

	ctxWithUserID := metadata.NewIncomingContext(ctx, metadata.New(map[string]string{keyUserID: uuid.New().String()}))
	tests := []Test{
		{
			name: "плохой URL",
			req: &pb.EncodeURLRequest{
				OriginalUrl: " http://foo.com",
			},
			ctxReq:          ctxWithUserID,
			wantError:       true,
			wantErrorStatus: codes.InvalidArgument,
		},
		{
			name: "не было uuid",
			req: &pb.EncodeURLRequest{
				OriginalUrl: "http://foo.com",
			},
			ctxReq:          ctx,
			wantError:       true,
			wantErrorStatus: codes.Unauthenticated,
		},
		{
			name: "плохой uuid",
			req: &pb.EncodeURLRequest{
				OriginalUrl: "http://foo.com",
			},
			ctxReq:          metadata.NewIncomingContext(ctx, metadata.New(map[string]string{keyUserID: "123"})),
			wantError:       true,
			wantErrorStatus: codes.InvalidArgument,
		},
		{
			name: "конфликт",
			req: &pb.EncodeURLRequest{
				OriginalUrl: "http://foo.com",
			},
			ctxReq:       ctxWithUserID,
			wantError:    false,
			wantResponse: &pb.EncodeURLResponse{Error: storage.ErrURLConflict.Error()},
			mockFunc: func() {
				suite.mockStorage.On("SaveURL", mock.Anything, mock.Anything, "http://foo.com", mock.AnythingOfType("string")).Return("123", storage.ErrURLConflict)
			},
		},
		{
			name: "внутренняя ошибка",
			req: &pb.EncodeURLRequest{
				OriginalUrl: "http://foo1.com",
			},
			ctxReq:    ctxWithUserID,
			wantError: true,
			mockFunc: func() {
				suite.mockStorage.On("SaveURL", mock.Anything, mock.Anything, "http://foo1.com", mock.AnythingOfType("string")).Return("", errors.New("encode error"))
			},
		},
		{
			name: "все хорошо",
			req: &pb.EncodeURLRequest{
				OriginalUrl: "http://foo2.com",
			},
			ctxReq:       ctxWithUserID,
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

		resp, err := suite.gs.EncodeURL(t.ctxReq, (t.req).(*pb.EncodeURLRequest))
		if !t.wantError {
			wr := (t.wantResponse).(*pb.EncodeURLResponse)
			if wr.Error != "" {
				suite.EqualValues(wr.Error, resp.Error, t.name)
			} else {
				suite.NotEmpty(resp.SavedLink, t.name)
			}
			continue
		}
		suite.Error(err, t.name)
		if t.wantErrorStatus > 0 {
			if e, ok := status.FromError(err); ok {
				suite.EqualValues(t.wantErrorStatus, e.Code(), t.name)
			} else {
				suite.Fail("должна содержаться ошибка", t.name)
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
func (suite *GRPCSuite) TestBatch() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxDuration)
	defer cancel()

	ctxWithUserID := metadata.NewIncomingContext(ctx, metadata.New(map[string]string{keyUserID: uuid.New().String()}))

	tests := []Test{
		{
			name: "в запросе не было uuid пользователя",
			req: &pb.BatchRequest{Elements: []*pb.BatchRequest_BatchRequestElement{
				{
					OriginalUrl: "123",
				},
			}},
			ctxReq:          ctx,
			wantError:       true,
			wantErrorStatus: codes.Unauthenticated,
		},
		{
			name: "uuid пользователя неверен",
			req: &pb.BatchRequest{Elements: []*pb.BatchRequest_BatchRequestElement{
				{
					OriginalUrl: "123",
				},
			}},
			ctxReq:          metadata.NewIncomingContext(ctx, metadata.New(map[string]string{keyUserID: "123"})),
			wantError:       true,
			wantErrorStatus: codes.InvalidArgument,
		},
		{
			name:      "ошибка при сохранении",
			req:       &pb.BatchRequest{},
			ctxReq:    ctxWithUserID,
			wantError: true,
			mockFunc: func() {
				suite.mockStorage.On("Batch", mock.Anything, mock.Anything, mock.Anything).Return(model.BatchResponse{}, errors.New("batch error")).Once()
			},
		},
		{
			name: "все хорошо",
			req: &pb.BatchRequest{
				Elements: []*pb.BatchRequest_BatchRequestElement{{OriginalUrl: "333"}},
			},
			ctxReq:    ctxWithUserID,
			wantError: false,
			mockFunc: func() {
				suite.mockStorage.On("Batch", mock.Anything, mock.Anything, mock.Anything).Return(model.BatchResponse{model.BatchResponseElement{OriginalURL: "333", ShortURL: "3"}}, nil).Once()
			},
			wantResponse: &pb.BatchResponse{Result: []*pb.BatchResponse_Result{{ShortUrl: "3"}}},
		},
	}
	for _, t := range tests {
		if t.mockFunc != nil {
			t.mockFunc()
		}
		resp, err := suite.gs.Batch(t.ctxReq, (t.req).(*pb.BatchRequest))
		if !t.wantError {
			wr := (t.wantResponse).(*pb.BatchResponse)
			suite.EqualValues(wr.Result[0].ShortUrl, resp.Result[0].ShortUrl)
			continue
		}
		suite.Error(err, t.name)
		if t.wantErrorStatus > 0 {
			if e, ok := status.FromError(err); ok {
				suite.EqualValues(t.wantErrorStatus, e.Code(), t.name)
			} else {
				suite.Fail("должна содержаться ошибка", t.name)
			}
		}
	}
}
func (suite *GRPCSuite) TestGetUserURLs() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxDuration)
	defer cancel()

	ctxWithUserID := metadata.NewIncomingContext(ctx, metadata.New(map[string]string{keyUserID: uuid.New().String()}))

	goodResult := []model.StorageJSON{
		{ShortURL: "1", OriginalURL: "111", UUID: "1", IsDeleted: true},
		{ShortURL: "2", OriginalURL: "222", UUID: "2", IsDeleted: false},
	}
	tests := []Test{
		{
			name:            "в запросе не было uuid пользователя",
			req:             &pb.UserURLsRequest{},
			ctxReq:          ctx,
			wantError:       true,
			wantErrorStatus: codes.Unauthenticated,
		},
		{
			name:            "uuid пользователя неверен",
			req:             &pb.UserURLsRequest{},
			ctxReq:          metadata.NewIncomingContext(ctx, metadata.New(map[string]string{keyUserID: "123"})),
			wantError:       true,
			wantErrorStatus: codes.InvalidArgument,
		},
		{
			name:      "ничего не найдено",
			req:       &pb.UserURLsRequest{},
			ctxReq:    ctxWithUserID,
			wantError: true,
			mockFunc: func() {
				suite.mockStorage.On("UserURLs", mock.Anything, mock.Anything).Return([]model.StorageJSON{}, storage.ErrURLNotFound).Once()
			},
			wantErrorStatus: codes.NotFound,
		},
		{
			name:      "ошибка при запросе",
			req:       &pb.UserURLsRequest{},
			ctxReq:    ctxWithUserID,
			wantError: true,
			mockFunc: func() {
				suite.mockStorage.On("UserURLs", mock.Anything, mock.Anything).Return([]model.StorageJSON{}, errors.New("get user urls error")).Once()
			},
		},
		{
			name:      "все хорошо",
			req:       &pb.UserURLsRequest{},
			ctxReq:    ctxWithUserID,
			wantError: false,
			mockFunc: func() {
				suite.mockStorage.On("UserURLs", mock.Anything, mock.Anything).Return(goodResult, nil).Once()
			},
			wantResponse: modelStorageJSONToUserURLsResponse(goodResult),
		},
	}
	for _, t := range tests {
		if t.mockFunc != nil {
			t.mockFunc()
		}
		resp, err := suite.gs.UserURLs(t.ctxReq, (t.req).(*pb.UserURLsRequest))
		if !t.wantError {
			wr := (t.wantResponse).(*pb.UserURLsResponse)
			suite.EqualValues(wr, resp)
			continue
		}
		suite.Error(err, t.name)
		if t.wantErrorStatus > 0 {
			if e, ok := status.FromError(err); ok {
				suite.EqualValues(t.wantErrorStatus, e.Code(), t.name)
			} else {
				suite.Fail("должна содержаться ошибка", t.name)
			}
		}
	}
}
func (suite *GRPCSuite) TestDeleteUserURLs() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxDuration)
	defer cancel()

	userID := uuid.New()
	ctxWithUserID := metadata.NewIncomingContext(ctx, metadata.New(map[string]string{keyUserID: userID.String()}))

	tests := []Test{
		{
			name:            "в запросе не было uuid пользователя",
			req:             &pb.DelUserRequest{},
			ctxReq:          ctx,
			wantError:       true,
			wantErrorStatus: codes.Unauthenticated,
		},
		{
			name:            "uuid пользователя неверен",
			req:             &pb.DelUserRequest{},
			ctxReq:          metadata.NewIncomingContext(ctx, metadata.New(map[string]string{keyUserID: "123"})),
			wantError:       true,
			wantErrorStatus: codes.InvalidArgument,
		},
		{
			name:      "ошибка при запросе",
			req:       &pb.DelUserRequest{ShortUrls: []string{"1", "2"}},
			ctxReq:    ctxWithUserID,
			wantError: true,
			mockFunc: func() {
				suite.mockStorage.On("DeleteURLs", mock.Anything, []model.DeleteURLMessage{{UserID: userID.String(), ShortURL: "1"}, {UserID: userID.String(), ShortURL: "2"}}).Return(errors.New("delete user urls error")).Once()
			},
		},
		{
			name:      "все хорошо",
			req:       &pb.DelUserRequest{ShortUrls: []string{"1", "2"}},
			ctxReq:    ctxWithUserID,
			wantError: false,
			mockFunc: func() {
				suite.mockStorage.On("DeleteURLs", mock.Anything, []model.DeleteURLMessage{{UserID: userID.String(), ShortURL: "1"}, {UserID: userID.String(), ShortURL: "2"}}).Return(nil).Once()
			},
		},
	}
	for _, t := range tests {
		if t.mockFunc != nil {
			t.mockFunc()
		}
		_, err := suite.gs.DeleteUserURLs(t.ctxReq, (t.req).(*pb.DelUserRequest))
		if !t.wantError {
			continue
		}
		suite.Error(err, t.name)
		if t.wantErrorStatus > 0 {
			if e, ok := status.FromError(err); ok {
				suite.EqualValues(t.wantErrorStatus, e.Code(), t.name)
			} else {
				suite.Fail("должна содержаться ошибка", t.name)
			}
		}
	}
}
func TestAppSuite(t *testing.T) {
	suite.Run(t, new(GRPCSuite))
}
