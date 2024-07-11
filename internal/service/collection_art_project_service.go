package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/repos"
)

// CollectionArtProjectService implements the CollectionArtProjectServiceInterface
type CollectionArtProjectService struct {
	repo repos.CollectionArtProjectRepositoryInterface
}

// NewCollectionArtProjectService creates a new instance of CollectionArtProjectService
func NewCollectionArtProjectService(repo repos.CollectionArtProjectRepositoryInterface) CollectionArtProjectServiceInterface {
	return &CollectionArtProjectService{repo: repo}
}

// AddArtProjectToCollection adds an art project to a collection
func (s *CollectionArtProjectService) AddArtProjectToCollection(ctx context.Context, collectionID, artProjectID string) (*models.CollectionArtProject, error) {
	cap := &models.CollectionArtProject{
		CollectionID: uuid.MustParse(collectionID),
		ArtProjectID: uuid.MustParse(artProjectID),
	}

	slog.InfoContext(ctx, "Adding art project to collection",
		"collectionID", collectionID,
		"artProjectID", artProjectID,
	)

	if err := s.repo.Create(ctx, cap); err != nil {
		slog.ErrorContext(ctx, "Failed to add art project to collection",
			"error", err,
			"collectionID", collectionID,
			"artProjectID", artProjectID,
		)
		return nil, err
	}

	slog.InfoContext(ctx, "Art project added to collection successfully",
		"collectionID", collectionID,
		"artProjectID", artProjectID,
	)

	return cap, nil
}

// FindByCollectionID finds all art projects in a collection by the collection ID
func (s *CollectionArtProjectService) FindByCollectionID(ctx context.Context, collectionID string) ([]models.CollectionArtProject, error) {
	slog.InfoContext(ctx, "Finding art projects by collection ID", "collectionID", collectionID)

	caps, err := s.repo.FindByCollectionID(ctx, collectionID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find art projects by collection ID",
			"error", err,
			"collectionID", collectionID,
		)
		return nil, err
	}

	slog.InfoContext(ctx, "Art projects found successfully",
		"collectionID", collectionID,
		"count", len(caps),
	)
	return caps, nil
}

// FindByArtProjectID finds all collections containing a specific art project by the art project ID
func (s *CollectionArtProjectService) FindByArtProjectID(ctx context.Context, artProjectID string) ([]models.CollectionArtProject, error) {
	slog.InfoContext(ctx, "Finding collections by art project ID", "artProjectID", artProjectID)

	caps, err := s.repo.FindByArtProjectID(ctx, artProjectID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find collections by art project ID",
			"error", err,
			"artProjectID", artProjectID,
		)
		return nil, err
	}

	slog.InfoContext(ctx, "Collections found successfully",
		"artProjectID", artProjectID,
		"count", len(caps),
	)
	return caps, nil
}

// Ensure CollectionArtProjectService implements CollectionArtProjectServiceInterface
var _ CollectionArtProjectServiceInterface = (*CollectionArtProjectService)(nil)
