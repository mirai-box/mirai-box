// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	context "context"
	io "io"

	mock "github.com/stretchr/testify/mock"

	model "github.com/mirai-box/mirai-box/internal/model"
)

// ArtProjectService is an autogenerated mock type for the ArtProjectService type
type ArtProjectService struct {
	mock.Mock
}

// AddRevision provides a mock function with given fields: ctx, revision, fileData
func (_m *ArtProjectService) AddRevision(ctx context.Context, revision *model.Revision, fileData io.Reader) error {
	ret := _m.Called(ctx, revision, fileData)

	if len(ret) == 0 {
		panic("no return value specified for AddRevision")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.Revision, io.Reader) error); ok {
		r0 = rf(ctx, revision, fileData)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateArtProject provides a mock function with given fields: ctx, artProject
func (_m *ArtProjectService) CreateArtProject(ctx context.Context, artProject *model.ArtProject) error {
	ret := _m.Called(ctx, artProject)

	if len(ret) == 0 {
		panic("no return value specified for CreateArtProject")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.ArtProject) error); ok {
		r0 = rf(ctx, artProject)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteArtProject provides a mock function with given fields: ctx, id
func (_m *ArtProjectService) DeleteArtProject(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for DeleteArtProject")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FindByID provides a mock function with given fields: ctx, id
func (_m *ArtProjectService) FindByID(ctx context.Context, id string) (*model.ArtProject, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for FindByID")
	}

	var r0 *model.ArtProject
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.ArtProject, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.ArtProject); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ArtProject)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindByUserID provides a mock function with given fields: ctx, userID
func (_m *ArtProjectService) FindByUserID(ctx context.Context, userID string) ([]model.ArtProject, error) {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for FindByUserID")
	}

	var r0 []model.ArtProject
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]model.ArtProject, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []model.ArtProject); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.ArtProject)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetArtProject provides a mock function with given fields: ctx, id
func (_m *ArtProjectService) GetArtProject(ctx context.Context, id string) (*model.ArtProject, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetArtProject")
	}

	var r0 *model.ArtProject
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.ArtProject, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.ArtProject); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ArtProject)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetArtProjectByRevision provides a mock function with given fields: ctx, userID, artProjectID, revisionID
func (_m *ArtProjectService) GetArtProjectByRevision(ctx context.Context, userID string, artProjectID string, revisionID string) (io.ReadCloser, *model.ArtProject, error) {
	ret := _m.Called(ctx, userID, artProjectID, revisionID)

	if len(ret) == 0 {
		panic("no return value specified for GetArtProjectByRevision")
	}

	var r0 io.ReadCloser
	var r1 *model.ArtProject
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (io.ReadCloser, *model.ArtProject, error)); ok {
		return rf(ctx, userID, artProjectID, revisionID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) io.ReadCloser); ok {
		r0 = rf(ctx, userID, artProjectID, revisionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadCloser)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) *model.ArtProject); ok {
		r1 = rf(ctx, userID, artProjectID, revisionID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.ArtProject)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, string, string, string) error); ok {
		r2 = rf(ctx, userID, artProjectID, revisionID)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetLatestRevision provides a mock function with given fields: ctx, artProjectID
func (_m *ArtProjectService) GetLatestRevision(ctx context.Context, artProjectID string) (*model.Revision, error) {
	ret := _m.Called(ctx, artProjectID)

	if len(ret) == 0 {
		panic("no return value specified for GetLatestRevision")
	}

	var r0 *model.Revision
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.Revision, error)); ok {
		return rf(ctx, artProjectID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Revision); ok {
		r0 = rf(ctx, artProjectID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Revision)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, artProjectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetRevisionByArtID provides a mock function with given fields: ctx, artID
func (_m *ArtProjectService) GetRevisionByArtID(ctx context.Context, artID string) (*model.Revision, error) {
	ret := _m.Called(ctx, artID)

	if len(ret) == 0 {
		panic("no return value specified for GetRevisionByArtID")
	}

	var r0 *model.Revision
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.Revision, error)); ok {
		return rf(ctx, artID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Revision); ok {
		r0 = rf(ctx, artID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Revision)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, artID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListArtProjects provides a mock function with given fields: ctx, userID
func (_m *ArtProjectService) ListArtProjects(ctx context.Context, userID string) ([]model.ArtProject, error) {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for ListArtProjects")
	}

	var r0 []model.ArtProject
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]model.ArtProject, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []model.ArtProject); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.ArtProject)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListRevisions provides a mock function with given fields: ctx, artProjectID
func (_m *ArtProjectService) ListRevisions(ctx context.Context, artProjectID string) ([]model.Revision, error) {
	ret := _m.Called(ctx, artProjectID)

	if len(ret) == 0 {
		panic("no return value specified for ListRevisions")
	}

	var r0 []model.Revision
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]model.Revision, error)); ok {
		return rf(ctx, artProjectID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []model.Revision); ok {
		r0 = rf(ctx, artProjectID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.Revision)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, artProjectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewArtProjectService creates a new instance of ArtProjectService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewArtProjectService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ArtProjectService {
	mock := &ArtProjectService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
