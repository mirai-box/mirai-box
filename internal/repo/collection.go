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
	AddRevisionToCollection(ctx context.Context, collectionID, revisionID string) error
	RemoveRevisionFromCollection(ctx context.Context, collectionID, revisionID string) error
	GetRevisionsByCollectionID(ctx context.Context, collectionID string) ([]model.Revision, error)
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

// AddRevisionToCollection adds an art project to a collection.
func (r *collectionRepo) AddRevisionToCollection(ctx context.Context, collectionID, revisionID string) error {
	logger := slog.With("repo", "AddRevisionToCollection", "collectionID", collectionID, "revisionID", revisionID)

	collectionArtProject := model.CollectionArtProject{
		CollectionID: uuid.MustParse(collectionID),
		RevisionID:   uuid.MustParse(revisionID),
	}

	if err := r.db.Create(&collectionArtProject).Error; err != nil {
		logger.Error("Failed to add revision to collection", "error", err)
		return err
	}

	logger.Info("Art project revision added to collection successfully")
	return nil
}

// RemoveRevisionFromCollection removes an art project from a collection.
func (r *collectionRepo) RemoveRevisionFromCollection(ctx context.Context, collectionID, artProjectID string) error {
	logger := slog.With("repo", "RemoveRevisionFromCollection", "collectionID", collectionID, "artProjectID", artProjectID)

	result := r.db.Where("collection_id = ? AND revision_id = ?", collectionID, artProjectID).
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

func (r *collectionRepo) GetRevisionsByCollectionID(ctx context.Context, collectionID string) ([]model.Revision, error) {
	logger := slog.With("repo", "GetRevisionsByCollectionID", "collectionID", collectionID)

	var revisions []model.Revision
	err := r.db.Table("revisions").
		Joins("JOIN collection_art_projects ON revisions.id = collection_art_projects.revision_id").
		Where("collection_art_projects.collection_id = ?", collectionID).
		Find(&revisions).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info("No revisions found for the collection")
			return nil, model.ErrRevisionNotFound
		}
		logger.Error("Failed to retrieve revisions for the collection", "error", err)
		return nil, err
	}

	if len(revisions) == 0 {
		logger.Info("No revisions found for the collection")
		return nil, model.ErrRevisionNotFound
	}

	logger.Info("Revisions retrieved successfully", "count", len(revisions))
	return revisions, nil
}
