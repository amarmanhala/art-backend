package model

import "time"

type ArtStyle struct {
	ID          int64     `json:"id"`
	Origin      string    `json:"origin"`
	Style       string    `json:"style"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	ImageURL    string    `json:"image_url"`
	BlobName    string    `json:"blob_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SaveArtStyleRequest struct {
	Origin      string   `json:"origin"`
	Style       string   `json:"style"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	ImageURL    string   `json:"image_url"`
	BlobName    string   `json:"blob_name"`
}
