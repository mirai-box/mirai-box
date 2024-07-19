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

func setupWebPageTestServer(t *testing.T) (*httptest.Server, *mocks.WebPageService) {
	r := chi.NewRouter()

	mockService := mocks.NewWebPageService(t)
	cookieStore := sessions.NewCookieStore([]byte("abc"))
	userMock := mocks.NewUserService(t)
	m := middleware.NewMiddleware(cookieStore, userMock)

	webPageHandler := handler.NewWebPageHandler(mockService)

	r.Route("/self", func(r chi.Router) {
		r.Use(m.MockAuthMiddleware)

		r.Post("/webpages", webPageHandler.CreateWebPage)
		r.Get("/webpages", webPageHandler.MyWebPages)
		r.With(middleware.ValidateUUID("id")).Get("/webpages/{id}", webPageHandler.MyWebPageByID)
		r.With(middleware.ValidateUUID("id")).Put("/webpages/{id}", webPageHandler.UpdateWebPage)
		r.With(middleware.ValidateUUID("id")).Delete("/webpages/{id}", webPageHandler.DeleteWebPage)
	})

	r.With(middleware.ValidateUUID("id")).Get("/webpages/{id}", webPageHandler.GetWebPage)
	r.Get("/webpages", webPageHandler.ListWebPages)
	r.With(middleware.ValidateUUID("userId")).Get("/users/{userId}/webpages", webPageHandler.ListUserWebPages)

	return httptest.NewServer(r), mockService
}

func TestWebPageHandler_CreateWebPage(t *testing.T) {
	server, mockService := setupWebPageTestServer(t)
	defer server.Close()

	t.Run("Success", func(t *testing.T) {
		webPage := &model.WebPage{
			ID:       uuid.New(),
			UserID:   uuid.New(),
			Title:    "Test Page",
			Html:     "<h1>Test</h1>",
			PageType: "main",
			Public:   true,
		}

		mockService.On("CreateWebPage", mock.Anything, mock.AnythingOfType("*model.WebPage")).
			Return(webPage, nil)

		body, _ := json.Marshal(webPage)
		req, _ := http.NewRequest("POST", server.URL+"/self/webpages", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", webPage.UserID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response model.WebPageResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, webPage.ID, response.ID)
		assert.Equal(t, webPage.Title, response.Title)
		mockService.AssertExpectations(t)
	})

	t.Run("Status Unauthorized", func(t *testing.T) {
		webPage := &model.WebPage{
			ID:       uuid.New(),
			UserID:   uuid.New(),
			Title:    "Test Page",
			Html:     "<h1>Test</h1>",
			PageType: "main",
			Public:   true,
		}

		body, _ := json.Marshal(webPage)
		req, _ := http.NewRequest("POST", server.URL+"/self/webpages", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestWebPageHandler_GetWebPage(t *testing.T) {
	server, mockService := setupWebPageTestServer(t)
	defer server.Close()

	webPage := &model.WebPage{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		Title:    "Test Page",
		Html:     "<h1>Test</h1>",
		PageType: "main",
		Public:   true,
	}

	t.Run("Success", func(t *testing.T) {
		mockService.On("GetWebPage", mock.Anything, webPage.ID.String()).Return(webPage, nil).Once()

		resp, err := http.Get(server.URL + "/webpages/" + webPage.ID.String())
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response model.WebPageResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, webPage.ID, response.ID)
		assert.Equal(t, webPage.Title, response.Title)

		mockService.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		nonExistentID := uuid.New()
		mockService.On("GetWebPage", mock.Anything, nonExistentID.String()).
			Return(nil, model.ErrWebPageNotFound).Once()

		resp, err := http.Get(server.URL + "/webpages/" + nonExistentID.String())
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/webpages/invalid-uuid")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestWebPageHandler_UpdateWebPage(t *testing.T) {
	server, mockService := setupWebPageTestServer(t)
	defer server.Close()

	webPage := &model.WebPage{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		Title:    "Updated Test Page",
		Html:     "<h1>Updated Test</h1>",
		PageType: "main",
		Public:   true,
	}

	t.Run("Success", func(t *testing.T) {
		mockService.On("GetWebPage", mock.Anything, webPage.ID.String()).Return(webPage, nil)
		mockService.On("UpdateWebPage", mock.Anything, mock.AnythingOfType("*model.WebPage")).Return(webPage, nil)

		body, _ := json.Marshal(webPage)
		req, _ := http.NewRequest("PUT", server.URL+"/self/webpages/"+webPage.ID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", webPage.UserID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response model.WebPageResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, webPage.ID, response.ID)
		assert.Equal(t, webPage.Title, response.Title)

		mockService.AssertExpectations(t)
	})
}

func TestWebPageHandler_CreateWebPage_Unauthorized(t *testing.T) {
	server, mockService := setupWebPageTestServer(t)
	defer server.Close()

	webPage := &model.WebPage{
		Title: "Test Page",
		Html:  "<h1>Test</h1>",
	}

	body, _ := json.Marshal(webPage)
	req, _ := http.NewRequest("POST", server.URL+"/self/webpages", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	// Not setting X-User-ID header to simulate unauthorized access

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	mockService.AssertNotCalled(t, "CreateWebPage")
}

func TestWebPageHandler_DeleteWebPage(t *testing.T) {
	server, mockService := setupWebPageTestServer(t)
	defer server.Close()

	webPageID := uuid.New()
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockService.On("GetWebPage", mock.Anything, webPageID.String()).
			Return(&model.WebPage{ID: webPageID, UserID: userID}, nil).Once()
		mockService.On("DeleteWebPage", mock.Anything, webPageID.String()).Return(nil).Once()

		req, _ := http.NewRequest("DELETE", server.URL+"/self/webpages/"+webPageID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockService.On("GetWebPage", mock.Anything, webPageID.String()).
			Return(nil, model.ErrWebPageNotFound).Once()

		req, _ := http.NewRequest("DELETE", server.URL+"/self/webpages/"+webPageID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		otherUserID := uuid.New()
		mockService.On("GetWebPage", mock.Anything, webPageID.String()).
			Return(&model.WebPage{ID: webPageID, UserID: otherUserID}, nil).Once()

		req, _ := http.NewRequest("DELETE", server.URL+"/self/webpages/"+webPageID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestWebPageHandler_ListWebPages(t *testing.T) {
	server, mockService := setupWebPageTestServer(t)
	defer server.Close()

	t.Run("Success", func(t *testing.T) {
		webPages := []model.WebPage{
			{ID: uuid.New(), Title: "Page 1"},
			{ID: uuid.New(), Title: "Page 2"},
		}

		mockService.On("ListWebPages", mock.Anything).Return(webPages, nil).Once()

		resp, err := http.Get(server.URL + "/webpages")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response []model.WebPageResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 2)
		mockService.AssertExpectations(t)
	})

	t.Run("No Pages Found", func(t *testing.T) {
		mockService.On("ListWebPages", mock.Anything).Return(nil, model.ErrWebPageNotFound).Once()

		resp, err := http.Get(server.URL + "/webpages")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestWebPageHandler_ListUserWebPages(t *testing.T) {
	server, mockService := setupWebPageTestServer(t)
	defer server.Close()

	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		webPages := []model.WebPage{
			{ID: uuid.New(), UserID: userID, Title: "Page 1"},
			{ID: uuid.New(), UserID: userID, Title: "Page 2"},
		}

		mockService.On("ListUserWebPages", mock.Anything, userID.String()).Return(webPages, nil).Once()

		resp, err := http.Get(server.URL + "/users/" + userID.String() + "/webpages")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response []model.WebPageResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 2)
		mockService.AssertExpectations(t)
	})

	t.Run("No Pages Found", func(t *testing.T) {
		mockService.On("ListUserWebPages", mock.Anything, userID.String()).Return(nil, model.ErrWebPageNotFound).Once()

		resp, err := http.Get(server.URL + "/users/" + userID.String() + "/webpages")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid User ID", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/users/invalid-uuid/webpages")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestWebPageHandler_MyWebPages(t *testing.T) {
	server, mockService := setupWebPageTestServer(t)
	defer server.Close()

	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		webPages := []model.WebPage{
			{ID: uuid.New(), UserID: userID, Title: "Page 1"},
			{ID: uuid.New(), UserID: userID, Title: "Page 2"},
		}

		mockService.On("ListUserWebPages", mock.Anything, userID.String()).Return(webPages, nil).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/webpages", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response []model.WebPageResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 2)
		mockService.AssertExpectations(t)
	})

	t.Run("No Pages Found", func(t *testing.T) {
		mockService.On("ListUserWebPages", mock.Anything, userID.String()).Return(nil, model.ErrWebPageNotFound).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/webpages", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/self/webpages", nil)
		// Not setting X-User-ID header

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestWebPageHandler_MyWebPageByID(t *testing.T) {
	server, mockService := setupWebPageTestServer(t)
	defer server.Close()

	userID := uuid.New()
	webPageID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		webPage := &model.WebPage{ID: webPageID, UserID: userID, Title: "My Page"}

		mockService.On("GetWebPage", mock.Anything, webPageID.String()).Return(webPage, nil).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/webpages/"+webPageID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response model.WebPageResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, webPage.ID, response.ID)
		assert.Equal(t, webPage.Title, response.Title)
		mockService.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockService.On("GetWebPage", mock.Anything, webPageID.String()).Return(nil, model.ErrWebPageNotFound).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/webpages/"+webPageID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized - Not Owner", func(t *testing.T) {
		otherUserID := uuid.New()
		webPage := &model.WebPage{ID: webPageID, UserID: otherUserID, Title: "Other User's Page"}

		mockService.On("GetWebPage", mock.Anything, webPageID.String()).Return(webPage, nil).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/webpages/"+webPageID.String(), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/self/webpages/invalid-uuid", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
