// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	context "context"

	model "github.com/kTowkA/shortener/internal/model"
	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// Storager is an autogenerated mock type for the Storager type
type Storager struct {
	mock.Mock
}

// Batch provides a mock function with given fields: ctx, userID, values
func (_m *Storager) Batch(ctx context.Context, userID uuid.UUID, values model.BatchRequest) (model.BatchResponse, error) {
	ret := _m.Called(ctx, userID, values)

	if len(ret) == 0 {
		panic("no return value specified for Batch")
	}

	var r0 model.BatchResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, model.BatchRequest) (model.BatchResponse, error)); ok {
		return rf(ctx, userID, values)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, model.BatchRequest) model.BatchResponse); ok {
		r0 = rf(ctx, userID, values)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(model.BatchResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, model.BatchRequest) error); ok {
		r1 = rf(ctx, userID, values)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Close provides a mock function with given fields:
func (_m *Storager) Close() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteURLs provides a mock function with given fields: ctx, deleteLinks
func (_m *Storager) DeleteURLs(ctx context.Context, deleteLinks []model.DeleteURLMessage) error {
	ret := _m.Called(ctx, deleteLinks)

	if len(ret) == 0 {
		panic("no return value specified for DeleteURLs")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []model.DeleteURLMessage) error); ok {
		r0 = rf(ctx, deleteLinks)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Ping provides a mock function with given fields: ctx
func (_m *Storager) Ping(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Ping")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RealURL provides a mock function with given fields: ctx, short
func (_m *Storager) RealURL(ctx context.Context, short string) (model.StorageJSON, error) {
	ret := _m.Called(ctx, short)

	if len(ret) == 0 {
		panic("no return value specified for RealURL")
	}

	var r0 model.StorageJSON
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (model.StorageJSON, error)); ok {
		return rf(ctx, short)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) model.StorageJSON); ok {
		r0 = rf(ctx, short)
	} else {
		r0 = ret.Get(0).(model.StorageJSON)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, short)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SaveURL provides a mock function with given fields: ctx, userID, real, short
func (_m *Storager) SaveURL(ctx context.Context, userID uuid.UUID, real string, short string) (string, error) {
	ret := _m.Called(ctx, userID, real, short)

	if len(ret) == 0 {
		panic("no return value specified for SaveURL")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string, string) (string, error)); ok {
		return rf(ctx, userID, real, short)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string, string) string); ok {
		r0 = rf(ctx, userID, real, short)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, string, string) error); ok {
		r1 = rf(ctx, userID, real, short)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UserURLs provides a mock function with given fields: ctx, userID
func (_m *Storager) UserURLs(ctx context.Context, userID uuid.UUID) ([]model.StorageJSON, error) {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for UserURLs")
	}

	var r0 []model.StorageJSON
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]model.StorageJSON, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []model.StorageJSON); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.StorageJSON)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewStorager creates a new instance of Storager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStorager(t interface {
	mock.TestingT
	Cleanup(func())
}) *Storager {
	mock := &Storager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
