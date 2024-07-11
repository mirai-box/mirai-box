package service

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/repos"
)

// ArtProjectManagementService implements the ArtProjectManagementServiceInterface
type ArtProjectManagementService struct {
	artProjectRepo repos.ArtProjectRepositoryInterface
	storageRepo    repos.StorageRepositoryInterface
	stashRepo      repos.StashRepositoryInterface
}

// NewArtProjectManagementService creates a new instance of ArtProjectManagementService
func NewArtProjectManagementService(
	artProjectRepo repos.ArtProjectRepositoryInterface,
	storageRepo repos.StorageRepositoryInterface,
	stashRepo repos.StashRepositoryInterface,
) ArtProjectManagementServiceInterface {
	return &ArtProjectManagementService{
		artProjectRepo: artProjectRepo,
		storageRepo:    storageRepo,
		stashRepo:      stashRepo,
	}
}

// CreateArtProjectAndRevision creates a new art project along with its first revision
func (aps *ArtProjectManagementService) CreateArtProjectAndRevision(ctx context.Context, userID string, fileData io.Reader, title, filename string) (*models.ArtProject, error) {
	slog.InfoContext(ctx, "Creating new art project and revision",
		"userID", userID,
		"title", title,
		"filename", filename,
	)

	buffer := make([]byte, 512)
	n, err := fileData.Read(buffer)
	if err != nil && err != io.EOF {
		slog.ErrorContext(ctx, "Failed to detect content type", "error", err, "filename", filename)
		return nil, err
	}

	contentType := http.DetectContentType(buffer)
	fileData = io.MultiReader(bytes.NewReader(buffer[:n]), fileData)
	revisionID := uuid.New()

	slog.InfoContext(ctx, "New ArtProject content type is", "contentType", contentType, "revisionID", revisionID)

	stash, err := aps.stashRepo.FindByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find stash for user", "error", err, "userID", userID)
		return nil, err
	}

	artProject := &models.ArtProject{
		ID:               uuid.New(),
		StashID:          stash.ID,
		UserID:           stash.UserID,
		Title:            title,
		CreatedAt:        time.Now(),
		ContentType:      contentType,
		Filename:         filename,
		LatestRevisionID: revisionID,
	}

	revision := &models.Revision{
		ID:           revisionID,
		ArtProjectID: artProject.ID,
		Version:      1,
		CreatedAt:    time.Now(),
	}

	filePath, fileInfo, err := aps.storageRepo.SaveRevision(ctx, fileData, userID, artProject.ID.String(), revision.Version)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to store file", "error", err, "userID", userID)
		return nil, err
	}
	revision.FilePath = filePath
	revision.Size = fileInfo.Size()

	slog.InfoContext(ctx, "Art Revison is stored", "FilePath", filePath, "revisionID", revisionID)

	if err := aps.artProjectRepo.SaveArtProjectAndRevision(ctx, artProject, revision); err != nil {
		slog.ErrorContext(ctx, "Failed to store art project and revision info", "error", err, "artProjectID", artProject.ID)
		return nil, err
	}

	stash.Files++
	stash.ArtProjects++
	stash.UsedSpace += revision.Size
	if err := aps.stashRepo.Update(ctx, stash); err != nil {
		slog.ErrorContext(ctx, "Failed to update stash stats", "error", err, "artProjectID", artProject.ID)
		return nil, err
	}

	slog.InfoContext(ctx, "Art project and revision created successfully", "artProjectID", artProject.ID)
	return artProject, nil
}

// AddRevision adds a new revision to an existing art project
func (aps *ArtProjectManagementService) AddRevision(ctx context.Context, userID, artProjectID string, fileData io.Reader, comment, filename string) (*models.Revision, error) {
	slog.InfoContext(ctx, "Adding new revision", "userID", userID, "artProjectID", artProjectID, "filename", filename)

	artProject, err := aps.artProjectRepo.GetArtProjectByID(ctx, artProjectID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find art-project for user", "error", err, "artProjectID", artProjectID)
		return nil, err
	}

	revisionID := uuid.New()

	revision := &models.Revision{
		ID:           revisionID,
		ArtProjectID: artProject.ID,
		Version:      aps.determineNextVersion(ctx, artProjectID),
		CreatedAt:    time.Now(),
		Comment:      comment,
	}

	stash, err := aps.stashRepo.FindByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find stash for user", "error", err, "userID", userID)
		return nil, err
	}

	filePath, fileInfo, err := aps.storageRepo.SaveRevision(ctx, fileData, userID, artProjectID, revision.Version)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to store new revision file", "error", err)
		return nil, err
	}
	revision.FilePath = filePath
	revision.Size = fileInfo.Size()

	if err := aps.artProjectRepo.SaveRevision(ctx, revision); err != nil {
		slog.ErrorContext(ctx, "Failed to save revision info", "error", err, "revisionID", revision.ID)
		return nil, err
	}

	if err := aps.artProjectRepo.UpdateLatestRevision(ctx, artProjectID, revision.ID); err != nil {
		slog.ErrorContext(ctx, "Failed to update latest revision info", "error", err, "revisionID", revision.ID)
		return nil, err
	}

	stash.Files++
	stash.UsedSpace += revision.Size
	if err := aps.stashRepo.Update(ctx, stash); err != nil {
		slog.ErrorContext(ctx, "Failed to update stash stats", "error", err, "artProjectID", artProjectID)
		return nil, err
	}

	slog.InfoContext(ctx, "Revision added successfully", "revisionID", revision.ID, "artProjectID", artProjectID)
	return revision, nil
}

// ListLatestRevisions lists the latest revisions of all art projects
func (aps *ArtProjectManagementService) ListLatestRevisions(ctx context.Context, userID string) ([]models.Revision, error) {
	slog.InfoContext(ctx, "Listing latest revisions")

	revisions, err := aps.artProjectRepo.ListLatestRevisions(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list latest revisions", "error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "Latest revisions listed successfully", "count", len(revisions))
	return revisions, nil
}

// ListAllArtProjects lists all art projects
func (aps *ArtProjectManagementService) ListAllArtProjects(ctx context.Context, userID string) ([]models.ArtProject, error) {
	slog.InfoContext(ctx, "Listing all art projects")

	artProjects, err := aps.artProjectRepo.ListAllArtProjects(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list all art projects", "error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "Art projects listed successfully", "count", len(artProjects))
	return artProjects, nil
}

// ListAllRevisions lists all revisions of a specific art project
func (aps *ArtProjectManagementService) ListAllRevisions(ctx context.Context, artProjectID string) ([]models.Revision, error) {
	slog.InfoContext(ctx, "Listing all revisions", "artProjectID", artProjectID)

	revisions, err := aps.artProjectRepo.ListAllRevisions(ctx, artProjectID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list all revisions", "error", err, "artProjectID", artProjectID)
		return nil, err
	}

	slog.InfoContext(ctx, "Revisions listed successfully", "artProjectID", artProjectID, "count", len(revisions))
	return revisions, nil
}

// determineNextVersion determines the next version number for a new revision of an art project
func (aps *ArtProjectManagementService) determineNextVersion(ctx context.Context, artProjectID string) int {
	maxVersion, err := aps.artProjectRepo.GetMaxRevisionVersion(ctx, artProjectID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to retrieve maximum revision version", "error", err, "artProjectID", artProjectID)
		return 1 // Default to version 1 in case of error
	}
	return maxVersion + 1
}

// Ensure ArtProjectManagementService implements ArtProjectManagementServiceInterface
var _ ArtProjectManagementServiceInterface = (*ArtProjectManagementService)(nil)
