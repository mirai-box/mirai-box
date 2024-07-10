package repos

import (
	"context"
	"log/slog"

	"github.com/mirai-box/mirai-box/internal/models"

	"gorm.io/gorm"
)

// CollectionRepository implements the CollectionRepositoryInterface
type CollectionRepository struct {
	DB *gorm.DB
}

// NewCollectionRepository creates a new instance of CollectionRepository
func NewCollectionRepository(db *gorm.DB) CollectionRepositoryInterface {
	return &CollectionRepository{DB: db}
}

// Create adds a new collection to the database
func (r *CollectionRepository) Create(ctx context.Context, collection *models.Collection) error {
	slog.InfoContext(ctx, "Creating new collection", "collectionID", collection.ID, "userID", collection.UserID)
	if err := r.DB.Create(collection).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to create collection", "error", err, "collectionID", collection.ID)
		return err
	}
	slog.InfoContext(ctx, "Collection created successfully", "collectionID", collection.ID)
	return nil
}

// FindByID retrieves a collection by its ID
func (r *CollectionRepository) FindByID(ctx context.Context, id string) (*models.Collection, error) {
	slog.InfoContext(ctx, "Finding collection by ID", "collectionID", id)
	var collection models.Collection
	err := r.DB.Preload("User").First(&collection, "id = ?", id).Error
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find collection by ID", "error", err, "collectionID", id)
		return nil, err
	}
	slog.InfoContext(ctx, "Collection found successfully", "collectionID", id)
	return &collection, nil
}

// FindByUserID retrieves all collections for a specific user
func (r *CollectionRepository) FindByUserID(ctx context.Context, userID string) ([]models.Collection, error) {
	slog.InfoContext(ctx, "Finding collections by user ID", "userID", userID)
	var collections []models.Collection
	err := r.DB.Preload("User").Where("user_id = ?", userID).Find(&collections).Error
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find collections by user ID", "error", err, "userID", userID)
		return nil, err
	}
	slog.InfoContext(ctx, "Collections found successfully", "userID", userID, "count", len(collections))
	return collections, nil
}

// Update modifies an existing collection in the database
func (r *CollectionRepository) Update(ctx context.Context, collection *models.Collection) error {
	slog.InfoContext(ctx, "Updating collection", "collectionID", collection.ID)
	if err := r.DB.Save(collection).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to update collection", "error", err, "collectionID", collection.ID)
		return err
	}
	slog.InfoContext(ctx, "Collection updated successfully", "collectionID", collection.ID)
	return nil
}

// Delete removes a collection from the database
func (r *CollectionRepository) Delete(ctx context.Context, id string) error {
	slog.InfoContext(ctx, "Deleting collection", "collectionID", id)
	if err := r.DB.Delete(&models.Collection{}, "id = ?", id).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to delete collection", "error", err, "collectionID", id)
		return err
	}
	slog.InfoContext(ctx, "Collection deleted successfully", "collectionID", id)
	return nil
}

// Ensure CollectionRepository implements CollectionRepositoryInterface
var _ CollectionRepositoryInterface = (*CollectionRepository)(nil)
