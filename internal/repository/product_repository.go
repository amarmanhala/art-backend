package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"art-backend/internal/model"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

const productColumns = `
	id, title, slug, description, price, currency, category, style, theme,
	orientation, size, image_url, thumbnail_url, stock_quantity, is_active,
	created_at, updated_at
`

func (r *ProductRepository) FindAll(ctx context.Context, filter model.ProductFilter, limit int, offset int) ([]model.Product, error) {
	where, args := productFilterWhere(filter)
	args = append(args, limit, offset)

	query := fmt.Sprintf(`
		SELECT %s
		FROM products
		%s
		ORDER BY id DESC
		LIMIT $%d OFFSET $%d
	`, productColumns, where, len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]model.Product, 0)
	for rows.Next() {
		var product model.Product
		err := rows.Scan(
			&product.ID,
			&product.Title,
			&product.Slug,
			&product.Description,
			&product.Price,
			&product.Currency,
			&product.Category,
			&product.Style,
			&product.Theme,
			&product.Orientation,
			&product.Size,
			&product.ImageURL,
			&product.ThumbnailURL,
			&product.StockQuantity,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	return products, rows.Err()
}

func (r *ProductRepository) Count(ctx context.Context, filter model.ProductFilter) (int64, error) {
	var total int64
	where, args := productFilterWhere(filter)

	query := `
		SELECT COUNT(*)
		FROM products
		` + where

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&total)

	return total, err
}

func (r *ProductRepository) FindBySlug(ctx context.Context, slug string) (model.Product, error) {
	var product model.Product

	err := r.db.QueryRowContext(ctx, `
		SELECT `+productColumns+`
		FROM products
		WHERE slug = $1 AND is_active = TRUE
	`, slug).Scan(
		&product.ID,
		&product.Title,
		&product.Slug,
		&product.Description,
		&product.Price,
		&product.Currency,
		&product.Category,
		&product.Style,
		&product.Theme,
		&product.Orientation,
		&product.Size,
		&product.ImageURL,
		&product.ThumbnailURL,
		&product.StockQuantity,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	return product, err
}

func (r *ProductRepository) FindFeatured(ctx context.Context, limit int) ([]model.Product, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+productColumns+`
		FROM products
		WHERE is_active = TRUE AND stock_quantity > 0
		ORDER BY id DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProducts(rows)
}

func (r *ProductRepository) Search(ctx context.Context, keyword string, limit int, offset int) ([]model.Product, error) {
	search := "%" + strings.ToLower(keyword) + "%"

	rows, err := r.db.QueryContext(ctx, `
		SELECT `+productColumns+`
		FROM products
		WHERE is_active = TRUE
		  AND (
			LOWER(title) LIKE $1
			OR LOWER(description) LIKE $1
			OR LOWER(category) LIKE $1
			OR LOWER(style) LIKE $1
			OR LOWER(theme) LIKE $1
		  )
		ORDER BY id DESC
		LIMIT $2 OFFSET $3
	`, search, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProducts(rows)
}

func (r *ProductRepository) CountSearch(ctx context.Context, keyword string) (int64, error) {
	var total int64
	search := "%" + strings.ToLower(keyword) + "%"

	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM products
		WHERE is_active = TRUE
		  AND (
			LOWER(title) LIKE $1
			OR LOWER(description) LIKE $1
			OR LOWER(category) LIKE $1
			OR LOWER(style) LIKE $1
			OR LOWER(theme) LIKE $1
		  )
	`, search).Scan(&total)

	return total, err
}

func (r *ProductRepository) FindDistinctValues(ctx context.Context, column string) ([]string, error) {
	allowedColumns := map[string]bool{
		"category": true,
		"style":    true,
		"theme":    true,
	}
	if !allowedColumns[column] {
		return nil, fmt.Errorf("invalid product column: %s", column)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT `+column+`
		FROM products
		WHERE is_active = TRUE AND `+column+` <> ''
		ORDER BY `+column+`
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := make([]string, 0)
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, rows.Err()
}

func (r *ProductRepository) FindByID(ctx context.Context, id int64) (model.Product, error) {
	var product model.Product

	err := r.db.QueryRowContext(ctx, `
		SELECT `+productColumns+`
		FROM products
		WHERE id = $1
	`, id).Scan(
		&product.ID,
		&product.Title,
		&product.Slug,
		&product.Description,
		&product.Price,
		&product.Currency,
		&product.Category,
		&product.Style,
		&product.Theme,
		&product.Orientation,
		&product.Size,
		&product.ImageURL,
		&product.ThumbnailURL,
		&product.StockQuantity,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	return product, err
}

func (r *ProductRepository) Create(ctx context.Context, request model.CreateProductRequest) (model.Product, error) {
	var product model.Product

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO products (
			title, slug, description, price, currency, category, style, theme,
			orientation, size, image_url, thumbnail_url, stock_quantity, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING `+productColumns+`
	`,
		request.Title,
		request.Slug,
		request.Description,
		request.Price,
		request.Currency,
		request.Category,
		request.Style,
		request.Theme,
		request.Orientation,
		request.Size,
		request.ImageURL,
		request.ThumbnailURL,
		request.StockQuantity,
		*request.IsActive,
	).Scan(
		&product.ID,
		&product.Title,
		&product.Slug,
		&product.Description,
		&product.Price,
		&product.Currency,
		&product.Category,
		&product.Style,
		&product.Theme,
		&product.Orientation,
		&product.Size,
		&product.ImageURL,
		&product.ThumbnailURL,
		&product.StockQuantity,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	return product, err
}

func productFilterWhere(filter model.ProductFilter) (string, []any) {
	conditions := []string{"is_active = TRUE"}
	args := make([]any, 0)

	addTextFilter := func(column string, value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}

		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf("%s = $%d", column, len(args)))
	}

	addTextFilter("category", filter.Category)
	addTextFilter("style", filter.Style)
	addTextFilter("theme", filter.Theme)
	addTextFilter("orientation", filter.Orientation)

	if filter.MinPrice != nil {
		args = append(args, *filter.MinPrice)
		conditions = append(conditions, fmt.Sprintf("price >= $%d", len(args)))
	}
	if filter.MaxPrice != nil {
		args = append(args, *filter.MaxPrice)
		conditions = append(conditions, fmt.Sprintf("price <= $%d", len(args)))
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func scanProducts(rows *sql.Rows) ([]model.Product, error) {
	products := make([]model.Product, 0)
	for rows.Next() {
		var product model.Product
		err := rows.Scan(
			&product.ID,
			&product.Title,
			&product.Slug,
			&product.Description,
			&product.Price,
			&product.Currency,
			&product.Category,
			&product.Style,
			&product.Theme,
			&product.Orientation,
			&product.Size,
			&product.ImageURL,
			&product.ThumbnailURL,
			&product.StockQuantity,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	return products, rows.Err()
}
