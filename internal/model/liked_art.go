package model

import "time"

const (
	LikedArtStatusLiked    = "liked"
	LikedArtStatusDisliked = "disliked"
)

type LikedArt struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	ProductID int64     `json:"product_id"`
	Status    string    `json:"status"`
	Product   Product   `json:"product"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SaveLikedArtRequest struct {
	ProductID int64  `json:"product_id"`
	Action    string `json:"action"`
}
