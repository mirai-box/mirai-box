package repos

import (
	"context"
	"log/slog"

	"github.com/mirai-box/mirai-box/internal/models"

	"gorm.io/gorm"
)

// RevisionRepository implements the RevisionRepositoryInterface
type RevisionRepository struct {
	DB *gorm.DB
}

// NewRevisionRepository creates a new instance of RevisionRepository
func NewRevisionRepository(db *gorm.DB) RevisionRepositoryInterface {
	return &RevisionRepository{DB: db}
}

// Create adds a new revision to the database
func (r *RevisionRepository) Create(ctx context.Context, revision *models.Revision) error {
	slog.InfoContext(ctx, "Creating new revision", "revisionID", revision.ID, "artProjectID", revision.ArtProjectID)
	if err := r.DB.Create(revision).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to create revision", "error", err, "revisionID", revision.ID)
		return err
	}
	slog.InfoContext(ctx, "Revision created successfully", "revisionID", revision.ID)
	return nil
}

// FindByID retrieves a revision by its ID
func (r *RevisionRepository) FindByID(ctx context.Context, id string) (*models.Revision, error) {
	slog.InfoContext(ctx, "Finding revision by ID", "revisionID", id)
	var revision models.Revision
	err := r.DB.Preload("ArtProject").First(&revision, "id = ?", id).Error
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find revision by ID", "error", err, "revisionID", id)
		return nil, err
	}
	slog.InfoContext(ctx, "Revision found successfully", "revisionID", id)
	return &revision, nil
}

// FindByArtProjectID retrieves all revisions for a specific art project
func (r *RevisionRepository) FindByArtProjectID(ctx context.Context, artProjectID string) ([]models.Revision, error) {
	slog.InfoContext(ctx, "Finding revisions by art project ID", "artProjectID", artProjectID)
	var revisions []models.Revision
	err := r.DB.Preload("ArtProject").Where("art_project_id = ?", artProjectID).Find(&revisions).Error
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find revisions by art project ID", "error", err, "artProjectID", artProjectID)
		return nil, err
	}
	slog.InfoContext(ctx, "Revisions found successfully", "artProjectID", artProjectID, "count", len(revisions))
	return revisions, nil
}

// Update modifies an existing revision in the database
func (r *RevisionRepository) Update(ctx context.Context, revision *models.Revision) error {
	slog.InfoContext(ctx, "Updating revision", "revisionID", revision.ID)
	if err := r.DB.Save(revision).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to update revision", "error", err, "revisionID", revision.ID)
		return err
	}
	slog.InfoContext(ctx, "Revision updated successfully", "revisionID", revision.ID)
	return nil
}

// Delete removes a revision from the database
func (r *RevisionRepository) Delete(ctx context.Context, id string) error {
	slog.InfoContext(ctx, "Deleting revision", "revisionID", id)
	if err := r.DB.Delete(&models.Revision{}, "id = ?", id).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to delete revision", "error", err, "revisionID", id)
		return err
	}
	slog.InfoContext(ctx, "Revision deleted successfully", "revisionID", id)
	return nil
}

// Ensure RevisionRepository implements RevisionRepositoryInterface
var _ RevisionRepositoryInterface = (*RevisionRepository)(nil)
