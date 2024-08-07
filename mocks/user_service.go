// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	context "context"

	model "github.com/mirai-box/mirai-box/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// UserService is an autogenerated mock type for the UserService type
type UserService struct {
	mock.Mock
}

// Authenticate provides a mock function with given fields: ctx, username, password
func (_m *UserService) Authenticate(ctx context.Context, username string, password string) (*model.User, error) {
	ret := _m.Called(ctx, username, password)

	if len(ret) == 0 {
		panic("no return value specified for Authenticate")
	}

	var r0 *model.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*model.User, error)); ok {
		return rf(ctx, username, password)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.User); ok {
		r0 = rf(ctx, username, password)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, username, password)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateUser provides a mock function with given fields: ctx, username, password, role
func (_m *UserService) CreateUser(ctx context.Context, username string, password string, role string) (*model.User, error) {
	ret := _m.Called(ctx, username, password, role)

	if len(ret) == 0 {
		panic("no return value specified for CreateUser")
	}

	var r0 *model.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (*model.User, error)); ok {
		return rf(ctx, username, password, role)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *model.User); ok {
		r0 = rf(ctx, username, password, role)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, username, password, role)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteUser provides a mock function with given fields: ctx, id
func (_m *UserService) DeleteUser(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for DeleteUser")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetStashByUserID provides a mock function with given fields: ctx, userID
func (_m *UserService) GetStashByUserID(ctx context.Context, userID string) (*model.Stash, error) {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetStashByUserID")
	}

	var r0 *model.Stash
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.Stash, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Stash); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Stash)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetStorageUsage provides a mock function with given fields: ctx, userID
func (_m *UserService) GetStorageUsage(ctx context.Context, userID string) (*model.StorageUsage, error) {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetStorageUsage")
	}

	var r0 *model.StorageUsage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.StorageUsage, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.StorageUsage); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.StorageUsage)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUser provides a mock function with given fields: ctx, id
func (_m *UserService) GetUser(ctx context.Context, id string) (*model.User, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetUser")
	}

	var r0 *model.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.User, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.User); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserByUsername provides a mock function with given fields: ctx, username
func (_m *UserService) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	ret := _m.Called(ctx, username)

	if len(ret) == 0 {
		panic("no return value specified for GetUserByUsername")
	}

	var r0 *model.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.User, error)); ok {
		return rf(ctx, username)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.User); ok {
		r0 = rf(ctx, username)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, username)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateStorageUsage provides a mock function with given fields: ctx, storageUsage
func (_m *UserService) UpdateStorageUsage(ctx context.Context, storageUsage *model.StorageUsage) error {
	ret := _m.Called(ctx, storageUsage)

	if len(ret) == 0 {
		panic("no return value specified for UpdateStorageUsage")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.StorageUsage) error); ok {
		r0 = rf(ctx, storageUsage)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateUser provides a mock function with given fields: ctx, user
func (_m *UserService) UpdateUser(ctx context.Context, user *model.User) error {
	ret := _m.Called(ctx, user)

	if len(ret) == 0 {
		panic("no return value specified for UpdateUser")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.User) error); ok {
		r0 = rf(ctx, user)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewUserService creates a new instance of UserService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewUserService(t interface {
	mock.TestingT
	Cleanup(func())
}) *UserService {
	mock := &UserService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
