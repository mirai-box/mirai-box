package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/repos"
)

// ArtProjectRetrievalService implements the ArtProjectRetrievalServiceInterface
type artProjectRetrievalService struct {
	artProjectRepo repos.ArtProjectRepositoryInterface
	storageRepo    repos.StorageRepositoryInterface
	secretKey      []byte
}

// NewArtProjectRetrievalService creates a new instance of ArtProjectRetrievalService
func NewArtProjectRetrievalService(
	artProjectRepo repos.ArtProjectRepositoryInterface,
	storageRepo repos.StorageRepositoryInterface,
	secretKey []byte,
) ArtProjectRetrievalServiceInterface {
	return &artProjectRetrievalService{
		artProjectRepo: artProjectRepo,
		storageRepo:    storageRepo,
		secretKey:      secretKey,
	}
}

func (s *artProjectRetrievalService) GetRevisionByArtID(ctx context.Context, artID string) (*models.Revision, error) {
	revisionID, userID, err := DecodeArtID(artID, s.secretKey)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to decode artID", "error", err, "artID", artID)
		return nil, err
	}

	revision, err := s.artProjectRepo.GetRevisionByID(ctx, revisionID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed get revision by ID", "error", err, "revisionID", revisionID)
		return nil, err
	}

	if revision.UserID.String() != userID {
		slog.ErrorContext(ctx, "userID does not match the revision's userID", "revision.UserID", revision.UserID, "userID", userID)
		return nil, errors.New("userID does not match the revision's userID")
	}

	return revision, nil
}

// CreateArtLink creates an art link with expiry and one-time use functionality
func (s *artProjectRetrievalService) CreateArtLink(ctx context.Context,
	revisionID uuid.UUID, duration time.Duration,
	oneTime bool, unlimited bool) (string, error) {

	// TODO: use GenerateArtID()
	token := "xxx"

	expiresAt := time.Now().Add(duration)
	if unlimited {
		expiresAt = time.Time{} // Zero value of time.Time to represent no expiration
	}

	artLink := models.ArtLink{
		Token:      token,
		RevisionID: revisionID,
		ExpiresAt:  expiresAt,
		OneTime:    oneTime,
	}

	if err := s.artProjectRepo.CreateArtLink(ctx, &artLink); err != nil {
		return "", err
	}

	return fmt.Sprintf("/art/%s", token), nil
}

// GetArtByToken retrieves the art using the provided token and validates it
func (s *artProjectRetrievalService) GetArtByToken(ctx context.Context, token string) (*models.Revision, error) {
	artLink, err := s.artProjectRepo.GetArtLinkByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Check for expiry
	if time.Now().After(artLink.ExpiresAt) {
		return nil, errors.New("link has expired")
	}

	// Check if one-time use and already used
	if artLink.OneTime && artLink.Used {
		return nil, errors.New("link has already been used")
	}

	// Mark as used if one-time use
	if artLink.OneTime {
		artLink.Used = true
		if err := s.artProjectRepo.UpdateArtLink(ctx, artLink); err != nil {
			return nil, err
		}
	}

	return &artLink.Revision, nil
}

// GetArtProjectByRevision retrieves a specific revision of an art project
func (s *artProjectRetrievalService) GetArtProjectByRevision(ctx context.Context, userID, artProjectID, revisionID string) (*os.File, *models.ArtProject, error) {
	slog.InfoContext(ctx, "Retrieving art project by revision", "artProjectID", artProjectID, "revisionID", revisionID, "userID", userID)

	rev, err := s.artProjectRepo.GetRevisionByID(ctx, revisionID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get revision", "error", err, "revisionID", revisionID)
		return nil, nil, err
	}

	if rev.ArtProjectID.String() != artProjectID {
		slog.ErrorContext(ctx, "Revision does not belong to the specified art project", "artProjectID", artProjectID, "revisionID", revisionID)
		return nil, nil, fmt.Errorf("revision does not belong to the specified art project")
	}

	artProject, err := s.artProjectRepo.GetArtProjectByID(ctx, rev.ArtProjectID.String())
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get art project", "error", err, "artProjectID", rev.ArtProjectID)
		return nil, nil, err
	}

	file, err := s.storageRepo.GetRevision(ctx, userID, artProject.ID.String(), rev.Version)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get file from storage", "error", err, "revisionID", revisionID)
		return nil, nil, err
	}

	slog.InfoContext(ctx, "Successfully retrieved art project by revision", "artProjectID", artProjectID, "revisionID", revisionID, "userID", userID)
	return file, artProject, nil
}

// GetArtProjectByID retrieves the latest revision of a specified art project
func (s *artProjectRetrievalService) GetArtProjectByID(ctx context.Context, userID, artProjectID string) (*os.File, *models.ArtProject, error) {
	slog.InfoContext(ctx, "Retrieving art project by ID", "artProjectID", artProjectID, "userID", userID)

	artProject, err := s.artProjectRepo.GetArtProjectByID(ctx, artProjectID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get art project", "error", err, "artProjectID", artProjectID)
		return nil, nil, err
	}

	rev, err := s.artProjectRepo.GetRevisionByID(ctx, artProject.LatestRevisionID.String())
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get latest revision", "error", err, "revisionID", artProject.LatestRevisionID)
		return nil, nil, err
	}

	file, err := s.storageRepo.GetRevision(ctx, userID, artProject.ID.String(), rev.Version)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get file from storage", "error", err, "revisionID", artProject.LatestRevisionID)
		return nil, nil, err
	}

	slog.InfoContext(ctx, "Successfully retrieved art project by ID", "artProjectID", artProjectID, "userID", userID)
	return file, artProject, nil
}

// Ensure ArtProjectRetrievalService implements ArtProjectRetrievalServiceInterface
var _ ArtProjectRetrievalServiceInterface = (*artProjectRetrievalService)(nil)
