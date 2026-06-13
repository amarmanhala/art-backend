package repository

import (
	"context"
	"database/sql"

	"art-backend/internal/model"
)

type CarouselRepository struct {
	db *sql.DB
}

func NewCarouselRepository(db *sql.DB) *CarouselRepository {
	return &CarouselRepository{db: db}
}

func (r *CarouselRepository) FindAll(ctx context.Context) ([]model.CarouselItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, image_url, blob_name, sort_order, is_active, created_at, updated_at
		FROM carousel_items
		ORDER BY sort_order, id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCarouselItems(rows)
}

func (r *CarouselRepository) FindActive(ctx context.Context) ([]model.CarouselItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, image_url, blob_name, sort_order, is_active, created_at, updated_at
		FROM carousel_items
		WHERE is_active = TRUE
		ORDER BY sort_order, id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCarouselItems(rows)
}

func (r *CarouselRepository) ReplaceAll(ctx context.Context, requests []model.SaveCarouselItemRequest) ([]model.CarouselItem, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM carousel_items`); err != nil {
		return nil, err
	}

	items := make([]model.CarouselItem, 0, len(requests))
	for index, request := range requests {
		isActive := true
		if request.IsActive != nil {
			isActive = *request.IsActive
		}

		sortOrder := request.SortOrder
		if sortOrder == 0 {
			sortOrder = index + 1
		}

		var item model.CarouselItem
		err := tx.QueryRowContext(ctx, `
			INSERT INTO carousel_items (title, description, image_url, blob_name, sort_order, is_active)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, title, description, image_url, blob_name, sort_order, is_active, created_at, updated_at
		`, request.Title, request.Description, request.ImageURL, request.BlobName, sortOrder, isActive).Scan(
			&item.ID,
			&item.Title,
			&item.Description,
			&item.ImageURL,
			&item.BlobName,
			&item.SortOrder,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *CarouselRepository) Create(ctx context.Context, request model.SaveCarouselItemRequest) (model.CarouselItem, error) {
	isActive := true
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	var item model.CarouselItem
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO carousel_items (title, description, image_url, blob_name, sort_order, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, title, description, image_url, blob_name, sort_order, is_active, created_at, updated_at
	`, request.Title, request.Description, request.ImageURL, request.BlobName, request.SortOrder, isActive).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.ImageURL,
		&item.BlobName,
		&item.SortOrder,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	return item, err
}

func (r *CarouselRepository) Update(ctx context.Context, id int64, request model.SaveCarouselItemRequest) (model.CarouselItem, error) {
	isActive := true
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	var item model.CarouselItem
	err := r.db.QueryRowContext(ctx, `
		UPDATE carousel_items
		SET title = $1,
			description = $2,
			image_url = $3,
			blob_name = $4,
			sort_order = $5,
			is_active = $6,
			updated_at = NOW()
		WHERE id = $7
		RETURNING id, title, description, image_url, blob_name, sort_order, is_active, created_at, updated_at
	`, request.Title, request.Description, request.ImageURL, request.BlobName, request.SortOrder, isActive, id).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.ImageURL,
		&item.BlobName,
		&item.SortOrder,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	return item, err
}

func (r *CarouselRepository) SetActive(ctx context.Context, id int64, isActive bool) (model.CarouselItem, error) {
	var item model.CarouselItem
	err := r.db.QueryRowContext(ctx, `
		UPDATE carousel_items
		SET is_active = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, title, description, image_url, blob_name, sort_order, is_active, created_at, updated_at
	`, isActive, id).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.ImageURL,
		&item.BlobName,
		&item.SortOrder,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	return item, err
}

func (r *CarouselRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM carousel_items WHERE id = $1`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func scanCarouselItems(rows *sql.Rows) ([]model.CarouselItem, error) {
	items := make([]model.CarouselItem, 0)
	for rows.Next() {
		var item model.CarouselItem
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Description,
			&item.ImageURL,
			&item.BlobName,
			&item.SortOrder,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
