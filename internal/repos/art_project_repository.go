package repos

import (
	"context"
	"log/slog"

	"github.com/mirai-box/mirai-box/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ArtProjectRepository implements the ArtProjectRepositoryInterface
type ArtProjectRepository struct {
	DB *gorm.DB
}

// NewArtProjectRepository creates a new instance of ArtProjectRepository
func NewArtProjectRepository(db *gorm.DB) ArtProjectRepositoryInterface {
	return &ArtProjectRepository{DB: db}
}

// Create adds a new art project to the database
func (r *ArtProjectRepository) Create(ctx context.Context, artProject *models.ArtProject) error {
	slog.InfoContext(ctx, "Creating new art project", "projectID", artProject.ID)
	if err := r.DB.Create(artProject).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to create art project", "error", err, "projectID", artProject.ID)
		return err
	}
	slog.InfoContext(ctx, "Art project created successfully", "projectID", artProject.ID)
	return nil
}

// FindByID retrieves an art project by its ID
func (r *ArtProjectRepository) FindByID(ctx context.Context, id string) (*models.ArtProject, error) {
	slog.InfoContext(ctx, "Finding art project by ID", "projectID", id)
	var artProject models.ArtProject
	if err := r.DB.First(&artProject, "id = ?", id).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to find art project", "error", err, "projectID", id)
		return nil, err
	}
	slog.InfoContext(ctx, "Art project found", "projectID", id)
	return &artProject, nil
}

// SaveArtProjectAndRevision saves both an art project and its revision in a single transaction
func (r *ArtProjectRepository) SaveArtProjectAndRevision(ctx context.Context,
	artProject *models.ArtProject,
	revision *models.Revision) error {

	slog.InfoContext(ctx, "Saving art project and revision",
		"artProject", artProject,
		"revision", revision,
	)

	err := r.DB.Transaction(func(tx *gorm.DB) error {
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
		slog.ErrorContext(ctx, "Transaction failed when saving art project and revision", "error", err)
		return err
	}
	slog.InfoContext(ctx, "Art project and revision saved successfully", "projectID", artProject.ID, "revisionID", revision.ID)
	return nil
}

// SaveRevision adds a new revision to the database
func (r *ArtProjectRepository) SaveRevision(ctx context.Context, revision *models.Revision) error {
	slog.InfoContext(ctx, "Saving revision", "revisionID", revision.ID)
	if err := r.DB.Create(revision).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to save revision", "error", err, "revisionID", revision.ID)
		return err
	}
	slog.InfoContext(ctx, "Revision saved successfully", "revisionID", revision.ID)
	return nil
}

// UpdateLatestRevision updates the latest revision ID for an art project
func (r *ArtProjectRepository) UpdateLatestRevision(ctx context.Context, artProjectID string, revisionID uuid.UUID) error {
	slog.InfoContext(ctx, "Updating latest revision", "projectID", artProjectID, "revisionID", revisionID)
	if err := r.DB.Model(&models.ArtProject{}).Where("id = ?", artProjectID).Update("latest_revision_id", revisionID).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to update latest revision", "error", err, "projectID", artProjectID)
		return err
	}
	slog.InfoContext(ctx, "Latest revision updated successfully", "projectID", artProjectID, "revisionID", revisionID)
	return nil
}

// ListLatestRevisions retrieves the latest revisions for all art projects belonging to the specified user
func (r *ArtProjectRepository) ListLatestRevisions(ctx context.Context, userID string) ([]models.Revision, error) {
	slog.InfoContext(ctx, "Listing latest revisions for all art projects")
	var revisions []models.Revision
	if err := r.DB.Raw(`
        SELECT r.*
        FROM revisions r
        JOIN art_projects a ON a.latest_revision_id = r.id
        JOIN stashes s ON a.stash_id = s.id
        WHERE s.user_id = ?
    `, userID).Scan(&revisions).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to list latest revisions", "error", err)
		return nil, err
	}
	slog.InfoContext(ctx, "Latest revisions retrieved successfully", "count", len(revisions))
	return revisions, nil
}

// ListAllArtProjects retrieves all art projects belonging to the specified user from the database
func (r *ArtProjectRepository) ListAllArtProjects(ctx context.Context, userID string) ([]models.ArtProject, error) {
	slog.InfoContext(ctx, "Listing all art projects")
	var artProjects []models.ArtProject
	if err := r.DB.Preload("Tags").Preload("Category").Preload("Stash").Where("stash_id IN (SELECT id FROM stashes WHERE user_id = ?)", userID).Find(&artProjects).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to list all art projects", "error", err)
		return nil, err
	}
	slog.InfoContext(ctx, "All art projects retrieved successfully", "count", len(artProjects))
	return artProjects, nil
}

// ListAllRevisions retrieves all revisions for a specific art project
func (r *ArtProjectRepository) ListAllRevisions(ctx context.Context, artProjectID string) ([]models.Revision, error) {
	slog.InfoContext(ctx, "Listing all revisions for art project", "projectID", artProjectID)
	var revisions []models.Revision
	if err := r.DB.Where("art_project_id = ?", artProjectID).Find(&revisions).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to list all revisions for art project", "error", err, "projectID", artProjectID)
		return nil, err
	}
	slog.InfoContext(ctx, "All revisions retrieved successfully", "projectID", artProjectID, "count", len(revisions))
	return revisions, nil
}

// GetMaxRevisionVersion retrieves the maximum revision version for a specific art project
func (r *ArtProjectRepository) GetMaxRevisionVersion(ctx context.Context, artProjectID string) (int, error) {
	slog.InfoContext(ctx, "Getting max revision version for art project", "projectID", artProjectID)
	var maxVersion int
	if err := r.DB.Model(&models.Revision{}).Where("art_project_id = ?", artProjectID).Select("MAX(version)").Scan(&maxVersion).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to get max revision version", "error", err, "projectID", artProjectID)
		return 0, err
	}
	slog.InfoContext(ctx, "Max revision version retrieved successfully", "projectID", artProjectID, "maxVersion", maxVersion)
	return maxVersion, nil
}

// GetRevisionByArtID retrieves a revision by its art ID
func (r *ArtProjectRepository) GetRevisionByArtID(ctx context.Context, artID string) (*models.Revision, error) {
	slog.InfoContext(ctx, "Getting revision by art ID", "artID", artID)
	var revision models.Revision
	if err := r.DB.First(&revision, "art_id = ?", artID).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to get revision by art ID", "error", err, "artID", artID)
		return nil, err
	}
	slog.InfoContext(ctx, "Revision retrieved successfully", "artID", artID, "revisionID", revision.ID)
	return &revision, nil
}

// GetRevisionByID retrieves a revision by its ID
func (r *ArtProjectRepository) GetRevisionByID(ctx context.Context, id string) (*models.Revision, error) {
	slog.InfoContext(ctx, "Getting revision by ID", "revisionID", id)
	var revision models.Revision
	if err := r.DB.First(&revision, "id = ?", id).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to get revision by ID", "error", err, "revisionID", id)
		return nil, err
	}
	slog.InfoContext(ctx, "Revision retrieved successfully", "revisionID", id)
	return &revision, nil
}

// GetArtProjectByID retrieves an art project by its ID
func (r *ArtProjectRepository) GetArtProjectByID(ctx context.Context, id string) (*models.ArtProject, error) {
	slog.InfoContext(ctx, "Getting art project by ID", "projectID", id)
	var artProject models.ArtProject
	if err := r.DB.First(&artProject, "id = ?", id).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to get art project by ID", "error", err, "projectID", id)
		return nil, err
	}
	slog.InfoContext(ctx, "Art project retrieved successfully", "projectID", id)
	return &artProject, nil
}

// FindByStashID retrieves all art projects associated with a specific stash ID
func (r *ArtProjectRepository) FindByStashID(ctx context.Context, stashID string) ([]models.ArtProject, error) {
	slog.InfoContext(ctx, "Finding art projects by stash ID", "stashID", stashID)
	var artProjects []models.ArtProject
	if err := r.DB.Where("stash_id = ?", stashID).Find(&artProjects).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to find art projects by stash ID", "error", err, "stashID", stashID)
		return nil, err
	}
	slog.InfoContext(ctx, "Art projects found successfully", "stashID", stashID, "count", len(artProjects))
	return artProjects, nil
}

// Ensure ArtProjectRepository implements ArtProjectRepositoryInterface
var _ ArtProjectRepositoryInterface = (*ArtProjectRepository)(nil)
