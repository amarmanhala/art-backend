package repository

import (
	"context"
	"database/sql"

	"art-backend/internal/model"

	"github.com/lib/pq"
)

type ArtStyleRepository struct {
	db *sql.DB
}

func NewArtStyleRepository(db *sql.DB) *ArtStyleRepository {
	return &ArtStyleRepository{db: db}
}

const artStyleColumns = `
	id, origin, style, description, tags, image_url, blob_name, created_at, updated_at
`

func (r *ArtStyleRepository) FindAll(ctx context.Context) ([]model.ArtStyle, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+artStyleColumns+`
		FROM art_styles
		ORDER BY origin, style, id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ArtStyle, 0)
	for rows.Next() {
		var item model.ArtStyle
		if err := rows.Scan(
			&item.ID,
			&item.Origin,
			&item.Style,
			&item.Description,
			pq.Array(&item.Tags),
			&item.ImageURL,
			&item.BlobName,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *ArtStyleRepository) FindByID(ctx context.Context, id int64) (model.ArtStyle, error) {
	var item model.ArtStyle
	err := r.db.QueryRowContext(ctx, `
		SELECT `+artStyleColumns+`
		FROM art_styles
		WHERE id = $1
	`, id).Scan(
		&item.ID,
		&item.Origin,
		&item.Style,
		&item.Description,
		pq.Array(&item.Tags),
		&item.ImageURL,
		&item.BlobName,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	return item, err
}

func (r *ArtStyleRepository) Create(ctx context.Context, request model.SaveArtStyleRequest) (model.ArtStyle, error) {
	var item model.ArtStyle
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO art_styles (origin, style, description, tags, image_url, blob_name)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING `+artStyleColumns+`
	`, request.Origin, request.Style, request.Description, pq.Array(request.Tags), request.ImageURL, request.BlobName).Scan(
		&item.ID,
		&item.Origin,
		&item.Style,
		&item.Description,
		pq.Array(&item.Tags),
		&item.ImageURL,
		&item.BlobName,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	return item, err
}

func (r *ArtStyleRepository) Update(ctx context.Context, id int64, request model.SaveArtStyleRequest) (model.ArtStyle, error) {
	var item model.ArtStyle
	err := r.db.QueryRowContext(ctx, `
		UPDATE art_styles
		SET origin = $1,
			style = $2,
			description = $3,
			tags = $4,
			image_url = $5,
			blob_name = $6,
			updated_at = NOW()
		WHERE id = $7
		RETURNING `+artStyleColumns+`
	`, request.Origin, request.Style, request.Description, pq.Array(request.Tags), request.ImageURL, request.BlobName, id).Scan(
		&item.ID,
		&item.Origin,
		&item.Style,
		&item.Description,
		pq.Array(&item.Tags),
		&item.ImageURL,
		&item.BlobName,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	return item, err
}

func (r *ArtStyleRepository) Delete(ctx context.Context, id int64) (bool, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM art_styles
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
