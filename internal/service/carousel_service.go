package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"art-backend/internal/model"
	"art-backend/internal/repository"
)

var ErrInvalidCarouselItems = errors.New("invalid carousel items")
var ErrCarouselItemNotFound = errors.New("carousel item not found")

type CarouselService struct {
	repository *repository.CarouselRepository
}

func NewCarouselService(repository *repository.CarouselRepository) *CarouselService {
	return &CarouselService{repository: repository}
}

func (s *CarouselService) GetActive(ctx context.Context) ([]model.CarouselItem, error) {
	return s.repository.FindActive(ctx)
}

func (s *CarouselService) GetAll(ctx context.Context) ([]model.CarouselItem, error) {
	return s.repository.FindAll(ctx)
}

func (s *CarouselService) Create(ctx context.Context, request model.SaveCarouselItemRequest) (model.CarouselItem, error) {
	request = normalizeCarouselItem(request)
	if request.ImageURL == "" {
		return model.CarouselItem{}, ErrInvalidCarouselItems
	}

	return s.repository.Create(ctx, request)
}

func (s *CarouselService) Update(ctx context.Context, id int64, request model.SaveCarouselItemRequest) (model.CarouselItem, error) {
	request = normalizeCarouselItem(request)
	if request.ImageURL == "" {
		return model.CarouselItem{}, ErrInvalidCarouselItems
	}

	item, err := s.repository.Update(ctx, id, request)
	if errors.Is(err, sql.ErrNoRows) {
		return model.CarouselItem{}, ErrCarouselItemNotFound
	}

	return item, err
}

func (s *CarouselService) SetActive(ctx context.Context, id int64, isActive bool) (model.CarouselItem, error) {
	item, err := s.repository.SetActive(ctx, id, isActive)
	if errors.Is(err, sql.ErrNoRows) {
		return model.CarouselItem{}, ErrCarouselItemNotFound
	}

	return item, err
}

func (s *CarouselService) Delete(ctx context.Context, id int64) error {
	err := s.repository.Delete(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrCarouselItemNotFound
	}

	return err
}

func (s *CarouselService) ReplaceAll(ctx context.Context, request model.SaveCarouselItemsRequest) ([]model.CarouselItem, error) {
	items := make([]model.SaveCarouselItemRequest, 0, len(request.Items))
	for index, item := range request.Items {
		item = normalizeCarouselItem(item)

		if item.ImageURL == "" {
			continue
		}

		if item.SortOrder == 0 {
			item.SortOrder = index + 1
		}

		items = append(items, item)
	}

	if len(items) == 0 {
		return nil, ErrInvalidCarouselItems
	}

	return s.repository.ReplaceAll(ctx, items)
}

func normalizeCarouselItem(item model.SaveCarouselItemRequest) model.SaveCarouselItemRequest {
	item.Title = strings.TrimSpace(item.Title)
	item.Description = strings.TrimSpace(item.Description)
	item.ImageURL = strings.TrimSpace(item.ImageURL)
	item.BlobName = strings.TrimSpace(item.BlobName)

	return item
}
