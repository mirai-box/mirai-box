package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/repos"
)

// RevisionService implements the RevisionServiceInterface
type revisionService struct {
	repo repos.RevisionRepositoryInterface
}

// NewRevisionService creates a new instance of RevisionService
func NewRevisionService(repo repos.RevisionRepositoryInterface) RevisionServiceInterface {
	return &revisionService{repo: repo}
}

// CreateRevision creates a new revision for an art project
func (s *revisionService) CreateRevision(ctx context.Context, artProjectID, filePath, comment string, version int, size int64) (*models.Revision, error) {
	artProjectUUID := uuid.MustParse(artProjectID)

	revision := &models.Revision{
		ID:           uuid.New(),
		ArtProjectID: artProjectUUID,
		FilePath:     filePath,
		Comment:      comment,
		Version:      version,
		Size:         size,
		CreatedAt:    time.Now(),
	}

	slog.InfoContext(ctx, "Creating new revision",
		"artProjectID", artProjectID,
		"filePath", filePath,
		"version", version,
	)

	if err := s.repo.Create(ctx, revision); err != nil {
		slog.ErrorContext(ctx, "Failed to create revision",
			"error", err,
			"artProjectID", artProjectID,
			"filePath", filePath,
			"version", version,
		)
		return nil, err
	}

	slog.InfoContext(ctx, "Revision created successfully",
		"revisionID", revision.ID,
		"artProjectID", artProjectID,
	)

	return revision, nil
}

// FindByID finds a revision by its ID
func (s *revisionService) FindByID(ctx context.Context, id string) (*models.Revision, error) {
	slog.InfoContext(ctx, "Finding revision by ID", "revisionID", id)

	revision, err := s.repo.FindByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find revision by ID",
			"error", err,
			"revisionID", id,
		)
		return nil, err
	}

	slog.InfoContext(ctx, "Revision found successfully", "revisionID", id)
	return revision, nil
}

// FindByArtProjectID finds all revisions for a specific art project by the art project ID
func (s *revisionService) FindByArtProjectID(ctx context.Context, artProjectID string) ([]models.Revision, error) {
	slog.InfoContext(ctx, "Finding revisions by art project ID", "artProjectID", artProjectID)

	revisions, err := s.repo.FindByArtProjectID(ctx, artProjectID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find revisions by art project ID",
			"error", err,
			"artProjectID", artProjectID,
		)
		return nil, err
	}

	slog.InfoContext(ctx, "Revisions found successfully",
		"artProjectID", artProjectID,
		"count", len(revisions),
	)
	return revisions, nil
}

// Ensure RevisionService implements RevisionServiceInterface
var _ RevisionServiceInterface = (*revisionService)(nil)
