package repository

import (
	"context"
	"database/sql"

	"art-backend/internal/model"
)

type CartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) GetOrCreate(ctx context.Context, userID int64) (model.Cart, error) {
	var cart model.Cart

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO carts (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO UPDATE
		SET updated_at = carts.updated_at
		RETURNING id, user_id, created_at, updated_at
	`, userID).Scan(
		&cart.ID,
		&cart.UserID,
		&cart.CreatedAt,
		&cart.UpdatedAt,
	)

	return cart, err
}

func (r *CartRepository) FindItems(ctx context.Context, cartID int64) ([]model.CartItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			ci.id,
			ci.cart_id,
			ci.quantity,
			(ci.quantity * p.price) AS subtotal,
			ci.created_at,
			ci.updated_at,
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
			p.stock_quantity,
			p.is_active,
			p.created_at,
			p.updated_at
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1
		ORDER BY ci.id DESC
	`, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.CartItem, 0)
	for rows.Next() {
		var item model.CartItem
		err := rows.Scan(
			&item.ID,
			&item.CartID,
			&item.Quantity,
			&item.Subtotal,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.Product.ID,
			&item.Product.Title,
			&item.Product.Slug,
			&item.Product.Description,
			&item.Product.Price,
			&item.Product.Currency,
			&item.Product.Category,
			&item.Product.Style,
			&item.Product.Theme,
			&item.Product.Orientation,
			&item.Product.Size,
			&item.Product.ImageURL,
			&item.Product.ThumbnailURL,
			&item.Product.StockQuantity,
			&item.Product.IsActive,
			&item.Product.CreatedAt,
			&item.Product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *CartRepository) AddItem(ctx context.Context, cartID int64, request model.AddCartItemRequest) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO cart_items (cart_id, product_id, quantity)
		SELECT $1, id, $2
		FROM products
		WHERE id = $3 AND is_active = TRUE
		ON CONFLICT (cart_id, product_id) DO UPDATE
		SET quantity = cart_items.quantity + EXCLUDED.quantity,
			updated_at = NOW()
	`, cartID, request.Quantity, request.ProductID)
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

	return r.touchCart(ctx, cartID)
}

func (r *CartRepository) UpdateItem(ctx context.Context, cartID int64, itemID int64, quantity int) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE cart_items
		SET quantity = $1, updated_at = NOW()
		WHERE id = $2 AND cart_id = $3
	`, quantity, itemID, cartID)
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

	return r.touchCart(ctx, cartID)
}

func (r *CartRepository) DeleteItem(ctx context.Context, cartID int64, itemID int64) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM cart_items
		WHERE id = $1 AND cart_id = $2
	`, itemID, cartID)
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

	return r.touchCart(ctx, cartID)
}

func (r *CartRepository) Clear(ctx context.Context, cartID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM cart_items
		WHERE cart_id = $1
	`, cartID)
	if err != nil {
		return err
	}

	return r.touchCart(ctx, cartID)
}

func (r *CartRepository) touchCart(ctx context.Context, cartID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE carts
		SET updated_at = NOW()
		WHERE id = $1
	`, cartID)

	return err
}
