package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/repos"
)

// StashService implements the StashServiceInterface
type StashService struct {
	repo repos.StashRepositoryInterface
}

// NewStashService creates a new instance of StashService
func NewStashService(repo repos.StashRepositoryInterface) StashServiceInterface {
	return &StashService{
		repo: repo,
	}
}

// CreateStash creates a new stash for a given user
func (s *StashService) CreateStash(ctx context.Context, userID string) (*models.Stash, error) {
	slog.InfoContext(ctx, "Creating new stash", "userID", userID)

	stash := &models.Stash{
		ID:     uuid.New(),
		UserID: uuid.MustParse(userID),
	}

	if err := s.repo.Create(ctx, stash); err != nil {
		slog.ErrorContext(ctx, "Failed to create stash", "error", err, "userID", userID)
		return nil, err
	}

	slog.InfoContext(ctx, "Stash created successfully", "stashID", stash.ID, "userID", userID)
	return stash, nil
}

// FindByID retrieves a stash by its ID
func (s *StashService) FindByID(ctx context.Context, id string) (*models.Stash, error) {
	slog.InfoContext(ctx, "Finding stash by ID", "stashID", id)

	stash, err := s.repo.FindByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find stash by ID", "error", err, "stashID", id)
		return nil, err
	}

	slog.InfoContext(ctx, "Stash found", "stashID", id)
	return stash, nil
}

// FindByUserID retrieves a stash by user ID
func (s *StashService) FindByUserID(ctx context.Context, userID string) (*models.Stash, error) {
	slog.InfoContext(ctx, "Finding stash by user ID", "userID", userID)

	stash, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find stash by user ID", "error", err, "userID", userID)
		return nil, err
	}

	slog.InfoContext(ctx, "Stash found for user", "userID", userID)
	return stash, nil
}

// Ensure StashService implements StashServiceInterface
var _ StashServiceInterface = (*StashService)(nil)
