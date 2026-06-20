package model

import "time"

type Order struct {
	ID                    int64       `json:"id"`
	OrderNumber           string      `json:"order_number"`
	UserID                int64       `json:"user_id"`
	Status                string      `json:"status"`
	PaymentStatus         string      `json:"payment_status"`
	Currency              string      `json:"currency"`
	Subtotal              float64     `json:"subtotal"`
	TaxAmount             float64     `json:"tax_amount"`
	ShippingAmount        float64     `json:"shipping_amount"`
	TotalAmount           float64     `json:"total_amount"`
	StripeSessionID       *string     `json:"stripe_session_id,omitempty"`
	StripePaymentIntentID *string     `json:"stripe_payment_intent_id,omitempty"`
	CustomerEmail         *string     `json:"customer_email,omitempty"`
	CustomerName          *string     `json:"customer_name,omitempty"`
	ShippingName          *string     `json:"shipping_name,omitempty"`
	ShippingPhone         *string     `json:"shipping_phone,omitempty"`
	ShippingLine1         *string     `json:"shipping_line1,omitempty"`
	ShippingLine2         *string     `json:"shipping_line2,omitempty"`
	ShippingCity          *string     `json:"shipping_city,omitempty"`
	ShippingState         *string     `json:"shipping_state,omitempty"`
	ShippingPostalCode    *string     `json:"shipping_postal_code,omitempty"`
	ShippingCountry       *string     `json:"shipping_country,omitempty"`
	CreatedAt             time.Time   `json:"created_at"`
	UpdatedAt             time.Time   `json:"updated_at"`
	Items                 []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID               int64     `json:"id"`
	OrderID          int64     `json:"order_id"`
	ProductID        int64     `json:"product_id"`
	ProductVariantID int64     `json:"product_variant_id"`
	ProductTitle     string    `json:"product_title"`
	ProductSlug      string    `json:"product_slug"`
	VariantSize      string    `json:"variant_size"`
	UnitPrice        float64   `json:"unit_price"`
	Quantity         int       `json:"quantity"`
	Subtotal         float64   `json:"subtotal"`
	ImageURL         string    `json:"image_url"`
	ThumbnailURL     string    `json:"thumbnail_url"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateOrderRequest struct {
	UserID         int64                    `json:"user_id"`
	OrderNumber    string                   `json:"order_number"`
	Currency       string                   `json:"currency"`
	Subtotal       float64                  `json:"subtotal"`
	TaxAmount      float64                  `json:"tax_amount"`
	ShippingAmount float64                  `json:"shipping_amount"`
	TotalAmount    float64                  `json:"total_amount"`
	Items          []CreateOrderItemRequest `json:"items"`
}

type CreateOrderItemRequest struct {
	ProductID        int64   `json:"product_id"`
	ProductVariantID int64   `json:"product_variant_id"`
	ProductTitle     string  `json:"product_title"`
	ProductSlug      string  `json:"product_slug"`
	VariantSize      string  `json:"variant_size"`
	UnitPrice        float64 `json:"unit_price"`
	Quantity         int     `json:"quantity"`
	Subtotal         float64 `json:"subtotal"`
	ImageURL         string  `json:"image_url"`
	ThumbnailURL     string  `json:"thumbnail_url"`
}

type OrderSummary struct {
	ID             int64     `json:"id"`
	OrderNumber    string    `json:"order_number"`
	Status         string    `json:"status"`
	PaymentStatus  string    `json:"payment_status"`
	Currency       string    `json:"currency"`
	Subtotal       float64   `json:"subtotal"`
	TaxAmount      float64   `json:"tax_amount"`
	ShippingAmount float64   `json:"shipping_amount"`
	TotalAmount    float64   `json:"total_amount"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CheckoutSessionResponse struct {
	CheckoutURL string    `json:"checkout_url"`
	SessionID   string    `json:"session_id"`
	OrderID     int64     `json:"order_id"`
	OrderNumber string    `json:"order_number"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type TrackOrderRequest struct {
	OrderNumber string `json:"order_number"`
	Email       string `json:"email"`
}
