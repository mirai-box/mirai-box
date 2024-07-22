package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/repo"
)

//go:generate go run github.com/vektra/mockery/v2@v2 --name=ArtLinkService --filename=art_link_service.go --output=../../mocks/

// ArtLinkService defines the interface for art link related operations
type ArtLinkService interface {
	CreateArtLink(ctx context.Context, revisionID uuid.UUID, duration time.Duration, oneTime bool) (string, error)
	GetArtLinkByToken(ctx context.Context, token string) (*model.ArtLink, error)
	UpdateArtLink(ctx context.Context, artLink *model.ArtLink) error
}

type artLinkService struct {
	artLinkRepo repo.ArtLinkRepository
}

// NewArtLinkService creates a new instance of ArtLinkService
func NewArtLinkService(artLinkRepo repo.ArtLinkRepository) ArtLinkService {
	return &artLinkService{
		artLinkRepo: artLinkRepo,
	}
}

// CreateArtLink creates a new art link
func (s *artLinkService) CreateArtLink(ctx context.Context, revisionID uuid.UUID, duration time.Duration, oneTime bool) (string, error) {
	logger := slog.With("method", "CreateArtLink", "revisionID", revisionID)

	if revisionID == uuid.Nil || duration <= 0 {
		logger.Warn("Invalid input parameters")
		return "", model.ErrInvalidInput
	}

	token := uuid.New().String()
	expiresAt := time.Now().Add(duration)

	artLink := &model.ArtLink{
		Token:      token,
		RevisionID: revisionID,
		ExpiresAt:  expiresAt,
		OneTime:    oneTime,
	}

	if err := s.artLinkRepo.CreateArtLink(ctx, artLink); err != nil {
		logger.Error("Failed to create art link", "error", err)
		return "", err
	}

	logger.Info("Art link created successfully", "token", token)
	return token, nil
}

// GetArtLinkByToken retrieves an art link by its token
func (s *artLinkService) GetArtLinkByToken(ctx context.Context, token string) (*model.ArtLink, error) {
	logger := slog.With("method", "GetArtLinkByToken", "token", token)

	if token == "" {
		logger.Warn("Invalid token")
		return nil, model.ErrInvalidInput
	}

	artLink, err := s.artLinkRepo.GetArtLinkByToken(ctx, token)
	if err != nil {
		if errors.Is(err, model.ErrArtLinkNotFound) {
			logger.Info("Art link not found")
			return nil, err
		}
		logger.Error("Failed to get art link", "error", err)
		return nil, err
	}

	logger.Info("Art link retrieved successfully")
	return artLink, nil
}

// UpdateArtLink updates an existing art link
func (s *artLinkService) UpdateArtLink(ctx context.Context, artLink *model.ArtLink) error {
	logger := slog.With("method", "UpdateArtLink", "token", artLink.Token)

	if artLink.Token == "" {
		logger.Warn("Invalid art link")
		return model.ErrInvalidInput
	}

	if err := s.artLinkRepo.UpdateArtLink(ctx, artLink); err != nil {
		if errors.Is(err, model.ErrArtLinkNotFound) {
			logger.Info("Art link not found")
			return err
		}
		logger.Error("Failed to update art link", "error", err)
		return err
	}

	logger.Info("Art link updated successfully")
	return nil
}
