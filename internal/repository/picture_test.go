package repository_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mirai-box/mirai-box/internal/config"
	"github.com/mirai-box/mirai-box/internal/database"
	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/repository"
)

var db *sqlx.DB

func TestMain(m *testing.M) {
	conn, err := database.NewConnection(&config.Config{
		Database: config.GetDatabaseConfg(),
	})
	if err != nil {
		log.Fatalf("Can't initialize application %v", err)
	}

	db = conn

	// Run tests
	code := m.Run()

	// Teardown
	db.Close()

	// Exit with the result of m.Run()
	os.Exit(code)
}

func TestSavePictureAndRevision(t *testing.T) {
	repo := repository.NewPictureRepository(db)

	// Prepare test data
	picture := &model.Picture{
		ID:               "0e39910c-844b-43bc-a2c7-9f76f352e601",
		Title:            "Default Title",
		ContentType:      "image/jpeg",
		Filename:         "E4rEwwoXIAQ9oFS.jpeg",
		CreatedAt:        time.Now(),
		LatestRevisionID: "68d6fad9-3ade-4e07-ac64-dd48b1dfe2a2",
	}

	revision := &model.Revision{
		ID:        "68d6fad9-3ade-4e07-ac64-dd48b1dfe2a2",
		PictureID: "0e39910c-844b-43bc-a2c7-9f76f352e601",
		Version:   1,
		FilePath:  "path/to/file",
		Comment:   "Initial revision",
		ArtID:     "art-id-1",
		CreatedAt: time.Now(),
	}

	// Clean up before test
	_, err := db.Exec("DELETE FROM pictures WHERE id = $1", picture.ID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM revisions WHERE id = $1", revision.ID)
	require.NoError(t, err)

	// Execute the method to test
	err = repo.SavePictureAndRevision(picture, revision)
	require.NoError(t, err)

	// Verify the picture was saved correctly
	var savedPicture model.Picture
	err = db.Get(&savedPicture, "SELECT id, title, content_type, filename, latest_revision_id, created_at FROM pictures WHERE id = $1", picture.ID)
	require.NoError(t, err)
	assert.Equal(t, picture.ID, savedPicture.ID)
	assert.Equal(t, picture.Title, savedPicture.Title)
	assert.Equal(t, picture.ContentType, savedPicture.ContentType)
	assert.Equal(t, picture.Filename, savedPicture.Filename)
	assert.Equal(t, picture.LatestRevisionID, savedPicture.LatestRevisionID)

	// Verify the revision was saved correctly
	var savedRevision model.Revision
	err = db.Get(&savedRevision, "SELECT id, picture_id, version, file_path, comment, art_id, created_at FROM revisions WHERE id = $1", revision.ID)
	require.NoError(t, err)
	assert.Equal(t, revision.ID, savedRevision.ID)
	assert.Equal(t, revision.PictureID, savedRevision.PictureID)
	assert.Equal(t, revision.Version, savedRevision.Version)
	assert.Equal(t, revision.FilePath, savedRevision.FilePath)
	assert.Equal(t, revision.Comment, savedRevision.Comment)
	assert.Equal(t, revision.ArtID, savedRevision.ArtID)

	_, err = db.Exec("DELETE FROM pictures WHERE id = $1", picture.ID)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM revisions WHERE id = $1", revision.ID)
	require.NoError(t, err)
}
