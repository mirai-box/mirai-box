package repo

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"gorm.io/gorm"

	"github.com/mirai-box/mirai-box/internal/model"
)

// FileStorageRepository defines the interface for file storage related operations.
type FileStorageRepository interface {
	SaveRevisionFile(ctx context.Context, fileData io.Reader, userID, artProjectID string, version int) (string, os.FileInfo, error)
	GetRevisionFile(ctx context.Context, userID, artProjectID string, version int) (io.ReadCloser, error)
	FindStashByUserID(ctx context.Context, userID string) (*model.Stash, error)
}

type fileStorageRepo struct {
	db   *gorm.DB
	root string
}

// NewFileStorageRepository creates a new instance of FileStorageRepository.
func NewFileStorageRepository(db *gorm.DB, root string) FileStorageRepository {
	return &fileStorageRepo{db: db, root: root}
}

// SaveRevisionFile saves a new revision of a file.
func (r *fileStorageRepo) SaveRevisionFile(ctx context.Context, fileData io.Reader, userID, artProjectID string, version int) (string, os.FileInfo, error) {
	logger := slog.With("method", "SaveRevisionFile", "userID", userID, "artProjectID", artProjectID, "version", version)

	filePath := filepath.Join(r.root, userID, artProjectID, "revisions", "v"+strconv.Itoa(version))
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		logger.Error("Failed to create directories", "error", err)
		return "", nil, err
	}

	file, err := os.Create(filePath)
	if err != nil {
		logger.Error("Failed to create file", "error", err)
		return "", nil, err
	}
	defer file.Close()

	if _, err = io.Copy(file, fileData); err != nil {
		logger.Error("Failed to write data to file", "error", err)
		return "", nil, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		logger.Error("Failed to get file info", "error", err)
		return "", nil, err
	}

	logger.Info("Revision file saved successfully")
	return filePath, fileInfo, nil
}

// GetRevisionFile retrieves a specific revision of a file.
func (r *fileStorageRepo) GetRevisionFile(ctx context.Context, userID, artProjectID string, version int) (io.ReadCloser, error) {
	logger := slog.With("method", "GetRevisionFile", "userID", userID, "artProjectID", artProjectID, "version", version)

	filePath := filepath.Join(r.root, userID, artProjectID, "revisions", "v"+strconv.Itoa(version))
	file, err := os.Open(filePath)
	if err != nil {
		logger.Error("Failed to open file", "error", err)
		return nil, err
	}

	logger.Info("Revision file retrieved successfully")
	return file, nil
}

// FindStashByUserID retrieves the stash for a specific user.
func (r *fileStorageRepo) FindStashByUserID(ctx context.Context, userID string) (*model.Stash, error) {
	logger := slog.With("method", "FindStashByUserID", "userID", userID)

	var stash model.Stash
	if err := r.db.Where("user_id = ?", userID).First(&stash).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Info("Stash not found")
			return nil, model.ErrStashNotFound
		}

		logger.Error("Failed to find stash", "error", err)
		return nil, err
	}

	logger.Info("Stash found successfully")
	return &stash, nil
}
