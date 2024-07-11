package repos

import (
	"context"
	"log/slog"

	"github.com/mirai-box/mirai-box/internal/models"

	"gorm.io/gorm"
)

// StorageUsageRepository implements the StorageUsageRepositoryInterface
type storageUsageRepository struct {
	DB *gorm.DB
}

// NewStorageUsageRepository creates a new instance of StorageUsageRepository
func NewStorageUsageRepository(db *gorm.DB) StorageUsageRepositoryInterface {
	return &storageUsageRepository{DB: db}
}

// Create adds a new storage usage record to the database
func (r *storageUsageRepository) Create(ctx context.Context, storageUsage *models.StorageUsage) error {
	slog.InfoContext(ctx, "Creating new storage usage record", "userID", storageUsage.UserID)

	if err := r.DB.Create(storageUsage).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to create storage usage record", "error", err, "userID", storageUsage.UserID)
		return err
	}

	slog.InfoContext(ctx, "Storage usage record created successfully", "userID", storageUsage.UserID)
	return nil
}

// FindByUserID retrieves a storage usage record by user ID
func (r *storageUsageRepository) FindByUserID(ctx context.Context, userID string) (*models.StorageUsage, error) {
	slog.InfoContext(ctx, "Finding storage usage record by user ID", "userID", userID)

	var storageUsage models.StorageUsage
	if err := r.DB.First(&storageUsage, "user_id = ?", userID).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to find storage usage record", "error", err, "userID", userID)
		return nil, err
	}

	slog.InfoContext(ctx, "Storage usage record found", "userID", userID)
	return &storageUsage, nil
}

// Update modifies an existing storage usage record in the database
func (r *storageUsageRepository) Update(ctx context.Context, storageUsage *models.StorageUsage) error {
	slog.InfoContext(ctx, "Updating storage usage record", "userID", storageUsage.UserID)

	if err := r.DB.Save(storageUsage).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to update storage usage record", "error", err, "userID", storageUsage.UserID)
		return err
	}

	slog.InfoContext(ctx, "Storage usage record updated successfully", "userID", storageUsage.UserID)
	return nil
}

// Delete removes a storage usage record from the database
func (r *storageUsageRepository) Delete(ctx context.Context, userID string) error {
	slog.InfoContext(ctx, "Deleting storage usage record", "userID", userID)

	if err := r.DB.Delete(&models.StorageUsage{}, "user_id = ?", userID).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to delete storage usage record", "error", err, "userID", userID)
		return err
	}

	slog.InfoContext(ctx, "Storage usage record deleted successfully", "userID", userID)
	return nil
}

// Ensure StorageUsageRepository implements StorageUsageRepositoryInterface
var _ StorageUsageRepositoryInterface = (*storageUsageRepository)(nil)
