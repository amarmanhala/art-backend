package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"art-backend/internal/model"
	"art-backend/internal/repository"

	"github.com/lib/pq"
)

var (
	ErrProductNotFound        = errors.New("product not found")
	ErrInvalidProduct         = errors.New("invalid product")
	ErrInvalidProductVariants = errors.New("invalid product variants")
	ErrDuplicateProductSlug   = errors.New("duplicate product slug")
	ErrInvalidPage            = errors.New("invalid page")
	ErrInvalidSearch          = errors.New("invalid search")
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
	request = normalizeCreateProductRequest(request)
	if !prepareCreateProductVariants(&request) {
		return model.Product{}, ErrInvalidProduct
	}

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

	product, err := s.repository.Create(ctx, request)
	if isUniqueProductSlugError(err) {
		return model.Product{}, ErrDuplicateProductSlug
	}

	return product, err
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
	if isUniqueProductSlugError(err) {
		return model.Product{}, ErrDuplicateProductSlug
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
	if isUniqueProductSlugError(err) {
		return model.Product{}, ErrDuplicateProductSlug
	}

	return product, err
}

func isUniqueProductSlugError(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == "products_slug_key"
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

	if len(request.Variants) > 0 {
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

func (s *ProductService) ReplaceVariants(ctx context.Context, id int64, request model.SaveProductVariantsRequest) (model.Product, error) {
	if id <= 0 {
		return model.Product{}, ErrProductNotFound
	}

	variants := normalizeProductVariants(request.Variants)
	if len(variants) == 0 {
		return model.Product{}, ErrInvalidProductVariants
	}
	if !ensureDefaultVariant(&variants) {
		return model.Product{}, ErrInvalidProductVariants
	}

	product, err := s.repository.ReplaceVariants(ctx, id, variants)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Product{}, ErrProductNotFound
	}

	return product, err
}

func normalizeCreateProductRequest(request model.CreateProductRequest) model.CreateProductRequest {
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
	request.Variants = normalizeProductVariants(request.Variants)
	return request
}

func normalizeProductVariants(variants []model.SaveProductVariantRequest) []model.SaveProductVariantRequest {
	normalized := make([]model.SaveProductVariantRequest, 0, len(variants))
	for _, variant := range variants {
		variant.Size = strings.TrimSpace(variant.Size)
		if variant.Size == "" {
			continue
		}
		normalized = append(normalized, variant)
	}

	return normalized
}

func ensureDefaultVariant(variants *[]model.SaveProductVariantRequest) bool {
	if len(*variants) == 0 {
		return false
	}

	defaultIndex := -1
	for i, variant := range *variants {
		if variant.IsDefault {
			defaultIndex = i
			break
		}
	}
	if defaultIndex < 0 {
		defaultIndex = 0
		(*variants)[0].IsDefault = true
	}

	if (*variants)[defaultIndex].Size == "" || (*variants)[defaultIndex].Price <= 0 {
		return false
	}

	for i := range *variants {
		if (*variants)[i].Price <= 0 {
			return false
		}
	}

	return true
}

func prepareCreateProductVariants(request *model.CreateProductRequest) bool {
	request.Variants = normalizeProductVariants(request.Variants)

	if len(request.Variants) == 0 {
		if strings.TrimSpace(request.Size) == "" || request.Price <= 0 {
			return false
		}

		request.Variants = []model.SaveProductVariantRequest{{
			Size:          strings.TrimSpace(request.Size),
			Price:         request.Price,
			StockQuantity: request.StockQuantity,
			IsDefault:     true,
		}}
		request.StockQuantity = request.StockQuantity
		return true
	}

	if !ensureDefaultVariant(&request.Variants) {
		return false
	}

	defaultVariant := request.Variants[0]
	totalStock := 0
	for _, variant := range request.Variants {
		totalStock += variant.StockQuantity
		if variant.IsDefault {
			defaultVariant = variant
		}
	}

	request.Size = defaultVariant.Size
	request.Price = defaultVariant.Price
	request.StockQuantity = totalStock
	return true
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
