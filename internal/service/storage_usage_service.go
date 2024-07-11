package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/repos"
)

// StorageUsageService implements the StorageUsageServiceInterface
type storageUsageService struct {
	repo repos.StorageUsageRepositoryInterface
}

// NewStorageUsageService creates a new StorageUsageService
func NewStorageUsageService(repo repos.StorageUsageRepositoryInterface) StorageUsageServiceInterface {
	return &storageUsageService{repo: repo}
}

// CreateStorageUsage creates a new storage usage record for a user
func (s *storageUsageService) CreateStorageUsage(ctx context.Context, userID string, quota int64) (*models.StorageUsage, error) {
	slog.InfoContext(ctx, "Creating new storage usage", "userID", userID, "quota", quota)

	storageUsage := &models.StorageUsage{
		UserID:    uuid.MustParse(userID),
		UsedSpace: 0,
		Quota:     quota,
	}

	if err := s.repo.Create(ctx, storageUsage); err != nil {
		slog.ErrorContext(ctx, "Failed to create storage usage", "error", err, "userID", userID)
		return nil, err
	}

	slog.InfoContext(ctx, "Storage usage created successfully", "userID", userID)
	return storageUsage, nil
}

// FindByUserID retrieves the storage usage for a specific user
func (s *storageUsageService) FindByUserID(ctx context.Context, userID string) (*models.StorageUsage, error) {
	slog.InfoContext(ctx, "Finding storage usage by user ID", "userID", userID)

	storageUsage, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find storage usage", "error", err, "userID", userID)
		return nil, err
	}

	slog.InfoContext(ctx, "Storage usage found successfully", "userID", userID)
	return storageUsage, nil
}

// UpdateStorageUsage updates the used space for a user's storage usage
func (s *storageUsageService) UpdateStorageUsage(ctx context.Context, userID string, usedSpace int64) (*models.StorageUsage, error) {
	slog.InfoContext(ctx, "Updating storage usage", "userID", userID, "usedSpace", usedSpace)

	storageUsage, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find storage usage for update", "error", err, "userID", userID)
		return nil, err
	}

	storageUsage.UsedSpace = usedSpace

	if err := s.repo.Update(ctx, storageUsage); err != nil {
		slog.ErrorContext(ctx, "Failed to update storage usage", "error", err, "userID", userID)
		return nil, err
	}

	slog.InfoContext(ctx, "Storage usage updated successfully", "userID", userID, "usedSpace", usedSpace)
	return storageUsage, nil
}

// Ensure StorageUsageService implements StorageUsageServiceInterface
var _ StorageUsageServiceInterface = (*storageUsageService)(nil)
