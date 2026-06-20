package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"art-backend/internal/model"
	"art-backend/internal/repository"
)

const MaxProductImages = 10

var ErrInvalidProductImages = errors.New("invalid product images")

type ProductImageService struct {
	repository *repository.ProductRepository
	storage    ProductImageStorage
}

func NewProductImageService(repository *repository.ProductRepository, storage ProductImageStorage) *ProductImageService {
	return &ProductImageService{repository: repository, storage: storage}
}

func (s *ProductImageService) ReplaceImages(ctx context.Context, productID int64, request model.SaveProductImagesRequest) (model.Product, error) {
	requestImages := normalizeProductImageRequests(request)
	if productID <= 0 || len(requestImages) == 0 || len(requestImages) > MaxProductImages {
		return model.Product{}, ErrInvalidProductImages
	}

	images := make([]model.ProductImage, 0, len(requestImages))
	for index, requestImage := range requestImages {
		blobName := strings.TrimSpace(requestImage.BlobName)
		if blobName == "" || !strings.HasPrefix(blobName, "products/originals/") {
			return model.Product{}, ErrInvalidProductImages
		}

		asset, err := s.storage.CreateThumbnail(ctx, blobName)
		if err != nil {
			return model.Product{}, err
		}

		sortOrder := requestImage.SortOrder
		if sortOrder <= 0 {
			sortOrder = index + 1
		}

		images = append(images, model.ProductImage{
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

	product, err := s.repository.ReplaceImages(ctx, productID, images)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Product{}, ErrProductNotFound
	}

	return product, err
}

func normalizeProductImageRequests(request model.SaveProductImagesRequest) []model.SaveProductImageRequest {
	if strings.TrimSpace(request.MainImage.BlobName) != "" {
		images := make([]model.SaveProductImageRequest, 0, 1+len(request.GalleryImages))
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
