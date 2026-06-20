package repository

import (
	"context"
	"database/sql"
	"strings"

	"art-backend/internal/model"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, request model.CreateOrderRequest) (model.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Order{}, err
	}
	defer tx.Rollback()

	var order model.Order
	err = tx.QueryRowContext(ctx, `
		INSERT INTO orders (
			order_number, user_id, status, payment_status, currency,
			subtotal, tax_amount, shipping_amount, total_amount
		)
		VALUES ($1, $2, 'pending', 'unpaid', $3, $4, $5, $6, $7)
		RETURNING id, order_number, user_id, status, payment_status, currency,
			subtotal, tax_amount, shipping_amount, total_amount, created_at, updated_at
	`, request.OrderNumber, request.UserID, request.Currency, request.Subtotal, request.TaxAmount, request.ShippingAmount, request.TotalAmount).Scan(
		&order.ID,
		&order.OrderNumber,
		&order.UserID,
		&order.Status,
		&order.PaymentStatus,
		&order.Currency,
		&order.Subtotal,
		&order.TaxAmount,
		&order.ShippingAmount,
		&order.TotalAmount,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return model.Order{}, err
	}

	for _, item := range request.Items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO order_items (
				order_id, product_id, product_variant_id, product_title, product_slug,
				variant_size, unit_price, quantity, subtotal, image_url, thumbnail_url
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`,
			order.ID,
			item.ProductID,
			item.ProductVariantID,
			item.ProductTitle,
			item.ProductSlug,
			item.VariantSize,
			item.UnitPrice,
			item.Quantity,
			item.Subtotal,
			item.ImageURL,
			item.ThumbnailURL,
		); err != nil {
			return model.Order{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return model.Order{}, err
	}

	return r.FindByID(ctx, request.UserID, order.ID)
}

func (r *OrderRepository) FindAllByUser(ctx context.Context, userID int64) ([]model.OrderSummary, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, order_number, status, payment_status, currency,
			subtotal, tax_amount, shipping_amount, total_amount, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.OrderSummary, 0)
	for rows.Next() {
		var order model.OrderSummary
		if err := rows.Scan(
			&order.ID,
			&order.OrderNumber,
			&order.Status,
			&order.PaymentStatus,
			&order.Currency,
			&order.Subtotal,
			&order.TaxAmount,
			&order.ShippingAmount,
			&order.TotalAmount,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, order)
	}

	return items, rows.Err()
}

func (r *OrderRepository) FindByID(ctx context.Context, userID int64, id int64) (model.Order, error) {
	var order model.Order
	var stripeSessionID sql.NullString
	var stripePaymentIntentID sql.NullString
	var customerEmail sql.NullString
	var customerName sql.NullString
	var shippingName sql.NullString
	var shippingPhone sql.NullString
	var shippingLine1 sql.NullString
	var shippingLine2 sql.NullString
	var shippingCity sql.NullString
	var shippingState sql.NullString
	var shippingPostalCode sql.NullString
	var shippingCountry sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, order_number, user_id, status, payment_status, currency,
			subtotal, tax_amount, shipping_amount, total_amount, stripe_session_id,
			stripe_payment_intent_id, customer_email, customer_name, shipping_name,
			shipping_phone, shipping_line1, shipping_line2, shipping_city,
			shipping_state, shipping_postal_code, shipping_country, created_at, updated_at
		FROM orders
		WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(
		&order.ID,
		&order.OrderNumber,
		&order.UserID,
		&order.Status,
		&order.PaymentStatus,
		&order.Currency,
		&order.Subtotal,
		&order.TaxAmount,
		&order.ShippingAmount,
		&order.TotalAmount,
		&stripeSessionID,
		&stripePaymentIntentID,
		&customerEmail,
		&customerName,
		&shippingName,
		&shippingPhone,
		&shippingLine1,
		&shippingLine2,
		&shippingCity,
		&shippingState,
		&shippingPostalCode,
		&shippingCountry,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return model.Order{}, err
	}

	order.StripeSessionID = nullStringToPointer(stripeSessionID)
	order.StripePaymentIntentID = nullStringToPointer(stripePaymentIntentID)
	order.CustomerEmail = nullStringToPointer(customerEmail)
	order.CustomerName = nullStringToPointer(customerName)
	order.ShippingName = nullStringToPointer(shippingName)
	order.ShippingPhone = nullStringToPointer(shippingPhone)
	order.ShippingLine1 = nullStringToPointer(shippingLine1)
	order.ShippingLine2 = nullStringToPointer(shippingLine2)
	order.ShippingCity = nullStringToPointer(shippingCity)
	order.ShippingState = nullStringToPointer(shippingState)
	order.ShippingPostalCode = nullStringToPointer(shippingPostalCode)
	order.ShippingCountry = nullStringToPointer(shippingCountry)

	items, err := r.FindItems(ctx, id)
	if err != nil {
		return model.Order{}, err
	}
	order.Items = items

	return order, nil
}

func (r *OrderRepository) FindByStripeSessionID(ctx context.Context, stripeSessionID string) (model.Order, error) {
	return r.findOne(ctx, `
		SELECT id, order_number, user_id, status, payment_status, currency,
			subtotal, tax_amount, shipping_amount, total_amount, stripe_session_id,
			stripe_payment_intent_id, customer_email, customer_name, shipping_name,
			shipping_phone, shipping_line1, shipping_line2, shipping_city,
			shipping_state, shipping_postal_code, shipping_country, created_at, updated_at
		FROM orders
		WHERE stripe_session_id = $1
	`, stripeSessionID)
}

func (r *OrderRepository) FindByStripeSessionIDForUser(ctx context.Context, userID int64, stripeSessionID string) (model.Order, error) {
	return r.findOne(ctx, `
		SELECT id, order_number, user_id, status, payment_status, currency,
			subtotal, tax_amount, shipping_amount, total_amount, stripe_session_id,
			stripe_payment_intent_id, customer_email, customer_name, shipping_name,
			shipping_phone, shipping_line1, shipping_line2, shipping_city,
			shipping_state, shipping_postal_code, shipping_country, created_at, updated_at
		FROM orders
		WHERE user_id = $1 AND stripe_session_id = $2
	`, userID, stripeSessionID)
}

func (r *OrderRepository) FindByOrderNumberAndCustomerEmail(ctx context.Context, orderNumber string, email string) (model.Order, error) {
	return r.findOne(ctx, `
		SELECT id, order_number, user_id, status, payment_status, currency,
			subtotal, tax_amount, shipping_amount, total_amount, stripe_session_id,
			stripe_payment_intent_id, customer_email, customer_name, shipping_name,
			shipping_phone, shipping_line1, shipping_line2, shipping_city,
			shipping_state, shipping_postal_code, shipping_country, created_at, updated_at
		FROM orders
		WHERE order_number = $1 AND LOWER(customer_email) = $2
	`, orderNumber, strings.ToLower(email))
}

func (r *OrderRepository) findOne(ctx context.Context, query string, args ...any) (model.Order, error) {
	var order model.Order
	var sessionID sql.NullString
	var paymentIntentID sql.NullString
	var customerEmail sql.NullString
	var customerName sql.NullString
	var shippingName sql.NullString
	var shippingPhone sql.NullString
	var shippingLine1 sql.NullString
	var shippingLine2 sql.NullString
	var shippingCity sql.NullString
	var shippingState sql.NullString
	var shippingPostalCode sql.NullString
	var shippingCountry sql.NullString

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&order.ID,
		&order.OrderNumber,
		&order.UserID,
		&order.Status,
		&order.PaymentStatus,
		&order.Currency,
		&order.Subtotal,
		&order.TaxAmount,
		&order.ShippingAmount,
		&order.TotalAmount,
		&sessionID,
		&paymentIntentID,
		&customerEmail,
		&customerName,
		&shippingName,
		&shippingPhone,
		&shippingLine1,
		&shippingLine2,
		&shippingCity,
		&shippingState,
		&shippingPostalCode,
		&shippingCountry,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return model.Order{}, err
	}

	order.StripeSessionID = nullStringToPointer(sessionID)
	order.StripePaymentIntentID = nullStringToPointer(paymentIntentID)
	order.CustomerEmail = nullStringToPointer(customerEmail)
	order.CustomerName = nullStringToPointer(customerName)
	order.ShippingName = nullStringToPointer(shippingName)
	order.ShippingPhone = nullStringToPointer(shippingPhone)
	order.ShippingLine1 = nullStringToPointer(shippingLine1)
	order.ShippingLine2 = nullStringToPointer(shippingLine2)
	order.ShippingCity = nullStringToPointer(shippingCity)
	order.ShippingState = nullStringToPointer(shippingState)
	order.ShippingPostalCode = nullStringToPointer(shippingPostalCode)
	order.ShippingCountry = nullStringToPointer(shippingCountry)

	items, err := r.FindItems(ctx, order.ID)
	if err != nil {
		return model.Order{}, err
	}
	order.Items = items

	return order, nil
}

func (r *OrderRepository) FindItems(ctx context.Context, orderID int64) ([]model.OrderItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, order_id, product_id, product_variant_id, product_title,
			product_slug, variant_size, unit_price, quantity, subtotal,
			image_url, thumbnail_url, created_at, updated_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY id
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.OrderItem, 0)
	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.ProductVariantID,
			&item.ProductTitle,
			&item.ProductSlug,
			&item.VariantSize,
			&item.UnitPrice,
			&item.Quantity,
			&item.Subtotal,
			&item.ImageURL,
			&item.ThumbnailURL,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *OrderRepository) UpdateStripeSession(ctx context.Context, orderID int64, stripeSessionID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE orders
		SET stripe_session_id = $1,
			updated_at = NOW()
		WHERE id = $2
	`, stripeSessionID, orderID)
	return err
}

func (r *OrderRepository) MarkPaidByStripeSessionID(ctx context.Context, stripeSessionID string, paymentIntentID string, paymentStatus string, subtotal float64, taxAmount float64, shippingAmount float64, totalAmount float64, customerEmail *string, customerName *string, shippingName *string, shippingPhone *string, shippingLine1 *string, shippingLine2 *string, shippingCity *string, shippingState *string, shippingPostalCode *string, shippingCountry *string) (bool, error) {
	var paymentIntentArg any
	if paymentIntentID != "" {
		paymentIntentArg = paymentIntentID
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE orders
		SET status = 'paid',
			payment_status = $1,
			stripe_payment_intent_id = $2,
			subtotal = $3,
			tax_amount = $4,
			shipping_amount = $5,
			total_amount = $6,
			customer_email = $7,
			customer_name = $8,
			shipping_name = $9,
			shipping_phone = $10,
			shipping_line1 = $11,
			shipping_line2 = $12,
			shipping_city = $13,
			shipping_state = $14,
			shipping_postal_code = $15,
			shipping_country = $16,
			updated_at = NOW()
		WHERE stripe_session_id = $17 AND status <> 'paid'
	`, paymentStatus, paymentIntentArg, subtotal, taxAmount, shippingAmount, totalAmount, toNullArg(customerEmail), toNullArg(customerName), toNullArg(shippingName), toNullArg(shippingPhone), toNullArg(shippingLine1), toNullArg(shippingLine2), toNullArg(shippingCity), toNullArg(shippingState), toNullArg(shippingPostalCode), toNullArg(shippingCountry), stripeSessionID)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

func (r *OrderRepository) MarkExpiredByStripeSessionID(ctx context.Context, stripeSessionID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE orders
		SET status = 'expired',
			payment_status = 'expired',
			updated_at = NOW()
		WHERE stripe_session_id = $1 AND status <> 'paid'
	`, stripeSessionID)
	return err
}

func (r *OrderRepository) MarkCheckoutFailed(ctx context.Context, orderID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE orders
		SET status = 'checkout_failed',
			payment_status = 'failed',
			updated_at = NOW()
		WHERE id = $1
	`, orderID)
	return err
}

func nullStringToPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}

func toNullArg(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
