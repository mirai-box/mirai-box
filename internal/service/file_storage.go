package service

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/repo"
)

//go:generate go run github.com/vektra/mockery/v2@v2 --name=FileStorageService --filename=file_storage_Service.go --output=../../mocks/
type FileStorageService interface {
	SaveRevisionFile(ctx context.Context, fileData io.Reader, userID, artProjectID string, version int) (string, os.FileInfo, error)
	GetRevisionFile(ctx context.Context, userID, artProjectID string, version int) (io.ReadCloser, error)
	FindStashByUserID(ctx context.Context, userID string) (*model.Stash, error)
}

type fileStorageService struct {
	fileStorageRepo repo.FileStorageRepository
}

func NewFileStorageService(fileStorageRepo repo.FileStorageRepository) FileStorageService {
	return &fileStorageService{
		fileStorageRepo: fileStorageRepo,
	}
}

func (s *fileStorageService) SaveRevisionFile(ctx context.Context, fileData io.Reader, userID, artProjectID string, version int) (string, os.FileInfo, error) {
	logger := slog.With("method", "SaveRevisionFile", "userID", userID, "artProjectID", artProjectID, "version", version)

	filePath, fileInfo, err := s.fileStorageRepo.SaveRevisionFile(ctx, fileData, userID, artProjectID, version)
	if err != nil {
		logger.Error("Failed to save revision file", "error", err)
		return "", nil, err
	}

	logger.Info("Revision file saved successfully", "filePath", filePath)
	return filePath, fileInfo, nil
}

func (s *fileStorageService) GetRevisionFile(ctx context.Context, userID, artProjectID string, version int) (io.ReadCloser, error) {
	logger := slog.With("method", "GetRevisionFile", "userID", userID, "artProjectID", artProjectID, "version", version)

	file, err := s.fileStorageRepo.GetRevisionFile(ctx, userID, artProjectID, version)
	if err != nil {
		logger.Error("Failed to get revision file", "error", err)
		return nil, err
	}

	logger.Info("Revision file retrieved successfully")
	return file, nil
}

func (s *fileStorageService) FindStashByUserID(ctx context.Context, userID string) (*model.Stash, error) {
	logger := slog.With("method", "FindStashByUserID", "userID", userID)

	stash, err := s.fileStorageRepo.FindStashByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to find stash", "error", err)
		return nil, err
	}

	logger.Info("Stash found successfully")
	return stash, nil
}
