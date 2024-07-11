package repos

import (
	"context"
	"log/slog"

	"github.com/mirai-box/mirai-box/internal/models"

	"gorm.io/gorm"
)

// StashRepository implements the StashRepositoryInterface
type StashRepository struct {
	DB *gorm.DB
}

// NewStashRepository creates a new instance of StashRepository
func NewStashRepository(db *gorm.DB) StashRepositoryInterface {
	return &StashRepository{DB: db}
}

// Create adds a new stash to the database
func (r *StashRepository) Create(ctx context.Context, stash *models.Stash) error {
	slog.InfoContext(ctx, "Creating new stash", "stashID", stash.ID, "userID", stash.UserID)
	if err := r.DB.Create(stash).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to create stash", "error", err, "stashID", stash.ID)
		return err
	}
	slog.InfoContext(ctx, "Stash created successfully", "stashID", stash.ID)
	return nil
}

// FindByID retrieves a stash by its ID
func (r *StashRepository) FindByID(ctx context.Context, id string) (*models.Stash, error) {
	slog.InfoContext(ctx, "Finding stash by ID", "stashID", id)

	var stash models.Stash
	err := r.DB.Preload("User").First(&stash, "id = ?", id).Error
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find stash by ID", "error", err, "stashID", id)
		return nil, err
	}

	slog.InfoContext(ctx, "Stash found successfully", "stashID", id)
	return &stash, nil
}

// FindByUserID retrieves a stash by user ID
func (r *StashRepository) FindByUserID(ctx context.Context, userID string) (*models.Stash, error) {
	slog.InfoContext(ctx, "StashRepository: Finding stash by userID", "userID", userID)

	var stash models.Stash
	err := r.DB.Preload("User").First(&stash, "user_id = ?", userID).Error
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find stash by user ID", "error", err, "userID", userID)
		return nil, err
	}

	slog.InfoContext(ctx, "Stash found successfully", "userID", userID, "stashID", stash.ID)
	return &stash, nil
}

// Update modifies an existing stash in the database
func (r *StashRepository) Update(ctx context.Context, stash *models.Stash) error {
	slog.InfoContext(ctx, "Updating stash", "stashID", stash.ID)
	if err := r.DB.Save(stash).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to update stash", "error", err, "stashID", stash.ID)
		return err
	}
	slog.InfoContext(ctx, "Stash updated successfully", "stashID", stash.ID)
	return nil
}

// Delete removes a stash from the database
func (r *StashRepository) Delete(ctx context.Context, id string) error {
	slog.InfoContext(ctx, "Deleting stash", "stashID", id)
	if err := r.DB.Delete(&models.Stash{}, "id = ?", id).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to delete stash", "error", err, "stashID", id)
		return err
	}
	slog.InfoContext(ctx, "Stash deleted successfully", "stashID", id)
	return nil
}

// Ensure StashRepository implements StashRepositoryInterface
var _ StashRepositoryInterface = (*StashRepository)(nil)
