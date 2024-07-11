package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/repos"
)

// ArtProjectRetrievalService implements the ArtProjectRetrievalServiceInterface
type ArtProjectRetrievalService struct {
	artProjectRepo repos.ArtProjectRepositoryInterface
	storageRepo    repos.StorageRepositoryInterface
}

// NewArtProjectRetrievalService creates a new instance of ArtProjectRetrievalService
func NewArtProjectRetrievalService(artProjectRepo repos.ArtProjectRepositoryInterface, storageRepo repos.StorageRepositoryInterface) ArtProjectRetrievalServiceInterface {
	return &ArtProjectRetrievalService{
		artProjectRepo: artProjectRepo,
		storageRepo:    storageRepo,
	}
}

// GetSharedArtProject retrieves the latest shared revision of an art project
func (aps *ArtProjectRetrievalService) GetSharedArtProject(ctx context.Context, artID string) (*os.File, *models.ArtProject, error) {
	slog.InfoContext(ctx, "Retrieving shared art project", "artID", artID)

	rev, err := aps.artProjectRepo.GetRevisionByArtID(ctx, artID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get revision for art ID", "error", err, "artID", artID)
		return nil, nil, err
	}

	artProject, err := aps.artProjectRepo.GetArtProjectByID(ctx, rev.ArtProjectID.String())
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get art project", "error", err, "artProjectID", rev.ArtProjectID)
		return nil, nil, err
	}

	userID := artProject.Stash.UserID.String()

	file, err := aps.storageRepo.GetRevision(ctx, userID, artProject.ID.String(), rev.Version)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get file from storage", "error", err, "artProjectID", artProject.ID)
		return nil, nil, err
	}

	slog.InfoContext(ctx, "Successfully retrieved shared art project", "artID", artID, "userID", userID)
	return file, artProject, nil
}

// GetArtProjectByRevision retrieves a specific revision of an art project
func (aps *ArtProjectRetrievalService) GetArtProjectByRevision(ctx context.Context, userID, artProjectID, revisionID string) (*os.File, *models.ArtProject, error) {
	slog.InfoContext(ctx, "Retrieving art project by revision", "artProjectID", artProjectID, "revisionID", revisionID, "userID", userID)

	rev, err := aps.artProjectRepo.GetRevisionByID(ctx, revisionID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get revision", "error", err, "revisionID", revisionID)
		return nil, nil, err
	}

	if rev.ArtProjectID.String() != artProjectID {
		slog.ErrorContext(ctx, "Revision does not belong to the specified art project", "artProjectID", artProjectID, "revisionID", revisionID)
		return nil, nil, fmt.Errorf("revision does not belong to the specified art project")
	}

	artProject, err := aps.artProjectRepo.GetArtProjectByID(ctx, rev.ArtProjectID.String())
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get art project", "error", err, "artProjectID", rev.ArtProjectID)
		return nil, nil, err
	}

	file, err := aps.storageRepo.GetRevision(ctx, userID, artProject.ID.String(), rev.Version)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get file from storage", "error", err, "revisionID", revisionID)
		return nil, nil, err
	}

	slog.InfoContext(ctx, "Successfully retrieved art project by revision", "artProjectID", artProjectID, "revisionID", revisionID, "userID", userID)
	return file, artProject, nil
}

// GetArtProjectByID retrieves the latest revision of a specified art project
func (aps *ArtProjectRetrievalService) GetArtProjectByID(ctx context.Context, userID, artProjectID string) (*os.File, *models.ArtProject, error) {
	slog.InfoContext(ctx, "Retrieving art project by ID", "artProjectID", artProjectID, "userID", userID)

	artProject, err := aps.artProjectRepo.GetArtProjectByID(ctx, artProjectID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get art project", "error", err, "artProjectID", artProjectID)
		return nil, nil, err
	}

	rev, err := aps.artProjectRepo.GetRevisionByID(ctx, artProject.LatestRevisionID.String())
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get latest revision", "error", err, "revisionID", artProject.LatestRevisionID)
		return nil, nil, err
	}

	file, err := aps.storageRepo.GetRevision(ctx, userID, artProject.ID.String(), rev.Version)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get file from storage", "error", err, "revisionID", artProject.LatestRevisionID)
		return nil, nil, err
	}

	slog.InfoContext(ctx, "Successfully retrieved art project by ID", "artProjectID", artProjectID, "userID", userID)
	return file, artProject, nil
}

// Ensure ArtProjectRetrievalService implements ArtProjectRetrievalServiceInterface
var _ ArtProjectRetrievalServiceInterface = (*ArtProjectRetrievalService)(nil)
