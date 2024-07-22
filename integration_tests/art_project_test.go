//go:build integration
// +build integration

package integration_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mirai-box/mirai-box/internal/app"
	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/repo"
)

func TestArtProjectIntegration(t *testing.T) {
	db, conf, cleanup := setupTestEnvironment(t)
	defer cleanup()

	router := app.SetupRoutes(db, conf)
	artProjectRepo := repo.NewArtProjectRepository(db)

	testUser := createTestUserRest(t, router, db)
	sessionCookie := loginTestUser(t, router, testUser)

	var createdArtProjectID string

	t.Run("Create ArtProject", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Add title field
		err := writer.WriteField("title", "Test Art Project")
		require.NoError(t, err)

		// Add file field
		file, err := os.Open("data/1.png")
		require.NoError(t, err)
		defer file.Close()

		part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
		require.NoError(t, err)

		_, err = io.Copy(part, file)
		require.NoError(t, err)

		err = writer.Close()
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/self/artprojects", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.AddCookie(sessionCookie)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusCreated, resp.Code)

		var artProjectResp model.ArtProjectResponse
		err = json.Unmarshal(resp.Body.Bytes(), &artProjectResp)
		require.NoError(t, err)

		assert.NotEmpty(t, artProjectResp.ID)
		assert.Equal(t, "Test Art Project", artProjectResp.Title)
		assert.Equal(t, "1.png", artProjectResp.Filename)
		assert.Equal(t, testUser.ID, artProjectResp.UserID)

		createdArtProjectID = artProjectResp.ID.String()

		// Check database
		dbArtProject, err := artProjectRepo.FindArtProjectByID(context.Background(), createdArtProjectID)
		assert.NoError(t, err)
		assert.Equal(t, artProjectResp.ID, dbArtProject.ID)
		assert.Equal(t, artProjectResp.Title, dbArtProject.Title)
		assert.Equal(t, testUser.ID, dbArtProject.UserID)
	})

	t.Run("AddRevision", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Add comment field
		err := writer.WriteField("comment", "Second version")
		require.NoError(t, err)

		// Add file field
		file, err := os.Open("data/2.png")
		require.NoError(t, err)
		defer file.Close()

		part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
		require.NoError(t, err)

		_, err = io.Copy(part, file)
		require.NoError(t, err)

		err = writer.Close()
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/self/artprojects/"+createdArtProjectID+"/revisions", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.AddCookie(sessionCookie)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusCreated, resp.Code)

		var revisionResp model.RevisionResponse
		err = json.Unmarshal(resp.Body.Bytes(), &revisionResp)
		require.NoError(t, err)

		assert.NotEmpty(t, revisionResp.ID)
		assert.Equal(t, "Second version", revisionResp.Comment)
		assert.Equal(t, 2, revisionResp.Version)
		assert.Equal(t, createdArtProjectID, revisionResp.ArtProjectID.String())

		// Check database
		dbRevision, err := artProjectRepo.FindRevisionByID(context.Background(), revisionResp.ID.String())
		assert.NoError(t, err)

		assert.Equal(t, revisionResp.ID, dbRevision.ID)
		assert.Equal(t, revisionResp.Comment, dbRevision.Comment)
		assert.Equal(t, revisionResp.Version, dbRevision.Version)
	})

	t.Run("ListArtProjects", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/self/artprojects", nil)
		req.AddCookie(sessionCookie)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var artProjects []model.ArtProjectResponse
		err := json.Unmarshal(resp.Body.Bytes(), &artProjects)
		require.NoError(t, err)

		assert.NotEmpty(t, artProjects)
		assert.Equal(t, 1, len(artProjects))
		assert.Equal(t, "Test Art Project", artProjects[0].Title)
	})

	t.Run("GetArtProjectByID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/self/artprojects/"+createdArtProjectID, nil)
		req.AddCookie(sessionCookie)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var artProject model.ArtProjectResponse
		err := json.Unmarshal(resp.Body.Bytes(), &artProject)
		require.NoError(t, err)

		assert.Equal(t, createdArtProjectID, artProject.ID.String())
		assert.Equal(t, "Test Art Project", artProject.Title)
	})

	t.Run("ListRevisions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/self/artprojects/"+createdArtProjectID+"/revisions", nil)
		req.AddCookie(sessionCookie)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var revisions []model.RevisionResponse
		err := json.Unmarshal(resp.Body.Bytes(), &revisions)
		require.NoError(t, err)

		assert.Equal(t, 2, len(revisions))
		assert.Equal(t, 1, revisions[0].Version)
		assert.Equal(t, 2, revisions[1].Version)
		assert.Equal(t, "Second version", revisions[1].Comment)
	})

	t.Run("GetArtByID", func(t *testing.T) {
		artProjects, err := artProjectRepo.FindByUserID(context.Background(), testUser.ID.String())
		require.NoError(t, err)
		require.NotEmpty(t, artProjects)

		revisions, err := artProjectRepo.ListAllRevisions(context.Background(), artProjects[0].ID.String())
		require.NoError(t, err)
		require.NotEmpty(t, revisions)
		assert.Equal(t, 2, len(revisions))
		expectedSize := fmt.Sprintf("%d", revisions[0].Size)

		req := httptest.NewRequest(http.MethodGet, "/art/"+revisions[0].ArtID, nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, expectedSize, resp.Header().Get("Content-Length"))
	})
}
