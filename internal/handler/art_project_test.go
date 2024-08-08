package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mirai-box/mirai-box/internal/handler"
	"github.com/mirai-box/mirai-box/internal/middleware"
	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/mocks"
)

func setupArtProjectTestServer(t *testing.T) (*httptest.Server, *mocks.ArtProjectService) {
	r := chi.NewRouter()

	mockService := mocks.NewArtProjectService(t)
	cookieStore := sessions.NewCookieStore([]byte("abc"))
	userMock := mocks.NewUserService(t)
	m := middleware.NewMiddleware(cookieStore, userMock)

	artProjectHandler := handler.NewArtProjectHandler(mockService)

	r.Get("/art/{artID}", artProjectHandler.GetArtByID)

	r.Route("/self", func(r chi.Router) {
		r.Use(m.MockAuthMiddleware)

		r.Get("/artprojects", artProjectHandler.MyArtProjects)
		r.With(middleware.ValidateUUID("artID")).Get("/artprojects/{artID}", artProjectHandler.MyArtProjectByID)

		r.With(middleware.ValidateUUID("artID")).
			Get("/artprojects/{artID}/revisions", artProjectHandler.ListRevisions)
		r.With(middleware.ValidateUUID("artID")).
			Post("/artprojects/{artID}/revisions", artProjectHandler.AddRevision)
		r.Post("/artprojects", artProjectHandler.CreateArtProject)
		r.With(middleware.ValidateUUID("artID")).
			With(middleware.ValidateUUID("revisionID")).
			Get("/artprojects/{artID}/revisions/{revisionID}", artProjectHandler.RevisionDownload)
	})

	return httptest.NewServer(r), mockService
}

func TestArtProjectHandler_CreateArtProject(t *testing.T) {
	server, mockService := setupArtProjectTestServer(t)
	defer server.Close()

	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		artProject := &model.ArtProject{
			ID:       uuid.New(),
			UserID:   userID,
			Title:    "Test Project",
			Filename: "test.png",
		}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("title", artProject.Title)
		part, _ := writer.CreateFormFile("file", artProject.Filename)
		_, _ = io.WriteString(part, "fake image content")
		writer.Close()

		mockService.On("CreateArtProject", mock.Anything, mock.AnythingOfType("*model.ArtProject")).Return(nil).Once()
		mockService.On("AddRevision", mock.Anything, mock.AnythingOfType("*model.Revision"), mock.Anything).Return(nil).Once()

		req, _ := http.NewRequest("POST", server.URL+"/self/artprojects", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response model.ArtProjectResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, artProject.Title, response.Title)
		assert.Equal(t, artProject.UserID, response.UserID)

		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("POST", server.URL+"/self/artprojects", nil)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Missing Title", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.png")
		_, _ = io.WriteString(part, "fake image content")
		writer.Close()

		req, _ := http.NewRequest("POST", server.URL+"/self/artprojects", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Missing File", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("title", "Test Project")
		writer.Close()

		req, _ := http.NewRequest("POST", server.URL+"/self/artprojects", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Service Error - CreateArtProject", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("title", "Test Project")
		part, _ := writer.CreateFormFile("file", "test.png")
		_, _ = io.WriteString(part, "fake image content")
		writer.Close()

		mockService.On("CreateArtProject", mock.Anything, mock.AnythingOfType("*model.ArtProject")).Return(errors.New("database error")).Once()

		req, _ := http.NewRequest("POST", server.URL+"/self/artprojects", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Service Error - AddRevision", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("title", "Test Project")
		part, _ := writer.CreateFormFile("file", "test.png")
		_, _ = io.WriteString(part, "fake image content")
		writer.Close()

		mockService.On("CreateArtProject", mock.Anything, mock.AnythingOfType("*model.ArtProject")).Return(nil).Once()
		mockService.On("AddRevision", mock.Anything, mock.AnythingOfType("*model.Revision"), mock.Anything).Return(errors.New("database error")).Once()

		req, _ := http.NewRequest("POST", server.URL+"/self/artprojects", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestArtProjectHandler_AddRevision(t *testing.T) {
	server, mockService := setupArtProjectTestServer(t)
	defer server.Close()

	userID := uuid.New()
	artProjectID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("comment", "Test Revision")
		part, _ := writer.CreateFormFile("file", "test.png")
		_, _ = io.WriteString(part, "fake image content")
		writer.Close()

		mockService.On("AddRevision", mock.Anything, mock.AnythingOfType("*model.Revision"), mock.Anything).Return(nil).Once()

		req, _ := http.NewRequest("POST", server.URL+"/self/artprojects/"+artProjectID.String()+"/revisions", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response model.RevisionResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "Test Revision", response.Comment)
		assert.Equal(t, userID, response.UserID)
		assert.Equal(t, artProjectID, response.ArtProjectID)

		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("POST", server.URL+"/self/artprojects/"+artProjectID.String()+"/revisions", nil)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Missing File", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("comment", "Test Revision")
		writer.Close()

		req, _ := http.NewRequest("POST", server.URL+"/self/artprojects/"+artProjectID.String()+"/revisions", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Service Error", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("comment", "Test Revision")
		part, _ := writer.CreateFormFile("file", "test.png")
		_, _ = io.WriteString(part, "fake image content")
		writer.Close()

		mockService.On("AddRevision", mock.Anything, mock.AnythingOfType("*model.Revision"), mock.Anything).Return(errors.New("database error")).Once()

		req, _ := http.NewRequest("POST", server.URL+"/self/artprojects/"+artProjectID.String()+"/revisions", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Art Project ID", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("comment", "Test Revision")
		part, _ := writer.CreateFormFile("file", "test.png")
		_, _ = io.WriteString(part, "fake image content")
		writer.Close()

		req, _ := http.NewRequest("POST", server.URL+"/self/artprojects/invalid-uuid/revisions", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestArtProjectHandler_ListRevisions(t *testing.T) {
	server, mockService := setupArtProjectTestServer(t)
	defer server.Close()

	userID := uuid.New()
	artProjectID := uuid.New()

	revisions := []model.Revision{
		{ID: uuid.New(), ArtProjectID: artProjectID, UserID: userID, Comment: "Revision 1"},
		{ID: uuid.New(), ArtProjectID: artProjectID, UserID: userID, Comment: "Revision 2"},
	}

	t.Run("Success", func(t *testing.T) {
		mockService.On("ListRevisions", mock.Anything, artProjectID.String()).Return(revisions, nil).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects/"+artProjectID.String()+"/revisions", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response []model.RevisionResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 2)
		assert.Equal(t, revisions[0].ID, response[0].ID)
		assert.Equal(t, revisions[1].ID, response[1].ID)

		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects/"+artProjectID.String()+"/revisions", nil)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("No Revisions Found", func(t *testing.T) {
		mockService.On("ListRevisions", mock.Anything, artProjectID.String()).Return([]model.Revision{}, nil).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects/"+artProjectID.String()+"/revisions", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response []model.RevisionResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 0)

		mockService.AssertExpectations(t)
	})

	t.Run("Service Error", func(t *testing.T) {
		mockService.On("ListRevisions", mock.Anything, artProjectID.String()).Return(nil, errors.New("database error")).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects/"+artProjectID.String()+"/revisions", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("User not authorized", func(t *testing.T) {
		mockService.On("ListRevisions", mock.Anything, artProjectID.String()).
			Return(revisions, nil).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects/"+artProjectID.String()+"/revisions", nil)
		req.Header.Set("X-User-ID", uuid.NewString())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestArtProjectHandler_RevisionDownload(t *testing.T) {
	server, mockService := setupArtProjectTestServer(t)
	defer server.Close()

	userID := uuid.New()
	artProjectID := uuid.New()
	revisionID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockFile := &bytes.Buffer{}
		_, _ = io.WriteString(mockFile, "fake image content")

		mockArtProject := &model.ArtProject{
			ID:          artProjectID,
			UserID:      userID,
			ContentType: "image/png",
			Filename:    "test.png",
		}

		// Properly set up the mock expectation
		mockService.On("GetArtProjectByRevision", mock.Anything, userID.String(), artProjectID.String(), revisionID.String()).
			Return(io.NopCloser(bytes.NewBuffer(mockFile.Bytes())), mockArtProject, nil).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects/"+artProjectID.String()+"/revisions/"+revisionID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, mockArtProject.ContentType, resp.Header.Get("Content-Type"))
		assert.Contains(t, resp.Header.Get("Content-Disposition"), mockArtProject.Filename)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, "fake image content", string(body))

		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects/"+artProjectID.String()+"/revisions/"+revisionID.String(), nil)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Not Found", func(t *testing.T) {
		// Set up mock to return ErrArtProjectNotFound
		mockService.On("GetArtProjectByRevision", mock.Anything, userID.String(), artProjectID.String(), revisionID.String()).
			Return(nil, nil, model.ErrArtProjectNotFound).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects/"+artProjectID.String()+"/revisions/"+revisionID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects/invalid-uuid/revisions/"+revisionID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Service Error", func(t *testing.T) {
		// Set up mock to return an error
		mockService.On("GetArtProjectByRevision", mock.Anything, userID.String(), artProjectID.String(), revisionID.String()).
			Return(nil, nil, errors.New("service error")).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects/"+artProjectID.String()+"/revisions/"+revisionID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestArtProjectHandler_GetArtByID(t *testing.T) {
	server, mockService := setupArtProjectTestServer(t)
	defer server.Close()

	artID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		revision := &model.Revision{
			ID:       uuid.New(),
			ArtID:    artID.String(),
			FilePath: "testing_data/1.png",
		}

		mockService.On("GetRevisionByArtID", mock.Anything, artID.String()).
			Return(revision, nil).Once()

		req, _ := http.NewRequest("GET", server.URL+"/art/"+artID.String(), nil)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockService.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockService.On("GetRevisionByArtID", mock.Anything, artID.String()).
			Return(nil, model.ErrArtProjectNotFound).Once()

		req, _ := http.NewRequest("GET", server.URL+"/art/"+artID.String(), nil)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Service Error", func(t *testing.T) {
		mockService.On("GetRevisionByArtID", mock.Anything, artID.String()).Return(nil, errors.New("database error")).Once()

		req, _ := http.NewRequest("GET", server.URL+"/art/"+artID.String(), nil)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestArtProjectHandler_MyArtProjects(t *testing.T) {
	server, mockService := setupArtProjectTestServer(t)
	defer server.Close()

	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		artProjects := []model.ArtProject{
			{ID: uuid.New(), UserID: userID, Title: "Project 1"},
			{ID: uuid.New(), UserID: userID, Title: "Project 2"},
		}

		mockService.On("FindByUserID", mock.Anything, userID.String()).
			Return(artProjects, nil).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response []model.ArtProjectResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 2)
		assert.Equal(t, artProjects[0].ID, response[0].ID)
		assert.Equal(t, artProjects[1].ID, response[1].ID)

		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects", nil)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("No Projects Found", func(t *testing.T) {
		mockService.On("FindByUserID", mock.Anything, userID.String()).Return([]model.ArtProject{}, nil).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response []model.ArtProjectResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 0)

		mockService.AssertExpectations(t)
	})

	t.Run("No Projects Found with error", func(t *testing.T) {
		mockService.On("FindByUserID", mock.Anything, userID.String()).
			Return([]model.ArtProject{}, model.ErrArtProjectNotFound).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response []model.ArtProjectResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 0)

		mockService.AssertExpectations(t)
	})

	t.Run("Service Error", func(t *testing.T) {
		mockService.On("FindByUserID", mock.Anything, userID.String()).Return(nil, errors.New("database error")).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestArtProjectHandler_MyArtProjectByID(t *testing.T) {
	server, mockService := setupArtProjectTestServer(t)
	defer server.Close()

	userID := uuid.New()
	artProjectID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		artProject := &model.ArtProject{ID: artProjectID, UserID: userID, Title: "My Project"}

		mockService.On("FindByID", mock.Anything, artProjectID.String()).
			Return(artProject, nil).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects/"+artProjectID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response model.ArtProjectResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, artProject.ID, response.ID)
		assert.Equal(t, artProject.Title, response.Title)

		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects/"+artProjectID.String(), nil)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockService.On("FindByID", mock.Anything, artProjectID.String()).Return(nil, model.ErrArtProjectNotFound).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects/"+artProjectID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Server error", func(t *testing.T) {
		mockService.On("FindByID", mock.Anything, artProjectID.String()).
			Return(nil, fmt.Errorf("some error")).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects/"+artProjectID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Not Owner", func(t *testing.T) {
		otherUserID := uuid.New()
		artProject := &model.ArtProject{ID: artProjectID, UserID: otherUserID, Title: "Other User's Project"}

		mockService.On("FindByID", mock.Anything, artProjectID.String()).Return(artProject, nil).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects/"+artProjectID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/self/artprojects/invalid-uuid", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
