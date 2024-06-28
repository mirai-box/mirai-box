package repository

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type dbStorageRepository struct {
	root string
}

func NewStorageRepository(root string) StorageRepository {
	return &dbStorageRepository{
		root: root,
	}
}

func (s *dbStorageRepository) SavePicture(fileData io.Reader, filePath string) error {
	// Ensure the directory structure is ready
	revisionPath := filepath.Join(s.root, filePath)
	if err := os.MkdirAll(filepath.Dir(revisionPath), os.ModePerm); err != nil {
		slog.Error("Failed to create directories", "error", err)
		return err
	}

	// Create the file
	file, err := os.Create(revisionPath)
	if err != nil {
		slog.Error("Failed to create file", "error", err)
		return err
	}
	defer file.Close()

	// Copy data from the provided io.Reader to the file
	if _, err = io.Copy(file, fileData); err != nil {
		slog.Error("Failed to write data to file", "error", err)
		return err
	}

	return nil
}

func (s *dbStorageRepository) GetPicture(filePath string) (*os.File, error) {
	revisionPath := filepath.Join(s.root, filePath)

	// Open the file
	file, err := os.Open(revisionPath)
	if err != nil {
		return nil, err
	}

	return file, nil
}
