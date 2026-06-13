package model

import "time"

type CarouselItem struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	BlobName    string    `json:"blob_name"`
	SortOrder   int       `json:"sort_order"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SaveCarouselItemsRequest struct {
	Items []SaveCarouselItemRequest `json:"items"`
}

type SaveCarouselItemRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	BlobName    string `json:"blob_name"`
	SortOrder   int    `json:"sort_order"`
	IsActive    *bool  `json:"is_active"`
}
