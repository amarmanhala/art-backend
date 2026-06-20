package model

import "time"

type ProductSize struct {
	ID        int64     `json:"id"`
	Label     string    `json:"label"`
	WidthIn   float64   `json:"width_in"`
	HeightIn  float64   `json:"height_in"`
	WidthCM   float64   `json:"width_cm"`
	HeightCM  float64   `json:"height_cm"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
