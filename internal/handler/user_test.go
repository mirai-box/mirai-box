package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/mirai-box/mirai-box/internal/handler"
	"github.com/mirai-box/mirai-box/internal/middleware"
	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupUserTestServer(t *testing.T) (*httptest.Server, *mocks.UserService) {
	r := chi.NewRouter()

	cookieStore := sessions.NewCookieStore([]byte("abc"))
	userMock := mocks.NewUserService(t)
	m := middleware.NewMiddleware(cookieStore, userMock)
	userHandler := handler.NewUserHandler(userMock, cookieStore)

	r.Post("/login", userHandler.Login)
	r.Get("/login/check", userHandler.LoginCheck)

	r.Route("/self", func(r chi.Router) {
		r.Use(m.MockAuthMiddleware)
		r.Get("/stash", userHandler.MyStash)
	})

	r.Route("/api", func(r chi.Router) {
		r.Post("/users", userHandler.CreateUser)
		r.Route("/users/{id}", func(r chi.Router) {
			r.Use(m.MockAuthMiddleware)
			r.Get("/", userHandler.GetUser)
			r.Put("/", userHandler.UpdateUser)
			r.Delete("/", userHandler.DeleteUser)
		})
	})

	return httptest.NewServer(r), userMock
}

func TestUserHandler_CreateUser(t *testing.T) {
	server, mockService := setupUserTestServer(t)
	defer server.Close()

	t.Run("Success", func(t *testing.T) {
		newUser := &model.User{
			ID:       uuid.New(),
			Username: "testuser",
			Role:     "user",
		}

		mockService.On("CreateUser", mock.Anything, "testuser", "password123", "user").Return(newUser, nil).Once()

		body := bytes.NewBufferString(`{"username":"testuser","password":"password123","role":"user"}`)
		req, _ := http.NewRequest("POST", server.URL+"/api/users", body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response model.UserResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, newUser.ID, response.ID)
		assert.Equal(t, newUser.Username, response.Username)

		mockService.AssertExpectations(t)
	})

	t.Run("Service Error", func(t *testing.T) {
		mockService.On("CreateUser", mock.Anything, "testuser", "password123", "user").
			Return(nil, fmt.Errorf("some error")).Once()

		body := bytes.NewBufferString(`{"username":"testuser","password":"password123","role":"user"}`)
		req, _ := http.NewRequest("POST", server.URL+"/api/users", body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid JSON string", func(t *testing.T) {
		body := bytes.NewBufferString(`{"username":"testuser","password":"password123"`)
		req, _ := http.NewRequest("POST", server.URL+"/api/users", body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid input data", func(t *testing.T) {
		body := bytes.NewBufferString(`{"username":"testuser","password":"","role":"user"}`)
		req, _ := http.NewRequest("POST", server.URL+"/api/users", body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var response model.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, response.Status)
		assert.Contains(t, response.Message, `Field validation for 'Password'`)

		mockService.AssertExpectations(t)
	})

}

func TestUserHandler_Login(t *testing.T) {
	server, mockService := setupUserTestServer(t)
	defer server.Close()

	t.Run("Success", func(t *testing.T) {
		user := &model.User{
			ID:       uuid.New(),
			Username: "testuser",
			Role:     "user",
		}

		mockService.On("Authenticate", mock.Anything, "testuser", "password123").Return(user, nil).Once()

		body := bytes.NewBufferString(`{"username":"testuser","password":"password123","keepSignedIn":false}`)
		req, _ := http.NewRequest("POST", server.URL+"/login", body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response model.UserResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, user.ID, response.ID)
		assert.Equal(t, user.Username, response.Username)

		mockService.AssertExpectations(t)
	})

	t.Run("Invalid  JSON", func(t *testing.T) {
		body := bytes.NewBufferString(`{"username":"testuser","password":"password123"`)
		req, _ := http.NewRequest("POST", server.URL+"/login", body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_MyStash(t *testing.T) {
	server, mockService := setupUserTestServer(t)
	defer server.Close()

	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		stash := &model.Stash{
			ID:     uuid.New(),
			UserID: userID,
		}

		mockService.On("GetStashByUserID", mock.Anything, userID.String()).Return(stash, nil).Once()

		req, _ := http.NewRequest("GET", server.URL+"/self/stash", nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response model.StashResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, stash.ID, response.ID)
		assert.Equal(t, stash.UserID, response.UserID)

		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/self/stash", nil)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestUserHandler_GetUser(t *testing.T) {
	server, mockService := setupUserTestServer(t)
	defer server.Close()

	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		user := &model.User{
			ID:       userID,
			Username: "testuser",
			Role:     "user",
		}

		mockService.On("GetUser", mock.Anything, userID.String()).Return(user, nil).Once()

		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/users/%s", server.URL, userID.String()), nil)
		req.Header.Set("X-Admin-ID", uuid.NewString())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response model.UserResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, user.ID, response.ID)
		assert.Equal(t, user.Username, response.Username)

		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/users/%s", server.URL, userID.String()), nil)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockService.On("GetUser", mock.Anything, userID.String()).Return(nil, model.ErrUserNotFound).Once()

		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/users/%s", server.URL, userID.String()), nil)
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_UpdateUser(t *testing.T) {
	server, mockService := setupUserTestServer(t)
	defer server.Close()

	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		updatedUser := &model.User{
			ID:       userID,
			Username: "updateduser",
			Role:     "admin",
		}

		mockService.On("UpdateUser", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil).Once()

		body := bytes.NewBufferString(`{"username":"updateduser","password":"newpassword","role":"admin"}`)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/api/users/%s", server.URL, userID.String()), body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Admin-ID", uuid.NewString())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response model.UserResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, updatedUser.ID, response.ID)
		assert.Equal(t, updatedUser.Username, response.Username)
		assert.Equal(t, updatedUser.Role, response.Role)

		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		body := bytes.NewBufferString(`{"username":"updateduser","password":"newpassword","role":"admin"}`)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/api/users/%s", server.URL, userID.String()), body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		body := bytes.NewBufferString(`{"username":"updateduser","password":"newpassword"`)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/api/users/%s", server.URL, userID.String()), body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestUserHandler_DeleteUser(t *testing.T) {
	server, mockService := setupUserTestServer(t)
	defer server.Close()

	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockService.On("DeleteUser", mock.Anything, userID.String()).Return(nil).Once()

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/users/%s", server.URL, userID.String()), nil)
		req.Header.Set("X-Admin-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		mockService.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/users/%s", server.URL, userID.String()), nil)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockService.On("DeleteUser", mock.Anything, userID.String()).Return(model.ErrUserNotFound).Once()

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/users/%s", server.URL, userID.String()), nil)
		req.Header.Set("X-Admin-ID", userID.String())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockService.AssertExpectations(t)
	})
}
