package service

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/repository"
)

type pictureManagementService struct {
	pictureRepo repository.PictureRepository
	storageRepo repository.StorageRepository
}

func NewPictureManagementService(pictureRepo repository.PictureRepository, storageRepo repository.StorageRepository) PictureManagementService {
	return &pictureManagementService{
		pictureRepo: pictureRepo,
		storageRepo: storageRepo,
	}
}

func (ps *pictureManagementService) CreatePictureAndRevision(fileData io.Reader, title, filename string) (*model.Picture, error) {
	// Read the file data into a buffer to detect the content type
	buffer := make([]byte, 512) // Use the first 512 bytes to detect the content type
	n, err := fileData.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}

	// Detect the content type
	contentType := http.DetectContentType(buffer)

	// Since the reader has been read, we need to combine the buffer and the rest of the reader
	fileData = io.MultiReader(bytes.NewReader(buffer[:n]), fileData)
	revisionID := uuid.NewString()

	picture := &model.Picture{
		ID:               uuid.New().String(),
		Title:            title,
		CreatedAt:        time.Now(),
		ContentType:      contentType,
		Filename:         filename,
		LatestRevisionID: revisionID,
	}

	if err := ps.pictureRepo.SavePicture(picture); err != nil {
		slog.Error("could not store picture info", "error", err, "picture", picture)
		return nil, err
	}

	revision := &model.Revision{
		ID:        revisionID,
		PictureID: picture.ID,
		Version:   1,
		CreatedAt: time.Now(),
		ArtID:     getArtID(picture.ID, revisionID),
	}

	filePath, err := ps.safeFile(fileData, filename, picture.ID, revision.Version)
	if err != nil {
		slog.Error("could not store file", "error", err)
	}

	revision.FilePath = filePath
	if err := ps.pictureRepo.SaveRevision(revision); err != nil {
		slog.Error("could not store revision info", "error", err, "revision", revision)
		return nil, err
	}

	return picture, nil
}

func (ps *pictureManagementService) AddRevision(pictureID string, fileData io.Reader, comment, filename string) (*model.Revision, error) {
	revisionID := uuid.NewString()
	revision := &model.Revision{
		ID:        revisionID,
		PictureID: pictureID,
		Version:   ps.determineNextVersion(pictureID),
		CreatedAt: time.Now(),
		Comment:   comment,
		ArtID:     getArtID(pictureID, revisionID),
	}

	filePath, err := ps.safeFile(fileData, filename, pictureID, revision.Version)
	if err != nil {
		slog.Error("could not store new revision of the file", "error", err)
		return nil, err
	}
	revision.FilePath = filePath

	if err := ps.pictureRepo.SaveRevision(revision); err != nil {
		slog.Error("could not save revision info", "error", err, "revision", revision)
		return nil, err
	}

	if err := ps.pictureRepo.UpdateLatestRevision(pictureID, revision.ID); err != nil {
		slog.Error("could not update latest revision info", "error", err, "revision", revision)
		return nil, err
	}

	return revision, nil
}

func (ps *pictureManagementService) ListLatestRevisions() ([]model.Revision, error) {
	revisions, err := ps.pictureRepo.ListLatestRevisions()
	if err != nil {
		slog.Error("could not list latest revisions", "error", err)
		return nil, err
	}

	return revisions, nil
}

func (ps *pictureManagementService) ListAllPictures() ([]model.Picture, error) {
	pictures, err := ps.pictureRepo.ListAllPictures()
	if err != nil {
		slog.Error("could not list all pictures", "error", err)
		return nil, err
	}

	if len(pictures) == 0 {
		return []model.Picture{}, nil
	}

	return pictures, nil
}

func (ps *pictureManagementService) ListAllRevisions(pictureID string) ([]model.Revision, error) {
	revisions, err := ps.pictureRepo.ListAllRevisions(pictureID)
	if err != nil {
		slog.Error("could not list all pictures", "error", err)
		return nil, err
	}

	return revisions, nil
}

func (ps *pictureManagementService) determineNextVersion(pictureID string) int {
	maxVersion, err := ps.pictureRepo.GetMaxRevisionVersion(pictureID)
	if err != nil {
		slog.Error("Failed to retrieve maximum revision version", "pictureID", pictureID)
		return 1 // Default to version 1 in case of error
	}
	return maxVersion + 1
}

func (ps *pictureManagementService) safeFile(fileData io.Reader, filename, pictureID string, version int) (string, error) {
	filePath := getFilePath(filename, pictureID, version)

	if err := ps.storageRepo.SavePicture(fileData, filePath); err != nil {
		slog.Error("could not store file", "error", err, "path", fileData)
		return "", err
	}

	return filePath, nil
}
