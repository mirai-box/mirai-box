package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/repos"
)

// ArtProjectService implements the ArtProjectServiceInterface
type artProjectService struct {
	repo      repos.ArtProjectRepositoryInterface
	stashRepo repos.StashRepositoryInterface
}

// NewArtProjectService creates a new instance of ArtProjectService
func NewArtProjectService(
	repo repos.ArtProjectRepositoryInterface,
	stashRepo repos.StashRepositoryInterface,
) ArtProjectServiceInterface {
	return &artProjectService{
		repo:      repo,
		stashRepo: stashRepo,
	}
}

// CreateArtProject creates a new art project
func (s *artProjectService) CreateArtProject(ctx context.Context, stashID, title string) (*models.ArtProject, error) {
	artProject := &models.ArtProject{
		ID:        uuid.New(),
		StashID:   uuid.MustParse(stashID),
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	slog.InfoContext(ctx, "Creating new art project",
		"artProjectID", artProject.ID,
		"stashID", stashID,
		"title", title,
	)

	if err := s.repo.Create(ctx, artProject); err != nil {
		slog.ErrorContext(ctx, "Failed to create art project",
			"error", err,
			"artProjectID", artProject.ID,
			"stashID", stashID,
		)
		return nil, err
	}

	slog.InfoContext(ctx, "Art project created successfully",
		"artProjectID", artProject.ID,
		"stashID", stashID,
	)

	return artProject, nil
}

// FindByID finds an art project by its ID
func (s *artProjectService) FindByID(ctx context.Context, id string) (*models.ArtProject, error) {
	slog.InfoContext(ctx, "ArtProjectService: Finding art project by ID", "artProjectID", id)

	artProject, err := s.repo.FindByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "ArtProjectService: Failed to find art project by ID",
			"error", err,
			"artProjectID", id,
		)
		return nil, err
	}

	slog.InfoContext(ctx, "ArtProjectService: Art project found successfully", "artProjectID", id)
	return artProject, nil
}

// FindByStashID finds all art projects by a stash ID
func (s *artProjectService) FindByStashID(ctx context.Context, stashID string) ([]models.ArtProject, error) {
	slog.InfoContext(ctx, "Finding art projects by stash ID", "stashID", stashID)

	artProjects, err := s.repo.FindByStashID(ctx, stashID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find art projects by stash ID",
			"error", err,
			"stashID", stashID,
		)
		return nil, err
	}

	slog.InfoContext(ctx, "Art projects found successfully",
		"stashID", stashID,
		"count", len(artProjects),
	)
	return artProjects, nil
}

// FindByStashID finds all art projects by a stash ID
func (s *artProjectService) FindByUserID(ctx context.Context, userID string) ([]models.ArtProject, error) {
	slog.InfoContext(ctx, "Finding art projects by user ID", "userID", userID)

	artProjects, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find art projects by user ID",
			"error", err,
			"userID", userID,
		)
		return nil, err
	}

	return artProjects, nil
}

// Ensure ArtProjectService implements ArtProjectServiceInterface
var _ ArtProjectServiceInterface = (*artProjectService)(nil)
