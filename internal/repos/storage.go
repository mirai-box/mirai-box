package repos

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
)

// DBStorageRepository implements the StorageRepositoryInterface
type dbStorageRepository struct {
	root string
}

// NewStorageRepository creates a new instance of DBStorageRepository
func NewStorageRepository(root string) StorageRepositoryInterface {
	return &dbStorageRepository{
		root: root,
	}
}

// SaveRevision saves a new revision of a file
func (s *dbStorageRepository) SaveRevision(ctx context.Context,
	fileData io.Reader, userID, artProjectID string, version int) (string, os.FileInfo, error) {
	slog.InfoContext(ctx, "Saving revision", "userID", userID, "artProjectID", artProjectID, "version", version)

	// Ensure the directory structure is ready
	filePath := filepath.Join(s.root, userID, artProjectID, "revisions", "v"+strconv.Itoa(version), "file")
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		slog.ErrorContext(ctx, "Failed to create directories", "error", err, "path", filePath)
		return "", nil, err
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create file", "error", err, "path", filePath)
		return "", nil, err
	}
	defer file.Close()

	// Copy data from the provided io.Reader to the file
	if _, err = io.Copy(file, fileData); err != nil {
		slog.ErrorContext(ctx, "Failed to write data to file", "error", err, "path", filePath)
		return "", nil, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get file info", "error", err, "path", filePath)
		return "", nil, err
	}

	slog.InfoContext(ctx, "Revision saved successfully", "path", filePath)
	return filePath, fileInfo, nil
}

// GetRevision retrieves a specific revision of a file
func (s *dbStorageRepository) GetRevision(ctx context.Context, userID, artProjectID string, version int) (*os.File, error) {
	slog.InfoContext(ctx, "Getting revision", "userID", userID, "artProjectID", artProjectID, "version", version)

	filePath := filepath.Join(s.root, userID, artProjectID, "revisions", "v"+strconv.Itoa(version), "file")

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to open file", "error", err, "path", filePath)
		return nil, err
	}

	slog.InfoContext(ctx, "Revision retrieved successfully", "path", filePath)
	return file, nil
}

// Ensure DBStorageRepository implements StorageRepositoryInterface
var _ StorageRepositoryInterface = (*dbStorageRepository)(nil)
