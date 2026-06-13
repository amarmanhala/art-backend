package model

import "time"

type Product struct {
	ID            int64          `json:"id"`
	Title         string         `json:"title"`
	Slug          string         `json:"slug"`
	Description   string         `json:"description"`
	Price         float64        `json:"price"`
	Currency      string         `json:"currency"`
	Category      string         `json:"category"`
	Style         string         `json:"style"`
	Theme         string         `json:"theme"`
	Orientation   string         `json:"orientation"`
	Size          string         `json:"size"`
	ImageURL      string         `json:"image_url"`
	ThumbnailURL  string         `json:"thumbnail_url"`
	OriginalURL   string         `json:"original_url"`
	Images        []ProductImage `json:"images,omitempty"`
	StockQuantity int            `json:"stock_quantity"`
	IsActive      bool           `json:"is_active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type ProductImage struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	ImageURL  string    `json:"image_url"`
	AltText   string    `json:"alt_text"`
	SortOrder int       `json:"sort_order"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	Title         string  `json:"title"`
	Slug          string  `json:"slug"`
	Description   string  `json:"description"`
	Price         float64 `json:"price"`
	Currency      string  `json:"currency"`
	Category      string  `json:"category"`
	Style         string  `json:"style"`
	Theme         string  `json:"theme"`
	Orientation   string  `json:"orientation"`
	Size          string  `json:"size"`
	ImageURL      string  `json:"image_url"`
	ThumbnailURL  string  `json:"thumbnail_url"`
	OriginalURL   string  `json:"original_url"`
	StockQuantity int     `json:"stock_quantity"`
	IsActive      *bool   `json:"is_active"`
}

type UpdateProductRequest struct {
	Title         *string  `json:"title"`
	Slug          *string  `json:"slug"`
	Description   *string  `json:"description"`
	Price         *float64 `json:"price"`
	Currency      *string  `json:"currency"`
	Category      *string  `json:"category"`
	Style         *string  `json:"style"`
	Theme         *string  `json:"theme"`
	Orientation   *string  `json:"orientation"`
	Size          *string  `json:"size"`
	ImageURL      *string  `json:"image_url"`
	ThumbnailURL  *string  `json:"thumbnail_url"`
	OriginalURL   *string  `json:"original_url"`
	StockQuantity *int     `json:"stock_quantity"`
	IsActive      *bool    `json:"is_active"`
}

type ProductFilter struct {
	Category    string
	Style       string
	Theme       string
	Orientation string
	MinPrice    *float64
	MaxPrice    *float64
}
