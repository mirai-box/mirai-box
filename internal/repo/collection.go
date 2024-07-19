package repo

import (
	"context"
	"errors"
	"log/slog"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/mirai-box/mirai-box/internal/model"
)

// CollectionRepository defines the interface for collection related database operations.
type CollectionRepository interface {
	CreateCollection(ctx context.Context, collection *model.Collection) error
	FindCollectionByID(ctx context.Context, id string) (*model.Collection, error)
	UpdateCollection(ctx context.Context, collection *model.Collection) error
	DeleteCollection(ctx context.Context, id string) error
	FindByUserID(ctx context.Context, userID string) ([]model.Collection, error)
	AddArtProjectToCollection(ctx context.Context, collectionID, artProjectID string) error
	RemoveArtProjectFromCollection(ctx context.Context, collectionID, artProjectID string) error
	FindArtProjectsByCollectionID(ctx context.Context, collectionID string) ([]model.CollectionArtProject, error)
	FindCollectionsByArtProjectID(ctx context.Context, artProjectID string) ([]model.CollectionArtProject, error)
}

type collectionRepo struct {
	db *gorm.DB
}

// NewCollectionRepository creates a new instance of CollectionRepository.
func NewCollectionRepository(db *gorm.DB) CollectionRepository {
	return &collectionRepo{db: db}
}

// CreateCollection adds a new collection to the database.
func (r *collectionRepo) CreateCollection(ctx context.Context, collection *model.Collection) error {
	logger := slog.With("method", "CreateCollection", "collectionID", collection.ID)

	if err := r.db.Create(collection).Error; err != nil {
		logger.Error("Failed to create collection", "error", err)
		return err
	}

	logger.Info("Collection created successfully")
	return nil
}

// FindCollectionByID retrieves a collection by its ID.
func (r *collectionRepo) FindCollectionByID(ctx context.Context, id string) (*model.Collection, error) {
	logger := slog.With("method", "FindCollectionByID", "collectionID", id)

	var collection model.Collection
	if err := r.db.Preload("User").First(&collection, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info("Collection not found")
			return nil, model.ErrCollectionNotFound
		}
		logger.Error("Failed to find collection", "error", err)
		return nil, err
	}

	logger.Info("Collection found successfully")
	return &collection, nil
}

// UpdateCollection updates an existing collection in the database.
func (r *collectionRepo) UpdateCollection(ctx context.Context, collection *model.Collection) error {
	logger := slog.With("method", "UpdateCollection", "collectionID", collection.ID)

	result := r.db.Save(collection)
	if result.Error != nil {
		logger.Error("Failed to update collection", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		logger.Info("Collection not found for update")
		return model.ErrCollectionNotFound
	}

	logger.Info("Collection updated successfully")
	return nil
}

// DeleteCollection removes a collection from the database.
func (r *collectionRepo) DeleteCollection(ctx context.Context, id string) error {
	logger := slog.With("method", "DeleteCollection", "collectionID", id)

	result := r.db.Delete(&model.Collection{}, "id = ?", id)
	if result.Error != nil {
		logger.Error("Failed to delete collection", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		logger.Info("Collection not found for deletion")
		return model.ErrCollectionNotFound
	}

	logger.Info("Collection deleted successfully")
	return nil
}

// FindByUserID retrieves all collections for a specific user.
func (r *collectionRepo) FindByUserID(ctx context.Context, userID string) ([]model.Collection, error) {
	logger := slog.With("method", "FindByUserID", "userID", userID)

	var collections []model.Collection
	if err := r.db.Where("user_id = ?", userID).Find(&collections).Error; err != nil {
		logger.Error("Failed to find collections by user ID", "error", err)
		return nil, err
	}

	if len(collections) == 0 {
		logger.Info("No collections found for user")
		return nil, model.ErrCollectionNotFound
	}

	logger.Info("Collections found successfully", "count", len(collections))
	return collections, nil
}

// AddArtProjectToCollection adds an art project to a collection.
func (r *collectionRepo) AddArtProjectToCollection(ctx context.Context, collectionID, artProjectID string) error {
	logger := slog.With("method", "AddArtProjectToCollection", "collectionID", collectionID, "artProjectID", artProjectID)

	collectionArtProject := model.CollectionArtProject{
		CollectionID: uuid.MustParse(collectionID),
		ArtProjectID: uuid.MustParse(artProjectID),
	}

	if err := r.db.Create(&collectionArtProject).Error; err != nil {
		logger.Error("Failed to add art project to collection", "error", err)
		return err
	}

	logger.Info("Art project added to collection successfully")
	return nil
}

// RemoveArtProjectFromCollection removes an art project from a collection.
func (r *collectionRepo) RemoveArtProjectFromCollection(ctx context.Context, collectionID, artProjectID string) error {
	logger := slog.With("method", "RemoveArtProjectFromCollection", "collectionID", collectionID, "artProjectID", artProjectID)

	result := r.db.Where("collection_id = ? AND art_project_id = ?", collectionID, artProjectID).
		Delete(&model.CollectionArtProject{})

	if result.Error != nil {
		logger.Error("Failed to remove art project from collection", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		logger.Info("Art project not found in collection")
		return model.ErrCollectionNotFound
	}

	logger.Info("Art project removed from collection successfully")
	return nil
}

// FindArtProjectsByCollectionID retrieves all art projects in a specific collection.
func (r *collectionRepo) FindArtProjectsByCollectionID(ctx context.Context, collectionID string) ([]model.CollectionArtProject, error) {
	logger := slog.With("method", "FindArtProjectsByCollectionID", "collectionID", collectionID)

	var collectionArtProjects []model.CollectionArtProject
	if err := r.db.Where("collection_id = ?", collectionID).Find(&collectionArtProjects).Error; err != nil {
		logger.Error("Failed to find art projects by collection ID", "error", err)
		return nil, err
	}

	if len(collectionArtProjects) == 0 {
		logger.Info("No art projects found in collection")
		return nil, model.ErrCollectionNotFound
	}

	logger.Info("Art projects found successfully", "count", len(collectionArtProjects))
	return collectionArtProjects, nil
}

// FindCollectionsByArtProjectID retrieves all collections that contain a specific art project.
func (r *collectionRepo) FindCollectionsByArtProjectID(ctx context.Context, artProjectID string) ([]model.CollectionArtProject, error) {
	logger := slog.With("method", "FindCollectionsByArtProjectID", "artProjectID", artProjectID)

	var collectionArtProjects []model.CollectionArtProject
	if err := r.db.Where("art_project_id = ?", artProjectID).Find(&collectionArtProjects).Error; err != nil {
		logger.Error("Failed to find collections by art project ID", "error", err)
		return nil, err
	}

	if len(collectionArtProjects) == 0 {
		logger.Info("No collections found containing the art project")
		return nil, model.ErrCollectionNotFound
	}

	logger.Info("Collections found successfully", "count", len(collectionArtProjects))
	return collectionArtProjects, nil
}
