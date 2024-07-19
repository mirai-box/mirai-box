package repo

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/mirai-box/mirai-box/internal/model"
)

// ArtProjectRepository defines the interface for art project related database operations.
type ArtProjectRepository interface {
	CreateArtProject(ctx context.Context, artProject *model.ArtProject) error
	FindArtProjectByID(ctx context.Context, id string) (*model.ArtProject, error)
	UpdateArtProject(ctx context.Context, artProject *model.ArtProject) error
	DeleteArtProject(ctx context.Context, id string) error
	SaveArtProjectAndRevision(ctx context.Context, artProject *model.ArtProject, revision *model.Revision) error
	SaveRevision(ctx context.Context, revision *model.Revision) error
	UpdateLatestRevision(ctx context.Context, artProjectID string, revisionID uuid.UUID) error
	ListLatestRevisions(ctx context.Context, userID string) ([]model.Revision, error)
	ListAllArtProjects(ctx context.Context, userID string) ([]model.ArtProject, error)
	ListAllRevisions(ctx context.Context, artProjectID string) ([]model.Revision, error)
	GetMaxRevisionVersion(ctx context.Context, artProjectID string) (int, error)
	GetRevisionByID(ctx context.Context, id string) (*model.Revision, error)
	FindByStashID(ctx context.Context, stashID string) ([]model.ArtProject, error)
	FindByUserID(ctx context.Context, userID string) ([]model.ArtProject, error)
	FindRevisionByID(ctx context.Context, id string) (*model.Revision, error)
}

type artProjectRepo struct {
	db *gorm.DB
}

// NewArtProjectRepository creates a new instance of ArtProjectRepository.
func NewArtProjectRepository(db *gorm.DB) ArtProjectRepository {
	return &artProjectRepo{db: db}
}

// CreateArtProject adds a new art project to the database.
func (r *artProjectRepo) CreateArtProject(ctx context.Context, artProject *model.ArtProject) error {
	logger := slog.With("method", "CreateArtProject", "artProjectID", artProject.ID)

	if err := r.db.Create(artProject).Error; err != nil {
		logger.Error("Failed to create art project", "error", err)
		return err
	}

	logger.Info("Art project created successfully")
	return nil
}

// FindArtProjectByID retrieves an art project by its ID.
func (r *artProjectRepo) FindArtProjectByID(ctx context.Context, id string) (*model.ArtProject, error) {
	logger := slog.With("method", "FindArtProjectByID", "artProjectID", id)

	var artProject model.ArtProject
	if err := r.db.Preload("Stash").Preload("User").First(&artProject, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info("Art project not found")
			return nil, model.ErrArtProjectNotFound
		}
		logger.Error("Failed to find art project", "error", err)
		return nil, err
	}

	logger.Info("Art project found successfully")
	return &artProject, nil
}

// UpdateArtProject updates an existing art project in the database.
func (r *artProjectRepo) UpdateArtProject(ctx context.Context, artProject *model.ArtProject) error {
	logger := slog.With("method", "UpdateArtProject", "artProjectID", artProject.ID)

	result := r.db.Save(artProject)
	if result.Error != nil {
		logger.Error("Failed to update art project", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		logger.Info("Art project not found for update")
		return model.ErrArtProjectNotFound
	}

	logger.Info("Art project updated successfully")
	return nil
}

// DeleteArtProject removes an art project from the database.
func (r *artProjectRepo) DeleteArtProject(ctx context.Context, id string) error {
	logger := slog.With("method", "DeleteArtProject", "artProjectID", id)

	result := r.db.Delete(&model.ArtProject{}, "id = ?", id)
	if result.Error != nil {
		logger.Error("Failed to delete art project", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		logger.Info("Art project not found for deletion")
		return model.ErrArtProjectNotFound
	}

	logger.Info("Art project deleted successfully")
	return nil
}

// SaveArtProjectAndRevision saves both an art project and its revision in a single transaction.
func (r *artProjectRepo) SaveArtProjectAndRevision(ctx context.Context, artProject *model.ArtProject, revision *model.Revision) error {
	logger := slog.With("method", "SaveArtProjectAndRevision", "artProjectID", artProject.ID, "revisionID", revision.ID)

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(artProject).Error; err != nil {
			slog.ErrorContext(ctx, "Failed to create art project in transaction", "error", err, "projectID", artProject.ID)
			return err
		}

		if err := tx.Create(revision).Error; err != nil {
			slog.ErrorContext(ctx, "Failed to create revision in transaction", "error", err, "revisionID", revision.ID)
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error("Failed to save art project and revision", "error", err)
		return err
	}

	logger.Info("Art project and revision saved successfully")
	return nil
}

// SaveRevision adds a new revision to the database
func (r *artProjectRepo) SaveRevision(ctx context.Context, revision *model.Revision) error {
	logger := slog.With("method", "SaveRevision", "revisionID", revision.ID)

	if err := r.db.Create(revision).Error; err != nil {
		logger.ErrorContext(ctx, "Failed to save new art revision", "error", err)
		return err
	}

	slog.InfoContext(ctx, "New art revision saved successfully")
	return nil
}

// UpdateLatestRevision updates the latest revision ID for an art project.
func (r *artProjectRepo) UpdateLatestRevision(ctx context.Context, artProjectID string, revisionID uuid.UUID) error {
	logger := slog.With("method", "UpdateLatestRevision", "artProjectID", artProjectID, "revisionID", revisionID)

	if err := r.db.Model(&model.ArtProject{}).Where("id = ?", artProjectID).Update("latest_revision_id", revisionID).Error; err != nil {
		logger.Error("Failed to update latest revision", "error", err)
		return err
	}

	logger.Info("Latest revision updated successfully")
	return nil
}

// ListLatestRevisions retrieves the latest revisions for all art projects belonging to the specified user.
func (r *artProjectRepo) ListLatestRevisions(ctx context.Context, userID string) ([]model.Revision, error) {
	logger := slog.With("method", "ListLatestRevisions", "userID", userID)

	var revisions []model.Revision
	if err := r.db.Joins("JOIN art_projects ON art_projects.latest_revision_id = revisions.id").
		Where("art_projects.user_id = ?", userID).
		Find(&revisions).Error; err != nil {
		logger.Error("Failed to list latest revisions", "error", err)
		return nil, err
	}

	if len(revisions) == 0 {
		logger.Info("No revisions found for user")
		return nil, model.ErrArtProjectNotFound
	}

	logger.Info("Latest revisions listed successfully", "count", len(revisions))
	return revisions, nil
}

// ListAllArtProjects retrieves all art projects belonging to the specified user.
func (r *artProjectRepo) ListAllArtProjects(ctx context.Context, userID string) ([]model.ArtProject, error) {
	logger := slog.With("method", "ListAllArtProjects", "userID", userID)

	var artProjects []model.ArtProject
	if err := r.db.Where("user_id = ?", userID).Find(&artProjects).Error; err != nil {
		logger.Error("Failed to list all art projects", "error", err)
		return nil, err
	}

	if len(artProjects) == 0 {
		logger.Info("No art projects found for user")
		return nil, model.ErrArtProjectNotFound
	}

	logger.Info("All art projects listed successfully", "count", len(artProjects))
	return artProjects, nil
}

// ListAllRevisions retrieves all revisions for a specific art project.
func (r *artProjectRepo) ListAllRevisions(ctx context.Context, artProjectID string) ([]model.Revision, error) {
	logger := slog.With("method", "ListAllRevisions", "artProjectID", artProjectID)

	var revisions []model.Revision
	if err := r.db.Where("art_project_id = ?", artProjectID).Find(&revisions).Error; err != nil {
		logger.Error("Failed to list all revisions", "error", err)
		return nil, err
	}

	if len(revisions) == 0 {
		logger.Info("No revisions found for art project")
		return nil, model.ErrArtProjectNotFound
	}

	logger.Info("All revisions listed successfully", "count", len(revisions))
	return revisions, nil
}

// GetMaxRevisionVersion retrieves the maximum revision version for a specific art project.
func (r *artProjectRepo) GetMaxRevisionVersion(ctx context.Context, artProjectID string) (int, error) {
	logger := slog.With("method", "GetMaxRevisionVersion", "artProjectID", artProjectID)

	var maxVersion int
	if err := r.db.Model(&model.Revision{}).Where("art_project_id = ?", artProjectID).Select("COALESCE(MAX(version), 0)").Scan(&maxVersion).Error; err != nil {
		logger.Error("Failed to get max revision version", "error", err)
		return 0, err
	}

	logger.Info("Max revision version retrieved successfully", "maxVersion", maxVersion)
	return maxVersion, nil
}

// FindByStashID retrieves all art projects associated with a specific stash ID.
func (r *artProjectRepo) FindByStashID(ctx context.Context, stashID string) ([]model.ArtProject, error) {
	logger := slog.With("method", "FindByStashID", "stashID", stashID)

	var artProjects []model.ArtProject
	if err := r.db.Where("stash_id = ?", stashID).Find(&artProjects).Error; err != nil {
		logger.Error("Failed to find art projects by stash ID", "error", err)
		return nil, err
	}

	if len(artProjects) == 0 {
		logger.Info("No art projects found for stash ID")
		return nil, model.ErrArtProjectNotFound
	}

	logger.Info("Art projects found successfully", "count", len(artProjects))
	return artProjects, nil
}

// FindByUserID retrieves all art projects for a specific user.
func (r *artProjectRepo) FindByUserID(ctx context.Context, userID string) ([]model.ArtProject, error) {
	logger := slog.With("method", "FindByUserID", "userID", userID)

	var artProjects []model.ArtProject
	if err := r.db.Where("user_id = ?", userID).Find(&artProjects).Error; err != nil {
		logger.Error("Failed to find art projects by user ID", "error", err)
		return nil, err
	}

	if len(artProjects) == 0 {
		logger.Info("No art projects found for user")
		return nil, model.ErrArtProjectNotFound
	}

	logger.Info("Art projects found successfully", "count", len(artProjects))
	return artProjects, nil
}

// FindRevisionByID retrieves a revision by its ID
func (r *artProjectRepo) FindRevisionByID(ctx context.Context, id string) (*model.Revision, error) {
	logger := slog.With("method", "FindRevisionByID", "revisionID", id)

	var revision model.Revision
	if err := r.db.Preload("ArtProject").First(&revision, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info("Revision not found")
			return nil, model.ErrArtProjectNotFound
		}
		logger.Error("Failed to get revision by ID", "error", err)
		return nil, err
	}

	logger.Info("Revision retrieved successfully")
	return &revision, nil
}

// GetRevisionByID retrieves a revision by its ID.
func (r *artProjectRepo) GetRevisionByID(ctx context.Context, id string) (*model.Revision, error) {
	logger := slog.With("method", "GetRevisionByID", "revisionID", id)

	var revision model.Revision
	if err := r.db.First(&revision, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info("Revision not found")
			return nil, model.ErrArtProjectNotFound
		}

		logger.Error("Failed to get revision", "error", err)
		return nil, err
	}

	logger.Info("Revision retrieved successfully")
	return &revision, nil
}
