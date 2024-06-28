package service

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/mirai-box/mirai-box/internal/repository"
	"github.com/mirai-box/mirai-box/internal/model"
)

type pictureRetrievalService struct {
	pictureRepo repository.PictureRepository
	storageRepo repository.StorageRepository
}

func NewPictureRetrievalService(pictureRepo repository.PictureRepository, storageRepo repository.StorageRepository) PictureRetrievalService {
	return &pictureRetrievalService{
		pictureRepo: pictureRepo,
		storageRepo: storageRepo,
	}
}

func (ps *pictureRetrievalService) GetSharedPicture(artID string) (*os.File, *model.Picture, error) {
	rev, err := ps.pictureRepo.GetRevisionByArtID(artID)
	if err != nil {
		slog.Error("service: could not get revision for art id", "error", err, "artID", artID)
		return nil, nil, err
	}

	pic, err := ps.pictureRepo.GetPictureByID(rev.PictureID)
	if err != nil {
		slog.Error("could not get picture for id", "error", err, "pictureID", pic.ID)
		return nil, nil, err
	}

	filePath := getFilePath(pic.Filename, pic.ID, rev.Version)

	file, err := ps.storageRepo.GetPicture(filePath)
	if err != nil {
		slog.Error("could not get file from storage", "error", err, "picID", pic.ID, "path", filePath)
		return nil, nil, err
	}

	return file, pic, nil
}

func (ps *pictureRetrievalService) GetPictureByRevision(pictureID, revisionID string) (*os.File, *model.Picture, error) {
	rev, err := ps.pictureRepo.GetRevisionByID(revisionID)
	if err != nil {
		slog.Error("could not get revision for id", "error", err, "revisionID", revisionID)
		return nil, nil, err
	}

	if rev.PictureID != pictureID {
		slog.Error("this revisionID is not for this pictureID", "pictureID", pictureID, "revisionID", revisionID)
		return nil, nil, fmt.Errorf("can't get the revision")
	}

	pic, err := ps.pictureRepo.GetPictureByID(rev.PictureID)
	if err != nil {
		slog.Error("could not get picture for id", "error", err, "pictureID", pic.ID)
		return nil, nil, err
	}

	filePath := getFilePath(pic.Filename, pic.ID, rev.Version)

	file, err := ps.storageRepo.GetPicture(filePath)
	if err != nil {
		slog.Error("could not get file from storage", "error", err, "revisionID", revisionID, "path", filePath)
		return nil, nil, err
	}

	return file, pic, nil
}

func (ps *pictureRetrievalService) GetPictureByID(pictureID string) (*os.File, *model.Picture, error) {
	pic, err := ps.pictureRepo.GetPictureByID(pictureID)
	if err != nil {
		slog.Error("could not get picture for id", "error", err, "pictureID", pictureID)
		return nil, nil, err
	}

	rev, err := ps.pictureRepo.GetRevisionByID(pic.LatestRevisionID)
	if err != nil {
		slog.Error("could not get revision for id", "error", err, "revisionID", pic.LatestRevisionID)
		return nil, nil, err
	}

	filePath := getFilePath(pic.Filename, pic.ID, rev.Version)

	file, err := ps.storageRepo.GetPicture(filePath)
	if err != nil {
		slog.Error("could not get file from storage", "error", err, "revisionID", pic.LatestRevisionID, "path", filePath)
		return nil, nil, err
	}

	return file, pic, nil
}
