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

func TestWebPageIntegration(t *testing.T) {
	db, conf, cleanup := setupTestEnvironment(t)
	defer cleanup()

	router := app.SetupRoutes(db, conf)
	webPageRepo := repo.NewWebPageRepository(db)

	// Create a test user
	testUser := createTestUser(t, db)
	// Login to get a session cookie
	sessionCookie := loginTestUser(t, router, testUser)

	t.Run("CreateWebPage", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"title": "Test Page",
			"html":  "<h1>This is a Test Page</h1>",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/self/webpages", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusCreated, resp.Code)

		var webPageResp model.WebPageResponse
		err := json.Unmarshal(resp.Body.Bytes(), &webPageResp)
		require.NoError(t, err)

		assert.NotEmpty(t, webPageResp.ID)
		assert.Equal(t, "Test Page", webPageResp.Title)
		assert.Equal(t, "<h1>This is a Test Page</h1>", webPageResp.Html)
		assert.Equal(t, testUser.ID, webPageResp.UserID)

		// Check database
		dbWebPage, err := webPageRepo.FindWebPageByID(context.Background(), webPageResp.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, webPageResp.ID, dbWebPage.ID)
		assert.Equal(t, webPageResp.Title, dbWebPage.Title)
		assert.Equal(t, webPageResp.Html, dbWebPage.Html)
		assert.Equal(t, testUser.ID, dbWebPage.UserID)
	})

	t.Run("ListWebPages", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/self/webpages", nil)
		req.AddCookie(sessionCookie)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var webPages []model.WebPageResponse
		err := json.Unmarshal(resp.Body.Bytes(), &webPages)
		require.NoError(t, err)

		assert.NotEmpty(t, webPages)
		assert.Equal(t, 1, len(webPages))
		assert.Equal(t, "Test Page", webPages[0].Title)
	})

	t.Run("GetWebPageByID", func(t *testing.T) {
		webPages, err := webPageRepo.FindWebPagesByUserID(context.Background(), testUser.ID.String())
		require.NoError(t, err)
		require.NotEmpty(t, webPages)

		req := httptest.NewRequest(http.MethodGet, "/self/webpages/"+webPages[0].ID.String(), nil)
		req.AddCookie(sessionCookie)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var webPage model.WebPageResponse
		err = json.Unmarshal(resp.Body.Bytes(), &webPage)
		require.NoError(t, err)

		assert.Equal(t, webPages[0].ID, webPage.ID)
		assert.Equal(t, "Test Page", webPage.Title)
	})

	t.Run("UpdateWebPage", func(t *testing.T) {
		webPages, err := webPageRepo.FindWebPagesByUserID(context.Background(), testUser.ID.String())
		require.NoError(t, err)
		require.NotEmpty(t, webPages)

		updateReqBody := map[string]interface{}{
			"page_type": "test",
		}
		jsonBody, _ := json.Marshal(updateReqBody)

		req := httptest.NewRequest(http.MethodPut, "/self/webpages/"+webPages[0].ID.String(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var updatedWebPage model.WebPageResponse
		err = json.Unmarshal(resp.Body.Bytes(), &updatedWebPage)
		require.NoError(t, err)

		assert.Equal(t, webPages[0].ID, updatedWebPage.ID)
		assert.Equal(t, "test", updatedWebPage.PageType)
		assert.Equal(t, "Test Page", updatedWebPage.Title) // Ensure other fields are not changed
	})
}
