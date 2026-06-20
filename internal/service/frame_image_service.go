package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"art-backend/internal/model"
	"art-backend/internal/repository"
)

const MaxFrameImages = 10

var ErrInvalidFrameImages = errors.New("invalid frame images")

type FrameImageService struct {
	repository *repository.FrameRepository
	storage    FrameImageStorage
}

func NewFrameImageService(repository *repository.FrameRepository, storage FrameImageStorage) *FrameImageService {
	return &FrameImageService{repository: repository, storage: storage}
}

func (s *FrameImageService) ReplaceImages(ctx context.Context, frameID int64, request model.SaveFrameImagesRequest) (model.Frame, error) {
	requestImages := normalizeFrameImageRequests(request)
	if frameID <= 0 || len(requestImages) == 0 || len(requestImages) > MaxFrameImages {
		return model.Frame{}, ErrInvalidFrameImages
	}

	images := make([]model.FrameImage, 0, len(requestImages))
	for index, requestImage := range requestImages {
		blobName := strings.TrimSpace(requestImage.BlobName)
		if blobName == "" || !strings.HasPrefix(blobName, "frames/originals/") {
			return model.Frame{}, ErrInvalidFrameImages
		}

		asset, err := s.storage.CreateThumbnail(ctx, blobName)
		if err != nil {
			return model.Frame{}, err
		}

		sortOrder := requestImage.SortOrder
		if sortOrder <= 0 {
			sortOrder = index + 1
		}

		images = append(images, model.FrameImage{
			ImageURL:          asset.ImageURL,
			ThumbnailURL:      asset.ThumbnailURL,
			OriginalURL:       asset.OriginalURL,
			BlobName:          asset.BlobName,
			ThumbnailBlobName: asset.ThumbnailBlobName,
			AltText:           strings.TrimSpace(requestImage.AltText),
			SortOrder:         sortOrder,
			IsPrimary:         index == 0,
		})
	}

	frame, err := s.repository.ReplaceImages(ctx, frameID, images)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Frame{}, ErrFrameNotFound
	}

	return frame, err
}

func normalizeFrameImageRequests(request model.SaveFrameImagesRequest) []model.SaveFrameImageRequest {
	if strings.TrimSpace(request.MainImage.BlobName) != "" {
		images := make([]model.SaveFrameImageRequest, 0, 1+len(request.GalleryImages))
		mainImage := request.MainImage
		mainImage.IsPrimary = true
		if mainImage.SortOrder <= 0 {
			mainImage.SortOrder = 1
		}
		images = append(images, mainImage)

		for index, galleryImage := range request.GalleryImages {
			galleryImage.IsPrimary = false
			if galleryImage.SortOrder <= 0 {
				galleryImage.SortOrder = index + 2
			}
			images = append(images, galleryImage)
		}

		return images
	}

	return request.Images
}
