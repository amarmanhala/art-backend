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
	orientation, size, image_url, thumbnail_url, original_url, stock_quantity, is_active,
	created_at, updated_at
`

const productImageColumns = `
	id, product_id, image_url, thumbnail_url, original_url, blob_name, thumbnail_blob_name,
	alt_text, sort_order, is_primary, created_at, updated_at
`

const productVariantColumns = `
	id, product_id, size, price, stock_quantity, is_default, created_at, updated_at
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
			&product.OriginalURL,
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	products, err = r.attachImages(ctx, products)
	if err != nil {
		return nil, err
	}

	return r.attachVariants(ctx, products)
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
		&product.OriginalURL,
		&product.StockQuantity,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		return product, err
	}

	return r.attachImagesToProduct(ctx, product)
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

	return r.scanProductsWithImages(ctx, rows)
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

	return r.scanProductsWithImages(ctx, rows)
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
		&product.OriginalURL,
		&product.StockQuantity,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		return product, err
	}

	return r.attachImagesToProduct(ctx, product)
}

func (r *ProductRepository) Create(ctx context.Context, request model.CreateProductRequest) (model.Product, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Product{}, err
	}
	defer tx.Rollback()

	var product model.Product
	err = tx.QueryRowContext(ctx, `
		INSERT INTO products (
			title, slug, description, price, currency, category, style, theme,
			orientation, size, image_url, thumbnail_url, original_url, stock_quantity, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
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
		request.OriginalURL,
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
		&product.OriginalURL,
		&product.StockQuantity,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return model.Product{}, err
	}

	variants := request.Variants
	if len(variants) == 0 {
		variants = []model.SaveProductVariantRequest{{
			Size:          request.Size,
			Price:         request.Price,
			StockQuantity: request.StockQuantity,
			IsDefault:     true,
		}}
	}
	if _, err := insertProductVariantsTx(ctx, tx, product.ID, variants); err != nil {
		return model.Product{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.Product{}, err
	}

	return r.attachImagesToProduct(ctx, product)
}

func (r *ProductRepository) UpdateByID(ctx context.Context, id int64, request model.UpdateProductRequest) (model.Product, error) {
	return r.update(ctx, "id = $1", []any{id}, request)
}

func (r *ProductRepository) UpdateBySlug(ctx context.Context, slug string, request model.UpdateProductRequest) (model.Product, error) {
	return r.update(ctx, "slug = $1", []any{slug}, request)
}

func (r *ProductRepository) update(ctx context.Context, where string, args []any, request model.UpdateProductRequest) (model.Product, error) {
	var product model.Product
	setClauses := make([]string, 0)

	addStringField := func(column string, value *string) {
		if value == nil {
			return
		}

		args = append(args, *value)
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", column, len(args)))
	}
	addFloatField := func(column string, value *float64) {
		if value == nil {
			return
		}

		args = append(args, *value)
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", column, len(args)))
	}
	addIntField := func(column string, value *int) {
		if value == nil {
			return
		}

		args = append(args, *value)
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", column, len(args)))
	}
	addBoolField := func(column string, value *bool) {
		if value == nil {
			return
		}

		args = append(args, *value)
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", column, len(args)))
	}

	addStringField("title", request.Title)
	addStringField("slug", request.Slug)
	addStringField("description", request.Description)
	addFloatField("price", request.Price)
	addStringField("currency", request.Currency)
	addStringField("category", request.Category)
	addStringField("style", request.Style)
	addStringField("theme", request.Theme)
	addStringField("orientation", request.Orientation)
	addStringField("size", request.Size)
	addStringField("image_url", request.ImageURL)
	addStringField("thumbnail_url", request.ThumbnailURL)
	addStringField("original_url", request.OriginalURL)
	addIntField("stock_quantity", request.StockQuantity)
	addBoolField("is_active", request.IsActive)

	if len(setClauses) == 0 {
		return product, nil
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	query := fmt.Sprintf(`
		UPDATE products
		SET %s
		WHERE %s
		RETURNING %s
	`, strings.Join(setClauses, ", "), where, productColumns)

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
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
		&product.OriginalURL,
		&product.StockQuantity,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return product, err
	}

	return r.attachImagesToProduct(ctx, product)
}

func (r *ProductRepository) DeleteByID(ctx context.Context, id int64) (bool, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM products
		WHERE id = $1
	`, id)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (r *ProductRepository) DeleteBySlug(ctx context.Context, slug string) (bool, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM products
		WHERE slug = $1
	`, slug)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (r *ProductRepository) ReplaceVariants(ctx context.Context, productID int64, variants []model.SaveProductVariantRequest) (model.Product, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Product{}, err
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM products WHERE id = $1)`, productID).Scan(&exists); err != nil {
		return model.Product{}, err
	}
	if !exists {
		return model.Product{}, sql.ErrNoRows
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM product_variants WHERE product_id = $1`, productID); err != nil {
		return model.Product{}, err
	}

	savedVariants, err := insertProductVariantsTx(ctx, tx, productID, variants)
	if err != nil {
		return model.Product{}, err
	}

	defaultVariant := savedVariants[0]
	totalStock := 0
	for _, variant := range savedVariants {
		totalStock += variant.StockQuantity
		if variant.IsDefault {
			defaultVariant = variant
		}
	}

	var product model.Product
	err = tx.QueryRowContext(ctx, `
		UPDATE products
		SET size = $1,
			price = $2,
			stock_quantity = $3,
			updated_at = NOW()
		WHERE id = $4
		RETURNING `+productColumns+`
	`, defaultVariant.Size, defaultVariant.Price, totalStock, productID).Scan(
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
		&product.OriginalURL,
		&product.StockQuantity,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return model.Product{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.Product{}, err
	}

	return r.attachImagesToProduct(ctx, product)
}

func (r *ProductRepository) ReplaceImages(ctx context.Context, productID int64, images []model.ProductImage) (model.Product, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Product{}, err
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM products WHERE id = $1)`, productID).Scan(&exists); err != nil {
		return model.Product{}, err
	}
	if !exists {
		return model.Product{}, sql.ErrNoRows
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM product_images WHERE product_id = $1`, productID); err != nil {
		return model.Product{}, err
	}

	savedImages := make([]model.ProductImage, 0, len(images))
	for _, image := range images {
		var saved model.ProductImage
		err := tx.QueryRowContext(ctx, `
			INSERT INTO product_images (
				product_id, image_url, thumbnail_url, original_url, blob_name,
				thumbnail_blob_name, alt_text, sort_order, is_primary
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING `+productImageColumns+`
		`,
			productID,
			image.ImageURL,
			image.ThumbnailURL,
			image.OriginalURL,
			image.BlobName,
			image.ThumbnailBlobName,
			image.AltText,
			image.SortOrder,
			image.IsPrimary,
		).Scan(
			&saved.ID,
			&saved.ProductID,
			&saved.ImageURL,
			&saved.ThumbnailURL,
			&saved.OriginalURL,
			&saved.BlobName,
			&saved.ThumbnailBlobName,
			&saved.AltText,
			&saved.SortOrder,
			&saved.IsPrimary,
			&saved.CreatedAt,
			&saved.UpdatedAt,
		)
		if err != nil {
			return model.Product{}, err
		}

		savedImages = append(savedImages, saved)
	}

	imageURL := ""
	thumbnailURL := ""
	originalURL := ""
	for _, image := range savedImages {
		if image.IsPrimary {
			imageURL = image.ImageURL
			thumbnailURL = image.ThumbnailURL
			originalURL = image.OriginalURL
			break
		}
	}

	var product model.Product
	err = tx.QueryRowContext(ctx, `
		UPDATE products
		SET image_url = $1,
			thumbnail_url = $2,
			original_url = $3,
			updated_at = NOW()
		WHERE id = $4
		RETURNING `+productColumns+`
	`, imageURL, thumbnailURL, originalURL, productID).Scan(
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
		&product.OriginalURL,
		&product.StockQuantity,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return model.Product{}, err
	}
	product.Images = savedImages

	if err := tx.Commit(); err != nil {
		return model.Product{}, err
	}

	return product, nil
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
			&product.OriginalURL,
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

func (r *ProductRepository) scanProductsWithImages(ctx context.Context, rows *sql.Rows) ([]model.Product, error) {
	products, err := scanProducts(rows)
	if err != nil {
		return nil, err
	}

	products, err = r.attachImages(ctx, products)
	if err != nil {
		return nil, err
	}

	return r.attachVariants(ctx, products)
}

func (r *ProductRepository) attachImagesToProduct(ctx context.Context, product model.Product) (model.Product, error) {
	products, err := r.attachImages(ctx, []model.Product{product})
	if err != nil {
		return model.Product{}, err
	}
	if len(products) == 0 {
		return product, nil
	}

	return r.attachVariantsToProduct(ctx, products[0])
}

func (r *ProductRepository) attachImages(ctx context.Context, products []model.Product) ([]model.Product, error) {
	if len(products) == 0 {
		return products, nil
	}

	ids := make([]string, 0, len(products))
	args := make([]any, 0, len(products))
	productIndexByID := make(map[int64]int, len(products))
	for index, product := range products {
		args = append(args, product.ID)
		ids = append(ids, fmt.Sprintf("$%d", len(args)))
		productIndexByID[product.ID] = index
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT `+productImageColumns+`
		FROM product_images
		WHERE product_id IN (`+strings.Join(ids, ", ")+`)
		ORDER BY product_id, sort_order, id
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var image model.ProductImage
		if err := rows.Scan(
			&image.ID,
			&image.ProductID,
			&image.ImageURL,
			&image.ThumbnailURL,
			&image.OriginalURL,
			&image.BlobName,
			&image.ThumbnailBlobName,
			&image.AltText,
			&image.SortOrder,
			&image.IsPrimary,
			&image.CreatedAt,
			&image.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if index, ok := productIndexByID[image.ProductID]; ok {
			products[index].Images = append(products[index].Images, image)
		}
	}

	return products, rows.Err()
}

func (r *ProductRepository) attachVariants(ctx context.Context, products []model.Product) ([]model.Product, error) {
	if len(products) == 0 {
		return products, nil
	}

	ids := make([]string, 0, len(products))
	args := make([]any, 0, len(products))
	productIndexByID := make(map[int64]int, len(products))
	for index, product := range products {
		args = append(args, product.ID)
		ids = append(ids, fmt.Sprintf("$%d", len(args)))
		productIndexByID[product.ID] = index
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT `+productVariantColumns+`
		FROM product_variants
		WHERE product_id IN (`+strings.Join(ids, ", ")+`)
		ORDER BY product_id, is_default DESC, id
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var variant model.ProductVariant
		if err := rows.Scan(
			&variant.ID,
			&variant.ProductID,
			&variant.Size,
			&variant.Price,
			&variant.StockQuantity,
			&variant.IsDefault,
			&variant.CreatedAt,
			&variant.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if index, ok := productIndexByID[variant.ProductID]; ok {
			products[index].Variants = append(products[index].Variants, variant)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for index := range products {
		if len(products[index].Variants) == 0 {
			continue
		}

		defaultVariant := products[index].Variants[0]
		totalStock := 0
		for _, variant := range products[index].Variants {
			totalStock += variant.StockQuantity
			if variant.IsDefault {
				defaultVariant = variant
			}
		}

		products[index].Size = defaultVariant.Size
		products[index].Price = defaultVariant.Price
		products[index].StockQuantity = totalStock
	}

	return products, nil
}

func (r *ProductRepository) attachVariantsToProduct(ctx context.Context, product model.Product) (model.Product, error) {
	products, err := r.attachVariants(ctx, []model.Product{product})
	if err != nil {
		return model.Product{}, err
	}
	if len(products) == 0 {
		return product, nil
	}

	return products[0], nil
}

func insertProductVariantsTx(ctx context.Context, tx *sql.Tx, productID int64, variants []model.SaveProductVariantRequest) ([]model.ProductVariant, error) {
	if len(variants) == 0 {
		return nil, sql.ErrNoRows
	}

	saved := make([]model.ProductVariant, 0, len(variants))
	for index, variant := range variants {
		if variant.Size == "" || variant.Price <= 0 {
			return nil, sql.ErrNoRows
		}

		isDefault := variant.IsDefault
		if index == 0 {
			isDefault = true
		}

		var item model.ProductVariant
		err := tx.QueryRowContext(ctx, `
			INSERT INTO product_variants (product_id, size, price, stock_quantity, is_default)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING `+productVariantColumns+`
		`, productID, variant.Size, variant.Price, variant.StockQuantity, isDefault).Scan(
			&item.ID,
			&item.ProductID,
			&item.Size,
			&item.Price,
			&item.StockQuantity,
			&item.IsDefault,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		saved = append(saved, item)
	}

	return saved, nil
}
