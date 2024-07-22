//go:build integration
// +build integration

package integration_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mirai-box/mirai-box/internal/app"
	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/repo"
)

func TestUserIntegration(t *testing.T) {
	db, conf, cleanup := setupTestEnvironment(t)
	defer cleanup()

	router := app.SetupRoutes(db, conf)
	userRepo := repo.NewUserRepository(db)

	t.Run("CreateUser", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"username": testUsername,
			"password": testPassword,
			"role":     "user",
		}
		jsonBody, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusCreated, resp.Code)

		var userResp model.UserResponse
		err = json.Unmarshal(resp.Body.Bytes(), &userResp)
		require.NoError(t, err)

		assert.NotEmpty(t, userResp.ID)
		assert.Equal(t, testUsername, userResp.Username)
		assert.Equal(t, "user", userResp.Role)

		// Check database using user repo
		dbUser, err := userRepo.FindUserByUsername(context.Background(), "testuser")
		assert.NoError(t, err)
		assert.Equal(t, userResp.ID, dbUser.ID)
		assert.Equal(t, userResp.Username, dbUser.Username)
		assert.Equal(t, userResp.Role, dbUser.Role)

		// Check stash creation
		stash, err := userRepo.GetStashByUserID(context.Background(), dbUser.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, dbUser.ID.String(), stash.UserID.String())
	})

	t.Run("Login", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"username":     "testuser",
			"password":     "testpass",
			"keepSignedIn": true,
		}
		jsonBody, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var userResp model.UserResponse
		err = json.Unmarshal(resp.Body.Bytes(), &userResp)
		require.NoError(t, err)

		assert.NotEmpty(t, userResp.ID)
		assert.Equal(t, "testuser", userResp.Username)

		// Check for session cookie
		cookies := resp.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == model.SessionCookieName {
				sessionCookie = cookie
				break
			}
		}
		assert.NotNil(t, sessionCookie)
		assert.True(t, sessionCookie.MaxAge > 0)
	})

	t.Run("LoginCheck", func(t *testing.T) {
		// First, perform login to get the session cookie
		loginReqBody := map[string]interface{}{
			"username":     "testuser",
			"password":     "testpass",
			"keepSignedIn": true,
		}
		jsonBody, _ := json.Marshal(loginReqBody)
		loginReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
		loginReq.Header.Set("Content-Type", "application/json")
		loginResp := httptest.NewRecorder()
		router.ServeHTTP(loginResp, loginReq)

		cookies := loginResp.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == model.SessionCookieName {
				sessionCookie = cookie
				break
			}
		}
		require.NotNil(t, sessionCookie)

		// Now perform login check
		checkReq := httptest.NewRequest(http.MethodGet, "/login/check", http.NoBody)
		checkReq.AddCookie(sessionCookie)
		checkResp := httptest.NewRecorder()

		router.ServeHTTP(checkResp, checkReq)

		assert.Equal(t, http.StatusOK, checkResp.Code)

		var response map[string]string
		err := json.Unmarshal(checkResp.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "OK", response["status"])
	})

	t.Run("LoginCheckUnauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/login/check", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}
