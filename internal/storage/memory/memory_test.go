package memory

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kTowkA/shortener/internal/model"
	"github.com/kTowkA/shortener/internal/storage"
	"github.com/stretchr/testify/suite"
)

type memorySuite struct {
	suite.Suite
	*Storage
}

type eqValue struct {
	ShortURL    string `json:"short_url,omitempty"`
	OriginalURL string `json:"original_url,omitempty"`
	IsDeleted   bool   `json:"is_deleted"`
}

func convert(x []model.StorageJSON) []eqValue {
	values := make([]eqValue, len(x))
	for i := range x {
		values[i] = eqValue{
			ShortURL:    x[i].ShortURL,
			OriginalURL: x[i].OriginalURL,
			IsDeleted:   x[i].IsDeleted,
		}
	}
	return values
}

func (suite *memorySuite) SetupSuite() {
	st, err := NewStorage("")
	suite.Require().NoError(err)
	suite.Storage = st
}
func (suite *memorySuite) TearDownSuite() {
	err := suite.Storage.Close()
	suite.Require().NoError(err)
}

func (suite *memorySuite) TestSaveURL() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user1 := uuid.New()
	tests := []struct {
		name          string
		userID        uuid.UUID
		real          string
		short         string
		expectedValue string
		expectedError error
	}{
		{
			"успешное сохранение",
			user1,
			"TestSaveURL_1_1",
			"TestSaveURL_1_2",
			"TestSaveURL_1_2",
			nil,
		},
		{
			"урл уже существует",
			uuid.New(),
			"TestSaveURL_1_1",
			"TestSaveURL_1_2",
			"",
			storage.ErrURLIsExist,
		},
		{
			"конфликт",
			user1,
			"TestSaveURL_1_1",
			"TestSaveURL_1_3",
			"TestSaveURL_1_2",
			storage.ErrURLConflict,
		},
	}
	for _, tt := range tests {
		r, err := suite.SaveURL(ctx, tt.userID, tt.real, tt.short)
		suite.EqualValues(tt.expectedError, err, tt.name)
		suite.EqualValues(tt.expectedValue, r, tt.name)
	}
}

func (suite *memorySuite) TestRealURL() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	real := "TestRealURL_1_1"
	short := "TestRealURL_1_2"
	user := uuid.New()

	// сохраняем
	_, err := suite.SaveURL(ctx, user, real, short)
	suite.NoError(err)

	tests := []struct {
		name          string
		short         string
		expectedValue model.StorageJSON
		expectedError error
	}{
		{
			name:          "ничего нет",
			short:         "TestRealURL_1_22",
			expectedValue: model.StorageJSON{},
			expectedError: storage.ErrURLNotFound,
		},
		{
			name:  "нашли",
			short: short,
			expectedValue: model.StorageJSON{
				OriginalURL: real,
				IsDeleted:   false,
			},
			expectedError: nil,
		},
	}
	for _, tt := range tests {
		resp, err := suite.RealURL(ctx, tt.short)
		suite.EqualValues(tt.expectedValue, resp, tt.name)
		suite.EqualValues(tt.expectedError, err, tt.name)
	}
}

func (suite *memorySuite) TestPing() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := suite.Ping(ctx)
	suite.NoError(err)
}

func (suite *memorySuite) TestBatch() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user := uuid.New()
	tests := []struct {
		name          string
		userID        uuid.UUID
		values        model.BatchRequest
		expectedValue model.BatchResponse
		expectedError error
	}{
		{
			"нормально",
			user,
			model.BatchRequest{
				{
					CorrelationID: "TestBatch_1_1",
					ShortURL:      "TestBatch_1_2",
					OriginalURL:   "TestBatch_1_3",
				},
			},
			model.BatchResponse{
				{
					CorrelationID: "TestBatch_1_1",
					OriginalURL:   "TestBatch_1_3",
				},
			},
			nil,
		},
		{
			"коллизия",
			user,
			model.BatchRequest{
				{
					CorrelationID: "TestBatch_2_1",
					ShortURL:      "TestBatch_1_2",
					OriginalURL:   "TestBatch_2_3",
				},
			},
			model.BatchResponse{
				{
					CorrelationID: "TestBatch_2_1",
					OriginalURL:   "TestBatch_2_3",
					Error:         storage.ErrURLIsExist,
					Collision:     true,
				},
			},
			nil,
		},
		{
			"конфликт",
			user,
			model.BatchRequest{
				{
					CorrelationID: "TestBatch_3_1",
					ShortURL:      "TestBatch_3_2",
					OriginalURL:   "TestBatch_1_3",
				},
			},
			model.BatchResponse{
				{
					CorrelationID: "TestBatch_3_1",
					OriginalURL:   "TestBatch_1_3",
					ShortURL:      "TestBatch_1_2",
					Error:         storage.ErrURLConflict,
					Collision:     false,
				},
			},
			nil,
		},
	}

	for _, tt := range tests {
		resp, _ := suite.Batch(ctx, tt.userID, tt.values)
		suite.EqualValues(tt.expectedValue, resp, tt.name)
	}
}

func (suite *memorySuite) TestUserURLs() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user1 := uuid.New()
	user2 := uuid.New()
	user3 := uuid.New()

	tests := []struct {
		name          string
		user          uuid.UUID
		values        model.BatchRequest
		expectedValue []eqValue
		expectedError error
	}{
		{
			name: "первый пользователь",
			user: user1,
			values: model.BatchRequest{
				{
					CorrelationID: "TestUserURLs_1_1",
					OriginalURL:   "TestUserURLs_1_2",
					ShortURL:      "TestUserURLs_1_3",
				},
				{
					CorrelationID: "TestUserURLs_2_1",
					OriginalURL:   "TestUserURLs_2_2",
					ShortURL:      "TestUserURLs_2_3",
				},
			},
			expectedError: nil,
			expectedValue: []eqValue{
				{
					ShortURL:    "TestUserURLs_1_3",
					OriginalURL: "TestUserURLs_1_2",
					IsDeleted:   false,
				},
				{
					ShortURL:    "TestUserURLs_2_3",
					OriginalURL: "TestUserURLs_2_2",
					IsDeleted:   false,
				},
			},
		},
		{
			name: "второй пользователь",
			user: user2,
			values: model.BatchRequest{
				{
					CorrelationID: "TestUserURLs_3_1",
					OriginalURL:   "TestUserURLs_3_2",
					ShortURL:      "TestUserURLs_3_3",
				},
				{
					CorrelationID: "TestUserURLs_4_1",
					OriginalURL:   "TestUserURLs_4_2",
					ShortURL:      "TestUserURLs_4_3",
				},
				{
					CorrelationID: "TestUserURLs_5_1",
					OriginalURL:   "TestUserURLs_5_2",
					ShortURL:      "TestUserURLs_5_3",
				},
			},
			expectedError: nil,
			expectedValue: []eqValue{
				{
					ShortURL:    "TestUserURLs_3_3",
					OriginalURL: "TestUserURLs_3_2",
					IsDeleted:   false,
				},
				{
					ShortURL:    "TestUserURLs_4_3",
					OriginalURL: "TestUserURLs_4_2",
					IsDeleted:   false,
				},
				{
					ShortURL:    "TestUserURLs_5_3",
					OriginalURL: "TestUserURLs_5_2",
					IsDeleted:   false,
				},
			},
		},
		{
			name:          "не найдено",
			user:          user3,
			expectedError: storage.ErrURLNotFound,
		},
	}

	for _, tt := range tests {
		_, err := suite.Batch(ctx, tt.user, tt.values)
		suite.NoError(err, tt.name)
		resp, err := suite.UserURLs(ctx, tt.user)
		suite.EqualValues(tt.expectedError, err, tt.name)
		newResp := convert(resp)
		suite.Len(newResp, len(tt.expectedValue))
		for i := range newResp {
			suite.Contains(tt.expectedValue, newResp[i])
		}
	}
}

func (suite *memorySuite) TestDeleteURLs() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user1 := uuid.New()
	user2 := uuid.New()

	tests := []struct {
		name          string
		user          uuid.UUID
		values        model.BatchRequest
		deletedValues []model.DeleteURLMessage
		expectedValue []eqValue
		expectedError error
	}{
		{
			name: "первый пользователь",
			user: user1,
			values: model.BatchRequest{
				{
					CorrelationID: "TestDeleteURLs_1_1",
					OriginalURL:   "TestDeleteURLs_1_2",
					ShortURL:      "TestDeleteURLs_1_3",
				},
				{
					CorrelationID: "TestDeleteURLs_2_1",
					OriginalURL:   "TestDeleteURLs_2_2",
					ShortURL:      "TestDeleteURLs_2_3",
				},
			},
			deletedValues: []model.DeleteURLMessage{
				{
					UserID:   user1.String(),
					ShortURL: "TestDeleteURLs_1_3",
				},
			},
			expectedError: nil,
			expectedValue: []eqValue{
				{
					ShortURL:    "TestDeleteURLs_1_3",
					OriginalURL: "TestDeleteURLs_1_2",
					IsDeleted:   true,
				},
				{
					ShortURL:    "TestDeleteURLs_2_3",
					OriginalURL: "TestDeleteURLs_2_2",
					IsDeleted:   false,
				},
			},
		},
		{
			name: "второй пользователь",
			user: user2,
			values: model.BatchRequest{
				{
					CorrelationID: "TestDeleteURLs_3_1",
					OriginalURL:   "TestDeleteURLs_3_2",
					ShortURL:      "TestDeleteURLs_3_3",
				},
				{
					CorrelationID: "TestDeleteURLs_4_1",
					OriginalURL:   "TestDeleteURLs_4_2",
					ShortURL:      "TestDeleteURLs_4_3",
				},
				{
					CorrelationID: "TestDeleteURLs_5_1",
					OriginalURL:   "TestDeleteURLs_5_2",
					ShortURL:      "TestDeleteURLs_5_3",
				},
			},
			deletedValues: []model.DeleteURLMessage{
				{
					UserID:   user2.String(),
					ShortURL: "TestDeleteURLs_4_3",
				},
			},
			expectedError: nil,
			expectedValue: []eqValue{
				{
					ShortURL:    "TestDeleteURLs_3_3",
					OriginalURL: "TestDeleteURLs_3_2",
					IsDeleted:   false,
				},
				{
					ShortURL:    "TestDeleteURLs_4_3",
					OriginalURL: "TestDeleteURLs_4_2",
					IsDeleted:   true,
				},
				{
					ShortURL:    "TestDeleteURLs_5_3",
					OriginalURL: "TestDeleteURLs_5_2",
					IsDeleted:   false,
				},
			},
		},
	}

	for _, tt := range tests {
		_, err := suite.Batch(ctx, tt.user, tt.values)
		suite.NoError(err, tt.name)
		err = suite.DeleteURLs(ctx, tt.deletedValues)
		suite.NoError(err, tt.name)
		resp, err := suite.UserURLs(ctx, tt.user)
		suite.EqualValues(tt.expectedError, err, tt.name)

		newResp := convert(resp)
		suite.Len(newResp, len(tt.expectedValue))
		for i := range newResp {
			suite.Contains(tt.expectedValue, newResp[i])
		}
	}
}
func (suite *memorySuite) TestStats() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// тут добавляем одно сохранение, чтобы в базе точно было одно значение
	_, err := suite.SaveURL(ctx, uuid.New(), "test_stat_1", "test_stat_2")
	suite.NoError(err)

	stats, err := suite.Stats(ctx)
	suite.NoError(err)
	suite.GreaterOrEqual(stats.TotalURLs, 1)
	suite.GreaterOrEqual(stats.TotalUsers, 1)
}
func TestMemorySuite(t *testing.T) {
	suite.Run(t, new(memorySuite))
}
