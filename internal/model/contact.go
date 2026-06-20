package model

import "time"

type ContactMessage struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	OrderNumber *string   `json:"order_number,omitempty"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
}

type ContactRequest struct {
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	OrderNumber *string `json:"order_number"`
	Message     string  `json:"message"`
}
