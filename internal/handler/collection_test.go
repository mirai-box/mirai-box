package handler_test

import (
	"bytes"
	"encoding/json"
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

func setupCollectionTestServer(t *testing.T) (*httptest.Server, *mocks.CollectionService) {
	r := chi.NewRouter()

	mockService := mocks.NewCollectionService(t)
	cookieStore := sessions.NewCookieStore([]byte("test-secret"))
	userMock := mocks.NewUserService(t)
	m := middleware.NewMiddleware(cookieStore, userMock)

	collectionHandler := handler.NewCollectionHandler(mockService)

	r.Route("/self/collections", func(r chi.Router) {
		r.Use(m.MockAuthMiddleware)
		r.Post("/", collectionHandler.CreateCollection)
		r.Get("/", collectionHandler.GetUserCollections)
		r.With(middleware.ValidateUUID("id")).Get("/{id}", collectionHandler.GetCollection)
		r.With(middleware.ValidateUUID("id")).Put("/{id}", collectionHandler.UpdateCollection)
		r.With(middleware.ValidateUUID("id")).Delete("/{id}", collectionHandler.DeleteCollection)
		r.With(middleware.ValidateUUID("id")).Post("/{id}/revisions", collectionHandler.AddRevisionToCollection)
		r.With(middleware.ValidateUUID("id")).Get("/{id}/revisions", collectionHandler.ListRevisions)
		r.With(middleware.ValidateUUID("id")).With(middleware.ValidateUUID("revisionID")).
			Delete("/{id}/revisions/{revisionID}", collectionHandler.RemoveRevisionFromCollection)
	})

	r.Get("/collection/{id}", collectionHandler.ListPublicRevisions)

	return httptest.NewServer(r), mockService
}

func TestCollectionHandler_CreateCollection(t *testing.T) {
	server, mockService := setupCollectionTestServer(t)
	defer server.Close()

	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		collection := &model.Collection{
			ID:    uuid.New(),
			Title: "Test Collection",
		}

		mockService.On("CreateCollection", mock.Anything, userID.String(), "Test Collection").Return(collection, nil).Once()

		body := bytes.NewBufferString(`{"title": "Test Collection"}`)
		req, _ := http.NewRequest("POST", server.URL+"/self/collections", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response model.Collection
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, collection.ID, response.ID)
		assert.Equal(t, collection.Title, response.Title)

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Input", func(t *testing.T) {
		body := bytes.NewBufferString(`{"title": ""}`)
		req, _ := http.NewRequest("POST", server.URL+"/self/collections", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userID.String())

		mockService.On("CreateCollection", mock.Anything, userID.String(), "").Return(nil, model.ErrInvalidInput).Once()

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		body := bytes.NewBufferString(`{"title": "Test Collection"}`)
		req, _ := http.NewRequest("POST", server.URL+"/self/collections", body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestCollectionHandler_GetCollection(t *testing.T) {
	server, mockService := setupCollectionTestServer(t)
	defer server.Close()

	userID := uuid.New()
	collectionID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		collection := &model.Collection{
			ID:    collectionID,
			Title: "Test Collection",
		}

		mockService.On("FindByID", mock.Anything, collectionID.String()).Return(collection, nil).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/collections/"+collectionID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response model.Collection
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, collection.ID, response.ID)
		assert.Equal(t, collection.Title, response.Title)

		mockService.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockService.On("FindByID", mock.Anything, collectionID.String()).Return(nil, model.ErrCollectionNotFound).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/collections/"+collectionID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/self/collections/invalid-uuid", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestCollectionHandler_GetUserCollections(t *testing.T) {
	server, mockService := setupCollectionTestServer(t)
	defer server.Close()

	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		collections := []model.Collection{
			{ID: uuid.New(), Title: "Collection 1"},
			{ID: uuid.New(), Title: "Collection 2"},
		}

		mockService.On("FindByUserID", mock.Anything, userID.String()).Return(collections, nil).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/collections", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response []model.Collection
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 2)
		assert.Equal(t, collections[0].ID, response[0].ID)
		assert.Equal(t, collections[1].ID, response[1].ID)

		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/self/collections", nil)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestCollectionHandler_UpdateCollection(t *testing.T) {
	server, mockService := setupCollectionTestServer(t)
	defer server.Close()

	userID := uuid.New()
	collectionID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		updatedCollection := &model.Collection{
			ID:    collectionID,
			Title: "Updated Collection",
		}

		mockService.On("UpdateCollection", mock.Anything, mock.AnythingOfType("*model.Collection")).Return(nil).Once()

		body, _ := json.Marshal(updatedCollection)
		req, _ := http.NewRequest("PUT", server.URL+"/self/collections/"+collectionID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response model.Collection
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, updatedCollection.ID, response.ID)
		assert.Equal(t, updatedCollection.Title, response.Title)

		mockService.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		updatedCollection := &model.Collection{
			ID:    collectionID,
			Title: "Updated Collection",
		}

		mockService.On("UpdateCollection", mock.Anything, mock.AnythingOfType("*model.Collection")).Return(model.ErrCollectionNotFound).Once()

		body, _ := json.Marshal(updatedCollection)
		req, _ := http.NewRequest("PUT", server.URL+"/self/collections/"+collectionID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", server.URL+"/self/collections/invalid-uuid", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestCollectionHandler_DeleteCollection(t *testing.T) {
	server, mockService := setupCollectionTestServer(t)
	defer server.Close()

	userID := uuid.New()
	collectionID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockService.On("DeleteCollection", mock.Anything, collectionID.String()).Return(nil).Once()

		req, _ := http.NewRequest("DELETE", server.URL+"/self/collections/"+collectionID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockService.On("DeleteCollection", mock.Anything, collectionID.String()).Return(model.ErrCollectionNotFound).Once()

		req, _ := http.NewRequest("DELETE", server.URL+"/self/collections/"+collectionID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", server.URL+"/self/collections/invalid-uuid", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestCollectionHandler_AddRevisionToCollection(t *testing.T) {
	server, mockService := setupCollectionTestServer(t)
	defer server.Close()

	userID := uuid.New()
	collectionID := uuid.New()
	revisionID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockService.On("AddRevisionToCollection", mock.Anything, collectionID.String(), revisionID.String()).Return(nil).Once()

		body := bytes.NewBufferString(`{"revisionID": "` + revisionID.String() + `"}`)
		req, _ := http.NewRequest("POST", server.URL+"/self/collections/"+collectionID.String()+"/revisions", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Collection Not Found", func(t *testing.T) {
		mockService.On("AddRevisionToCollection", mock.Anything, collectionID.String(), revisionID.String()).Return(model.ErrCollectionNotFound).Once()

		body := bytes.NewBufferString(`{"revisionID": "` + revisionID.String() + `"}`)
		req, _ := http.NewRequest("POST", server.URL+"/self/collections/"+collectionID.String()+"/revisions", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestCollectionHandler_ListRevisions(t *testing.T) {
	server, mockService := setupCollectionTestServer(t)
	defer server.Close()

	userID := uuid.New()
	collectionID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		revisions := []model.Revision{
			{ID: uuid.New(), Comment: "Revision 1"},
			{ID: uuid.New(), Comment: "Revision 2"},
		}

		mockService.On("GetRevisionsByCollectionID", mock.Anything, collectionID.String()).Return(revisions, nil).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/collections/"+collectionID.String()+"/revisions", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response []model.Revision
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 2)
		assert.Equal(t, revisions[0].ID, response[0].ID)
		assert.Equal(t, revisions[1].ID, response[1].ID)

		mockService.AssertExpectations(t)
	})

	t.Run("Collection Not Found", func(t *testing.T) {
		mockService.On("GetRevisionsByCollectionID", mock.Anything, collectionID.String()).Return(nil, model.ErrCollectionNotFound).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/collections/"+collectionID.String()+"/revisions", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/self/collections/invalid-uuid/revisions", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestCollectionHandler_ListPublicRevisions(t *testing.T) {
	server, mockService := setupCollectionTestServer(t)
	defer server.Close()

	collectionID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		revisions := []model.Revision{
			{ID: uuid.New(), Comment: "Public Revision 1"},
			{ID: uuid.New(), Comment: "Public Revision 2"},
		}

		mockService.On("GetRevisionsByPublicCollectionID", mock.Anything, collectionID.String()).Return(revisions, nil).Once()

		req, _ := http.NewRequest("GET", server.URL+"/collection/"+collectionID.String(), nil)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response []model.PublicRevisionResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 2)
		assert.Equal(t, revisions[0].Comment, response[0].Comment)
		assert.Equal(t, revisions[1].Comment, response[1].Comment)

		mockService.AssertExpectations(t)
	})

	t.Run("Collection Not Found", func(t *testing.T) {
		mockService.On("GetRevisionsByPublicCollectionID", mock.Anything, collectionID.String()).
			Return(nil, model.ErrCollectionNotFound).Once()

		req, _ := http.NewRequest("GET", server.URL+"/collection/"+collectionID.String(), nil)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestCollectionHandler_RemoveRevisionFromCollection(t *testing.T) {
	server, mockService := setupCollectionTestServer(t)
	defer server.Close()

	userID := uuid.New()
	collectionID := uuid.New()
	revisionID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockService.On("RemoveRevisionFromCollection", mock.Anything, collectionID.String(), revisionID.String()).Return(nil).Once()

		req, _ := http.NewRequest("DELETE", server.URL+"/self/collections/"+collectionID.String()+"/revisions/"+revisionID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Revision Not Found in Collection", func(t *testing.T) {
		mockService.On("RemoveRevisionFromCollection", mock.Anything, collectionID.String(), revisionID.String()).Return(model.ErrCollectionNotFound).Once()

		req, _ := http.NewRequest("DELETE", server.URL+"/self/collections/"+collectionID.String()+"/revisions/"+revisionID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", server.URL+"/self/collections/"+collectionID.String()+"/revisions/invalid-uuid", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", server.URL+"/self/collections/"+collectionID.String()+"/revisions/"+revisionID.String(), nil)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
