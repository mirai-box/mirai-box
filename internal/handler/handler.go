package handler

import (
	"github.com/mirai-box/mirai-box/internal/service"
)

type PictureManagementHandler struct {
	service service.PictureManagementService
}

type PictureRetrievalHandler struct {
	service service.PictureRetrievalService
}

type GalleryHandler struct {
	service service.GalleryService
}

func NewPictureManagementHandler(svc service.PictureManagementService) *PictureManagementHandler {
	return &PictureManagementHandler{
		service: svc,
	}
}

func NewPictureRetrievalHandler(svc service.PictureRetrievalService) *PictureRetrievalHandler {
	return &PictureRetrievalHandler{
		service: svc,
	}
}

func NewGalleryHandler(service service.GalleryService) *GalleryHandler {
	return &GalleryHandler{service: service}
}