package model

import "time"

type Cart struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CartItem struct {
	ID        int64     `json:"id"`
	CartID    int64     `json:"cart_id"`
	Product   Product   `json:"product"`
	Quantity  int       `json:"quantity"`
	Subtotal  float64   `json:"subtotal"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CartResponse struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	Items      []CartItem `json:"items"`
	TotalItems int        `json:"total_items"`
	TotalPrice float64    `json:"total_price"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type AddCartItemRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity"`
}
