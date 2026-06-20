package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"art-backend/internal/model"
)

type FrameRepository struct {
	db *sql.DB
}

func NewFrameRepository(db *sql.DB) *FrameRepository {
	return &FrameRepository{db: db}
}

const frameColumns = `
	id, vendor_name, frame_name, color, description, article_number, product_detail,
	material_description, care, price,
	picture_width_cm, picture_width_in, picture_height_cm, picture_height_in,
	frame_width_cm, frame_width_in, frame_height_cm, frame_height_in,
	frame_depth_cm, frame_depth_in,
	created_at, updated_at
`

const frameImageColumns = `
	id, frame_id, image_url, thumbnail_url, original_url, blob_name, thumbnail_blob_name,
	alt_text, sort_order, is_primary, created_at, updated_at
`

func (r *FrameRepository) FindAll(ctx context.Context) ([]model.Frame, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+frameColumns+`
		FROM frames
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanFramesWithImages(ctx, rows)
}

func (r *FrameRepository) FindByID(ctx context.Context, id int64) (model.Frame, error) {
	var frame model.Frame
	err := r.db.QueryRowContext(ctx, `
		SELECT `+frameColumns+`
		FROM frames
		WHERE id = $1
	`, id).Scan(
		&frame.ID,
		&frame.VendorName,
		&frame.FrameName,
		&frame.Color,
		&frame.Description,
		&frame.ArticleNumber,
		&frame.ProductDetail,
		&frame.MaterialDescription,
		&frame.Care,
		&frame.Price,
		&frame.Measurements.PictureWidthCm,
		&frame.Measurements.PictureWidthIn,
		&frame.Measurements.PictureHeightCm,
		&frame.Measurements.PictureHeightIn,
		&frame.Measurements.FrameWidthCm,
		&frame.Measurements.FrameWidthIn,
		&frame.Measurements.FrameHeightCm,
		&frame.Measurements.FrameHeightIn,
		&frame.Measurements.FrameDepthCm,
		&frame.Measurements.FrameDepthIn,
		&frame.CreatedAt,
		&frame.UpdatedAt,
	)
	if err != nil {
		return frame, err
	}

	return r.attachImagesToFrame(ctx, frame)
}

func (r *FrameRepository) Create(ctx context.Context, request model.SaveFrameRequest) (model.Frame, error) {
	var frame model.Frame
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO frames (
			vendor_name, frame_name, color, description, article_number, product_detail,
			material_description, care, price,
			picture_width_cm, picture_width_in, picture_height_cm, picture_height_in,
			frame_width_cm, frame_width_in, frame_height_cm, frame_height_in,
			frame_depth_cm, frame_depth_in
		)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9,
			$10, $11, $12, $13,
			$14, $15, $16, $17,
			$18, $19
		)
		RETURNING `+frameColumns+`
	`,
		request.VendorName,
		request.FrameName,
		request.Color,
		request.Description,
		request.ArticleNumber,
		request.ProductDetail,
		request.MaterialDescription,
		request.Care,
		request.Price,
		request.Measurements.PictureWidthCm,
		request.Measurements.PictureWidthIn,
		request.Measurements.PictureHeightCm,
		request.Measurements.PictureHeightIn,
		request.Measurements.FrameWidthCm,
		request.Measurements.FrameWidthIn,
		request.Measurements.FrameHeightCm,
		request.Measurements.FrameHeightIn,
		request.Measurements.FrameDepthCm,
		request.Measurements.FrameDepthIn,
	).Scan(
		&frame.ID,
		&frame.VendorName,
		&frame.FrameName,
		&frame.Color,
		&frame.Description,
		&frame.ArticleNumber,
		&frame.ProductDetail,
		&frame.MaterialDescription,
		&frame.Care,
		&frame.Price,
		&frame.Measurements.PictureWidthCm,
		&frame.Measurements.PictureWidthIn,
		&frame.Measurements.PictureHeightCm,
		&frame.Measurements.PictureHeightIn,
		&frame.Measurements.FrameWidthCm,
		&frame.Measurements.FrameWidthIn,
		&frame.Measurements.FrameHeightCm,
		&frame.Measurements.FrameHeightIn,
		&frame.Measurements.FrameDepthCm,
		&frame.Measurements.FrameDepthIn,
		&frame.CreatedAt,
		&frame.UpdatedAt,
	)
	if err != nil {
		return frame, err
	}

	return frame, nil
}

func (r *FrameRepository) Update(ctx context.Context, id int64, request model.SaveFrameRequest) (model.Frame, error) {
	var frame model.Frame
	err := r.db.QueryRowContext(ctx, `
		UPDATE frames
		SET vendor_name = $1,
			frame_name = $2,
			color = $3,
			description = $4,
			article_number = $5,
			product_detail = $6,
			material_description = $7,
			care = $8,
			price = $9,
			picture_width_cm = $10,
			picture_width_in = $11,
			picture_height_cm = $12,
			picture_height_in = $13,
			frame_width_cm = $14,
			frame_width_in = $15,
			frame_height_cm = $16,
			frame_height_in = $17,
			frame_depth_cm = $18,
			frame_depth_in = $19,
			updated_at = NOW()
		WHERE id = $20
		RETURNING `+frameColumns+`
	`,
		request.VendorName,
		request.FrameName,
		request.Color,
		request.Description,
		request.ArticleNumber,
		request.ProductDetail,
		request.MaterialDescription,
		request.Care,
		request.Price,
		request.Measurements.PictureWidthCm,
		request.Measurements.PictureWidthIn,
		request.Measurements.PictureHeightCm,
		request.Measurements.PictureHeightIn,
		request.Measurements.FrameWidthCm,
		request.Measurements.FrameWidthIn,
		request.Measurements.FrameHeightCm,
		request.Measurements.FrameHeightIn,
		request.Measurements.FrameDepthCm,
		request.Measurements.FrameDepthIn,
		id,
	).Scan(
		&frame.ID,
		&frame.VendorName,
		&frame.FrameName,
		&frame.Color,
		&frame.Description,
		&frame.ArticleNumber,
		&frame.ProductDetail,
		&frame.MaterialDescription,
		&frame.Care,
		&frame.Price,
		&frame.Measurements.PictureWidthCm,
		&frame.Measurements.PictureWidthIn,
		&frame.Measurements.PictureHeightCm,
		&frame.Measurements.PictureHeightIn,
		&frame.Measurements.FrameWidthCm,
		&frame.Measurements.FrameWidthIn,
		&frame.Measurements.FrameHeightCm,
		&frame.Measurements.FrameHeightIn,
		&frame.Measurements.FrameDepthCm,
		&frame.Measurements.FrameDepthIn,
		&frame.CreatedAt,
		&frame.UpdatedAt,
	)
	if err != nil {
		return frame, err
	}

	return r.attachImagesToFrame(ctx, frame)
}

func (r *FrameRepository) Delete(ctx context.Context, id int64) (bool, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM frames
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

func (r *FrameRepository) ReplaceImages(ctx context.Context, frameID int64, images []model.FrameImage) (model.Frame, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Frame{}, err
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM frames WHERE id = $1)`, frameID).Scan(&exists); err != nil {
		return model.Frame{}, err
	}
	if !exists {
		return model.Frame{}, sql.ErrNoRows
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM frame_images WHERE frame_id = $1`, frameID); err != nil {
		return model.Frame{}, err
	}

	savedImages := make([]model.FrameImage, 0, len(images))
	for _, image := range images {
		var saved model.FrameImage
		err := tx.QueryRowContext(ctx, `
			INSERT INTO frame_images (
				frame_id, image_url, thumbnail_url, original_url, blob_name,
				thumbnail_blob_name, alt_text, sort_order, is_primary
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING `+frameImageColumns+`
		`,
			frameID,
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
			&saved.FrameID,
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
			return model.Frame{}, err
		}

		savedImages = append(savedImages, saved)
	}

	var frame model.Frame
	err = tx.QueryRowContext(ctx, `
		SELECT `+frameColumns+`
		FROM frames
		WHERE id = $1
	`, frameID).Scan(
		&frame.ID,
		&frame.VendorName,
		&frame.FrameName,
		&frame.Color,
		&frame.Description,
		&frame.ArticleNumber,
		&frame.ProductDetail,
		&frame.MaterialDescription,
		&frame.Care,
		&frame.Price,
		&frame.Measurements.PictureWidthCm,
		&frame.Measurements.PictureWidthIn,
		&frame.Measurements.PictureHeightCm,
		&frame.Measurements.PictureHeightIn,
		&frame.Measurements.FrameWidthCm,
		&frame.Measurements.FrameWidthIn,
		&frame.Measurements.FrameHeightCm,
		&frame.Measurements.FrameHeightIn,
		&frame.Measurements.FrameDepthCm,
		&frame.Measurements.FrameDepthIn,
		&frame.CreatedAt,
		&frame.UpdatedAt,
	)
	if err != nil {
		return model.Frame{}, err
	}
	frame.Images = savedImages

	if err := tx.Commit(); err != nil {
		return model.Frame{}, err
	}

	return frame, nil
}

func (r *FrameRepository) attachImagesToFrame(ctx context.Context, frame model.Frame) (model.Frame, error) {
	frames, err := r.attachImages(ctx, []model.Frame{frame})
	if err != nil {
		return model.Frame{}, err
	}
	if len(frames) == 0 {
		return model.Frame{}, nil
	}

	return frames[0], nil
}

func (r *FrameRepository) attachImages(ctx context.Context, frames []model.Frame) ([]model.Frame, error) {
	if len(frames) == 0 {
		return frames, nil
	}

	ids := make([]string, 0, len(frames))
	args := make([]any, 0, len(frames))
	frameIndexByID := make(map[int64]int, len(frames))
	for index, frame := range frames {
		args = append(args, frame.ID)
		ids = append(ids, fmt.Sprintf("$%d", len(args)))
		frameIndexByID[frame.ID] = index
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT `+frameImageColumns+`
		FROM frame_images
		WHERE frame_id IN (`+strings.Join(ids, ", ")+`)
		ORDER BY frame_id, sort_order, id
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var image model.FrameImage
		if err := rows.Scan(
			&image.ID,
			&image.FrameID,
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

		if index, ok := frameIndexByID[image.FrameID]; ok {
			frames[index].Images = append(frames[index].Images, image)
		}
	}

	return frames, rows.Err()
}

func (r *FrameRepository) scanFramesWithImages(ctx context.Context, rows *sql.Rows) ([]model.Frame, error) {
	frames := make([]model.Frame, 0)
	for rows.Next() {
		var frame model.Frame
		if err := rows.Scan(
			&frame.ID,
			&frame.VendorName,
			&frame.FrameName,
			&frame.Color,
			&frame.Description,
			&frame.ArticleNumber,
			&frame.ProductDetail,
			&frame.MaterialDescription,
			&frame.Care,
			&frame.Price,
			&frame.Measurements.PictureWidthCm,
			&frame.Measurements.PictureWidthIn,
			&frame.Measurements.PictureHeightCm,
			&frame.Measurements.PictureHeightIn,
			&frame.Measurements.FrameWidthCm,
			&frame.Measurements.FrameWidthIn,
			&frame.Measurements.FrameHeightCm,
			&frame.Measurements.FrameHeightIn,
			&frame.Measurements.FrameDepthCm,
			&frame.Measurements.FrameDepthIn,
			&frame.CreatedAt,
			&frame.UpdatedAt,
		); err != nil {
			return nil, err
		}

		frames = append(frames, frame)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return r.attachImages(ctx, frames)
}
