package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/repo"
)

//go:generate go run github.com/vektra/mockery/v2@v2 --name=ArtProjectService --filename=art_project_service.go --output=../../mocks/
type ArtProjectService interface {
	CreateArtProject(ctx context.Context, artProject *model.ArtProject) error
	GetArtProject(ctx context.Context, id string) (*model.ArtProject, error)
	// UpdateArtProject(ctx context.Context, artProject *model.ArtProject) error
	DeleteArtProject(ctx context.Context, id string) error
	ListArtProjects(ctx context.Context, userID string) ([]model.ArtProject, error)
	AddRevision(ctx context.Context, revision *model.Revision, fileData io.Reader) error
	GetLatestRevision(ctx context.Context, artProjectID string) (*model.Revision, error)
	ListRevisions(ctx context.Context, artProjectID string) ([]model.Revision, error)
	FindByID(ctx context.Context, id string) (*model.ArtProject, error)
	FindByUserID(ctx context.Context, userID string) ([]model.ArtProject, error)
	GetRevisionByArtID(ctx context.Context, artID string) (*model.Revision, error)
	GetArtProjectByRevision(ctx context.Context, userID, artProjectID, revisionID string) (io.ReadCloser, *model.ArtProject, error)
}

// ArtProjectService implements the ArtProjectServiceInterface
type artProjectService struct {
	userRepo        repo.UserRepository
	artRepo         repo.ArtProjectRepository
	fileStorageRepo repo.FileStorageRepository
	secretKey       []byte
}

// NewArtProjectService creates a new instance of ArtProjectService
func NewArtProjectService(
	ur repo.UserRepository,
	ar repo.ArtProjectRepository,
	fs repo.FileStorageRepository,
	secretKey []byte,
) ArtProjectService {

	return &artProjectService{
		userRepo:        ur,
		artRepo:         ar,
		fileStorageRepo: fs,
		secretKey:       secretKey,
	}
}

// CreateArtProject creates a new art project
func (s *artProjectService) CreateArtProject(ctx context.Context, artProject *model.ArtProject) error {
	logger := slog.With("method", "CreateArtProject",
		"artProjectID", artProject.ID,
		"title", artProject.Title,
	)

	stash, err := s.userRepo.GetStashByUserID(ctx, artProject.UserID.String())
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find stash for user", "error", err, "userID", artProject.UserID.String())
		return err
	}

	artProject.StashID = stash.ID
	if err := s.artRepo.CreateArtProject(ctx, artProject); err != nil {
		logger.ErrorContext(ctx, "Failed to create art project", "error", err)
		return err
	}

	logger.InfoContext(ctx, "Art project created successfully")

	return nil
}

// GetArtProject finds an art project by its ID
func (s *artProjectService) GetArtProject(ctx context.Context, id string) (*model.ArtProject, error) {
	slog.InfoContext(ctx, "ArtProjectService: Finding art project by ID", "artProjectID", id)

	artProject, err := s.artRepo.FindArtProjectByID(ctx, id)
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
func (s *artProjectService) FindByStashID(ctx context.Context, stashID string) ([]model.ArtProject, error) {
	slog.InfoContext(ctx, "Finding art projects by stash ID", "stashID", stashID)

	artProjects, err := s.artRepo.FindByStashID(ctx, stashID)
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

func (s *artProjectService) FindByUserID(ctx context.Context, userID string) ([]model.ArtProject, error) {
	slog.InfoContext(ctx, "Finding art projects by user ID", "userID", userID)

	artProjects, err := s.artRepo.FindByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find art projects by user ID",
			"error", err,
			"userID", userID,
		)
		return nil, err
	}

	return artProjects, nil
}

func (s *artProjectService) AddRevision(ctx context.Context, revision *model.Revision, fileData io.Reader) error {
	logger := slog.With("method", "AddRevision", "userID", revision.UserID, "revisionID", revision.ID, "artProjectID", revision.ArtProjectID)

	if revision.UserID == uuid.Nil || revision.ArtProjectID == uuid.Nil || fileData == nil {
		logger.Warn("Invalid input parameters")
		return model.ErrInvalidInput
	}

	nextVersion := s.determineNextVersion(ctx, revision.ArtProjectID.String())

	filePath, fileInfo, err := s.fileStorageRepo.SaveRevisionFile(ctx, fileData, revision.UserID.String(), revision.ArtProjectID.String(), nextVersion)
	if err != nil {
		logger.Error("Failed to store revision file", "error", err)
		return fmt.Errorf("failed to store revision file: %w", err)
	}

	artID, err := GenerateArtID(revision.ID, revision.UserID, s.secretKey)
	if err != nil {
		logger.Error("Failed to generate artID", "error", err)
		return fmt.Errorf("failed to generate artID: %w", err)
	}

	revision.Version = nextVersion
	revision.FilePath = filePath
	revision.Size = fileInfo.Size()
	revision.ArtID = artID

	if err := s.artRepo.SaveRevision(ctx, revision); err != nil {
		logger.Error("Failed to save revision", "error", err)
		return fmt.Errorf("failed to save revision: %w", err)
	}

	if err := s.artRepo.UpdateLatestRevision(ctx, revision.ArtProjectID.String(), revision.ID); err != nil {
		logger.Error("Failed to update latest revision", "error", err)
		return fmt.Errorf("failed to update latest revision: %w", err)
	}

	stash, err := s.userRepo.GetStashByUserID(ctx, revision.UserID.String())
	if err != nil {
		logger.Error("Failed to find stash", "error", err)
		return fmt.Errorf("failed to find stash: %w", err)
	}

	stash.Files++
	stash.ArtProjects++
	stash.UsedSpace += revision.Size
	if err := s.userRepo.UpdateStash(ctx, stash); err != nil {
		logger.Error("Failed to update stash stats", "error", err)
		return fmt.Errorf("failed to update stash stats: %w", err)
	}

	logger.Info("Revision added successfully")
	return nil
}

func (s *artProjectService) GetRevisionByArtID(ctx context.Context, artID string) (*model.Revision, error) {
	revisionID, userID, err := DecodeArtID(artID, s.secretKey)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to decode artID", "error", err, "artID", artID)
		return nil, err
	}

	revision, err := s.artRepo.FindRevisionByID(ctx, revisionID)
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

func (s *artProjectService) GetLatestRevision(ctx context.Context, artProjectID string) (*model.Revision, error) {
	logger := slog.With("method", "GetLatestRevision", "artProjectID", artProjectID)

	artProject, err := s.artRepo.FindArtProjectByID(ctx, artProjectID)
	if err != nil {
		logger.Error("Failed to get art project", "error", err)
		return nil, err
	}

	revision, err := s.artRepo.GetRevisionByID(ctx, artProject.LatestRevisionID.String())
	if err != nil {
		logger.Error("Failed to get latest revision", "error", err)
		return nil, err
	}

	logger.Info("Latest revision retrieved successfully")
	return revision, nil
}

func (s *artProjectService) DeleteArtProject(ctx context.Context, id string) error {
	logger := slog.With("method", "DeleteArtProject", "artProjectID", id)

	if id == "" {
		logger.Warn("Invalid input parameters")
		return model.ErrInvalidInput
	}

	if err := s.artRepo.DeleteArtProject(ctx, id); err != nil {
		logger.Error("Failed to delete art project", "error", err)
		return err
	}

	logger.Info("Art project deleted successfully")
	return nil
}

func (s *artProjectService) ListArtProjects(ctx context.Context, userID string) ([]model.ArtProject, error) {
	logger := slog.With("method", "ListArtProjects", "userID", userID)

	if userID == "" {
		logger.Warn("Invalid input parameters")
		return nil, model.ErrInvalidInput
	}

	artProjects, err := s.artRepo.ListAllArtProjects(ctx, userID)
	if err != nil {
		logger.Error("Failed to list art projects", "error", err)
		return nil, err
	}

	if len(artProjects) == 0 {
		logger.Info("No art projects found")
		return []model.ArtProject{}, nil
	}

	logger.Info("Art projects listed successfully", "count", len(artProjects))
	return artProjects, nil
}

func (s *artProjectService) ListRevisions(ctx context.Context, artProjectID string) ([]model.Revision, error) {
	logger := slog.With("method", "ListRevisions", "artProjectID", artProjectID)

	if artProjectID == "" {
		logger.Warn("Invalid input: empty artProjectID")
		return nil, model.ErrInvalidInput
	}

	revisions, err := s.artRepo.ListAllRevisions(ctx, artProjectID)
	if err != nil {
		logger.Error("Failed to list revisions", "error", err)
		return nil, err
	}

	if len(revisions) == 0 {
		logger.Info("No revisions found for art project")
		return []model.Revision{}, nil
	}

	logger.Info("Revisions listed successfully", "count", len(revisions))
	return revisions, nil
}

func (s *artProjectService) FindByID(ctx context.Context, id string) (*model.ArtProject, error) {
	logger := slog.With("method", "FindByID", "artProjectID", id)

	if id == "" {
		logger.Warn("Invalid input: empty id")
		return nil, model.ErrInvalidInput
	}

	artProject, err := s.artRepo.FindArtProjectByID(ctx, id)
	if err != nil {
		logger.Error("Failed to find art project", "error", err)
		return nil, err
	}

	logger.Info("Art project found successfully")
	return artProject, nil
}

func (s *artProjectService) GetArtProjectByRevision(ctx context.Context, userID, artProjectID, revisionID string) (io.ReadCloser, *model.ArtProject, error) {
	logger := slog.With("service", "GetArtProjectByRevision", "artProjectID", artProjectID, "revisionID", revisionID, "userID", userID)

	rev, err := s.artRepo.FindRevisionByID(ctx, revisionID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get revision", "error", err)
		return nil, nil, err
	}

	if rev.ArtProjectID.String() != artProjectID {
		logger.ErrorContext(ctx, "Revision does not belong to the specified art project")
		return nil, nil, fmt.Errorf("revision does not belong to the specified art project")
	}

	artProject, err := s.artRepo.FindArtProjectByID(ctx, rev.ArtProjectID.String())
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get art project", "error", err)
		return nil, nil, err
	}

	file, err := s.fileStorageRepo.GetRevisionFile(ctx, userID, artProject.ID.String(), rev.Version)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get file from storage", "error", err)
		return nil, nil, err
	}

	logger.InfoContext(ctx, "Successfully retrieved art project by revision")
	return file, artProject, nil
}

// determineNextVersion determines the next version number for a new revision of an art project
func (s *artProjectService) determineNextVersion(ctx context.Context, artProjectID string) int {
	maxVersion, err := s.artRepo.GetMaxRevisionVersion(ctx, artProjectID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to retrieve maximum revision version", "error", err, "artProjectID", artProjectID)
		return 1 // Default to version 1 in case of error
	}
	return maxVersion + 1
}
