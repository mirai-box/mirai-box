package repos

import (
	"context"
	"log/slog"

	"github.com/mirai-box/mirai-box/internal/models"

	"gorm.io/gorm"
)

// CollectionArtProjectRepository implements the CollectionArtProjectRepositoryInterface
type collectionArtProjectRepository struct {
	DB *gorm.DB
}

// NewCollectionArtProjectRepository creates a new instance of CollectionArtProjectRepository
func NewCollectionArtProjectRepository(db *gorm.DB) CollectionArtProjectRepositoryInterface {
	return &collectionArtProjectRepository{DB: db}
}

// Create adds a new collection art project to the database
func (r *collectionArtProjectRepository) Create(ctx context.Context, collectionArtProject *models.CollectionArtProject) error {
	slog.InfoContext(ctx, "Creating new collection art project", "collectionID", collectionArtProject.CollectionID, "artProjectID", collectionArtProject.ArtProjectID)
	if err := r.DB.Create(collectionArtProject).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to create collection art project", "error", err, "collectionID", collectionArtProject.CollectionID, "artProjectID", collectionArtProject.ArtProjectID)
		return err
	}
	slog.InfoContext(ctx, "Collection art project created successfully", "collectionID", collectionArtProject.CollectionID, "artProjectID", collectionArtProject.ArtProjectID)
	return nil
}

// FindByCollectionID retrieves all collection art projects for a specific collection
func (r *collectionArtProjectRepository) FindByCollectionID(ctx context.Context, collectionID string) ([]models.CollectionArtProject, error) {
	slog.InfoContext(ctx, "Finding collection art projects by collection ID", "collectionID", collectionID)
	var collectionArtProjects []models.CollectionArtProject
	err := r.DB.Preload("Collection").Preload("ArtProject").Where("collection_id = ?", collectionID).Find(&collectionArtProjects).Error
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find collection art projects by collection ID", "error", err, "collectionID", collectionID)
		return nil, err
	}
	slog.InfoContext(ctx, "Collection art projects found successfully", "collectionID", collectionID, "count", len(collectionArtProjects))
	return collectionArtProjects, nil
}

// FindByArtProjectID retrieves all collection art projects for a specific art project
func (r *collectionArtProjectRepository) FindByArtProjectID(ctx context.Context, artProjectID string) ([]models.CollectionArtProject, error) {
	slog.InfoContext(ctx, "Finding collection art projects by art project ID", "artProjectID", artProjectID)
	var collectionArtProjects []models.CollectionArtProject
	err := r.DB.Preload("Collection").Preload("ArtProject").Where("art_project_id = ?", artProjectID).Find(&collectionArtProjects).Error
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find collection art projects by art project ID", "error", err, "artProjectID", artProjectID)
		return nil, err
	}
	slog.InfoContext(ctx, "Collection art projects found successfully", "artProjectID", artProjectID, "count", len(collectionArtProjects))
	return collectionArtProjects, nil
}

// Delete removes a collection art project from the database
func (r *collectionArtProjectRepository) Delete(ctx context.Context, collectionID string, artProjectID string) error {
	slog.InfoContext(ctx, "Deleting collection art project", "collectionID", collectionID, "artProjectID", artProjectID)
	if err := r.DB.Delete(&models.CollectionArtProject{}, "collection_id = ? AND art_project_id = ?", collectionID, artProjectID).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to delete collection art project", "error", err, "collectionID", collectionID, "artProjectID", artProjectID)
		return err
	}
	slog.InfoContext(ctx, "Collection art project deleted successfully", "collectionID", collectionID, "artProjectID", artProjectID)
	return nil
}

// Ensure CollectionArtProjectRepository implements CollectionArtProjectRepositoryInterface
var _ CollectionArtProjectRepositoryInterface = (*collectionArtProjectRepository)(nil)
