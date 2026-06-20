package repository

import (
	"context"
	"database/sql"

	"art-backend/internal/model"
)

type LikedArtRepository struct {
	db *sql.DB
}

func NewLikedArtRepository(db *sql.DB) *LikedArtRepository {
	return &LikedArtRepository{db: db}
}

func (r *LikedArtRepository) FindLikedByUser(ctx context.Context, userID int64) ([]model.LikedArt, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			la.id,
			la.user_id,
			la.product_id,
			la.status,
			la.created_at,
			la.updated_at,
			p.id,
			p.title,
			p.slug,
			p.description,
			p.price,
			p.currency,
			p.category,
			p.style,
			p.theme,
			p.orientation,
			p.size,
			p.image_url,
			p.thumbnail_url,
			p.original_url,
			p.stock_quantity,
			p.is_active,
			p.created_at,
			p.updated_at
		FROM liked_arts la
		JOIN products p ON p.id = la.product_id
		WHERE la.user_id = $1
		  AND la.status = $2
		  AND p.is_active = TRUE
		ORDER BY la.updated_at DESC, la.id DESC
	`, userID, model.LikedArtStatusLiked)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	likedArts := make([]model.LikedArt, 0)
	products := make([]model.Product, 0)
	for rows.Next() {
		var likedArt model.LikedArt
		err := rows.Scan(
			&likedArt.ID,
			&likedArt.UserID,
			&likedArt.ProductID,
			&likedArt.Status,
			&likedArt.CreatedAt,
			&likedArt.UpdatedAt,
			&likedArt.Product.ID,
			&likedArt.Product.Title,
			&likedArt.Product.Slug,
			&likedArt.Product.Description,
			&likedArt.Product.Price,
			&likedArt.Product.Currency,
			&likedArt.Product.Category,
			&likedArt.Product.Style,
			&likedArt.Product.Theme,
			&likedArt.Product.Orientation,
			&likedArt.Product.Size,
			&likedArt.Product.ImageURL,
			&likedArt.Product.ThumbnailURL,
			&likedArt.Product.OriginalURL,
			&likedArt.Product.StockQuantity,
			&likedArt.Product.IsActive,
			&likedArt.Product.CreatedAt,
			&likedArt.Product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		products = append(products, likedArt.Product)
		likedArts = append(likedArts, likedArt)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	products, err = (&ProductRepository{db: r.db}).attachImages(ctx, products)
	if err != nil {
		return nil, err
	}
	products, err = (&ProductRepository{db: r.db}).attachVariants(ctx, products)
	if err != nil {
		return nil, err
	}
	for index := range likedArts {
		likedArts[index].Product = products[index]
	}

	return likedArts, nil
}

func (r *LikedArtRepository) SetStatus(ctx context.Context, userID int64, productID int64, status string) (model.LikedArt, error) {
	var likedArt model.LikedArt

	err := r.db.QueryRowContext(ctx, `
		WITH upsert AS (
			INSERT INTO liked_arts (user_id, product_id, status)
			SELECT $1, p.id, $3
			FROM products p
			WHERE p.id = $2 AND p.is_active = TRUE
			ON CONFLICT (user_id, product_id) DO UPDATE
			SET status = EXCLUDED.status,
				updated_at = NOW()
			RETURNING id, user_id, product_id, status, created_at, updated_at
		)
		SELECT
			upsert.id,
			upsert.user_id,
			upsert.product_id,
			upsert.status,
			upsert.created_at,
			upsert.updated_at,
			p.id,
			p.title,
			p.slug,
			p.description,
			p.price,
			p.currency,
			p.category,
			p.style,
			p.theme,
			p.orientation,
			p.size,
			p.image_url,
			p.thumbnail_url,
			p.original_url,
			p.stock_quantity,
			p.is_active,
			p.created_at,
			p.updated_at
		FROM upsert
		JOIN products p ON p.id = upsert.product_id
	`, userID, productID, status).Scan(
		&likedArt.ID,
		&likedArt.UserID,
		&likedArt.ProductID,
		&likedArt.Status,
		&likedArt.CreatedAt,
		&likedArt.UpdatedAt,
		&likedArt.Product.ID,
		&likedArt.Product.Title,
		&likedArt.Product.Slug,
		&likedArt.Product.Description,
		&likedArt.Product.Price,
		&likedArt.Product.Currency,
		&likedArt.Product.Category,
		&likedArt.Product.Style,
		&likedArt.Product.Theme,
		&likedArt.Product.Orientation,
		&likedArt.Product.Size,
		&likedArt.Product.ImageURL,
		&likedArt.Product.ThumbnailURL,
		&likedArt.Product.OriginalURL,
		&likedArt.Product.StockQuantity,
		&likedArt.Product.IsActive,
		&likedArt.Product.CreatedAt,
		&likedArt.Product.UpdatedAt,
	)
	if err != nil {
		return likedArt, err
	}

	product, err := (&ProductRepository{db: r.db}).attachImagesToProduct(ctx, likedArt.Product)
	if err != nil {
		return model.LikedArt{}, err
	}
	likedArt.Product = product

	return likedArt, nil
}
