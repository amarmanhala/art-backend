package repository

import (
	"context"
	"database/sql"

	"art-backend/internal/model"
)

type ProductSizeRepository struct {
	db *sql.DB
}

func NewProductSizeRepository(db *sql.DB) *ProductSizeRepository {
	return &ProductSizeRepository{db: db}
}

func (r *ProductSizeRepository) FindAll(ctx context.Context) ([]model.ProductSize, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, label, width_in, height_in, width_cm, height_cm, sort_order, created_at, updated_at
		FROM product_sizes
		ORDER BY sort_order, id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ProductSize, 0)
	for rows.Next() {
		var item model.ProductSize
		if err := rows.Scan(
			&item.ID,
			&item.Label,
			&item.WidthIn,
			&item.HeightIn,
			&item.WidthCM,
			&item.HeightCM,
			&item.SortOrder,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}
