package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"art-backend/internal/model"
	"art-backend/internal/repository"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidProduct  = errors.New("invalid product")
	ErrInvalidPage     = errors.New("invalid page")
	ErrInvalidSearch   = errors.New("invalid search")
)

type ProductService struct {
	repository *repository.ProductRepository
}

func NewProductService(repository *repository.ProductRepository) *ProductService {
	return &ProductService{repository: repository}
}

func (s *ProductService) GetAll(ctx context.Context, filter model.ProductFilter, page int, size int) (model.Page[model.Product], error) {
	if page < 0 || size <= 0 || size > 100 {
		return model.Page[model.Product]{}, ErrInvalidPage
	}

	total, err := s.repository.Count(ctx, filter)
	if err != nil {
		return model.Page[model.Product]{}, err
	}

	offset := page * size
	products, err := s.repository.FindAll(ctx, filter, size, offset)
	if err != nil {
		return model.Page[model.Product]{}, err
	}

	return model.NewPage(products, page, size, total), nil
}

func (s *ProductService) GetBySlug(ctx context.Context, slug string) (model.Product, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return model.Product{}, ErrProductNotFound
	}

	product, err := s.repository.FindBySlug(ctx, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Product{}, ErrProductNotFound
	}

	return product, err
}

func (s *ProductService) GetFeatured(ctx context.Context) ([]model.Product, error) {
	return s.repository.FindFeatured(ctx, 8)
}

func (s *ProductService) Search(ctx context.Context, keyword string, page int, size int) (model.Page[model.Product], error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return model.Page[model.Product]{}, ErrInvalidSearch
	}
	if page < 0 || size <= 0 || size > 100 {
		return model.Page[model.Product]{}, ErrInvalidPage
	}

	total, err := s.repository.CountSearch(ctx, keyword)
	if err != nil {
		return model.Page[model.Product]{}, err
	}

	offset := page * size
	products, err := s.repository.Search(ctx, keyword, size, offset)
	if err != nil {
		return model.Page[model.Product]{}, err
	}

	return model.NewPage(products, page, size, total), nil
}

func (s *ProductService) GetCategories(ctx context.Context) ([]string, error) {
	return s.repository.FindDistinctValues(ctx, "category")
}

func (s *ProductService) GetStyles(ctx context.Context) ([]string, error) {
	return s.repository.FindDistinctValues(ctx, "style")
}

func (s *ProductService) GetThemes(ctx context.Context) ([]string, error) {
	return s.repository.FindDistinctValues(ctx, "theme")
}

func (s *ProductService) GetByID(ctx context.Context, id int64) (model.Product, error) {
	product, err := s.repository.FindByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Product{}, ErrProductNotFound
	}

	return product, err
}

func (s *ProductService) Create(ctx context.Context, request model.CreateProductRequest) (model.Product, error) {
	request.Title = strings.TrimSpace(request.Title)
	request.Slug = strings.TrimSpace(request.Slug)
	request.Description = strings.TrimSpace(request.Description)
	request.Currency = strings.TrimSpace(request.Currency)
	request.Category = strings.TrimSpace(request.Category)
	request.Style = strings.TrimSpace(request.Style)
	request.Theme = strings.TrimSpace(request.Theme)
	request.Orientation = strings.TrimSpace(request.Orientation)
	request.Size = strings.TrimSpace(request.Size)
	request.ImageURL = strings.TrimSpace(request.ImageURL)
	request.ThumbnailURL = strings.TrimSpace(request.ThumbnailURL)

	if request.Currency == "" {
		request.Currency = "USD"
	}
	if request.IsActive == nil {
		isActive := true
		request.IsActive = &isActive
	}

	if request.Title == "" || request.Slug == "" || request.Category == "" || request.Price <= 0 {
		return model.Product{}, ErrInvalidProduct
	}

	return s.repository.Create(ctx, request)
}
