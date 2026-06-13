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
	request.OriginalURL = strings.TrimSpace(request.OriginalURL)

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

func (s *ProductService) UpdateByID(ctx context.Context, id int64, request model.UpdateProductRequest) (model.Product, error) {
	if id <= 0 {
		return model.Product{}, ErrProductNotFound
	}
	if !sanitizeUpdateProductRequest(&request) {
		return model.Product{}, ErrInvalidProduct
	}

	product, err := s.repository.UpdateByID(ctx, id, request)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Product{}, ErrProductNotFound
	}

	return product, err
}

func (s *ProductService) UpdateBySlug(ctx context.Context, slug string, request model.UpdateProductRequest) (model.Product, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return model.Product{}, ErrProductNotFound
	}
	if !sanitizeUpdateProductRequest(&request) {
		return model.Product{}, ErrInvalidProduct
	}

	product, err := s.repository.UpdateBySlug(ctx, slug, request)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Product{}, ErrProductNotFound
	}

	return product, err
}

func sanitizeUpdateProductRequest(request *model.UpdateProductRequest) bool {
	hasField := false

	trimString := func(value **string) bool {
		if *value == nil {
			return true
		}

		hasField = true
		trimmed := strings.TrimSpace(**value)
		*value = &trimmed
		return true
	}

	if !trimString(&request.Title) ||
		!trimString(&request.Slug) ||
		!trimString(&request.Description) ||
		!trimString(&request.Currency) ||
		!trimString(&request.Category) ||
		!trimString(&request.Style) ||
		!trimString(&request.Theme) ||
		!trimString(&request.Orientation) ||
		!trimString(&request.Size) ||
		!trimString(&request.ImageURL) ||
		!trimString(&request.ThumbnailURL) ||
		!trimString(&request.OriginalURL) {
		return false
	}

	if request.Price != nil {
		hasField = true
		if *request.Price <= 0 {
			return false
		}
	}
	if request.StockQuantity != nil {
		hasField = true
		if *request.StockQuantity < 0 {
			return false
		}
	}
	if request.IsActive != nil {
		hasField = true
	}

	if request.Title != nil && *request.Title == "" {
		return false
	}
	if request.Slug != nil && *request.Slug == "" {
		return false
	}
	if request.Category != nil && *request.Category == "" {
		return false
	}

	return hasField
}

func (s *ProductService) DeleteByID(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrProductNotFound
	}

	deleted, err := s.repository.DeleteByID(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrProductNotFound
	}

	return nil
}

func (s *ProductService) DeleteBySlug(ctx context.Context, slug string) error {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return ErrProductNotFound
	}

	deleted, err := s.repository.DeleteBySlug(ctx, slug)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrProductNotFound
	}

	return nil
}
