package model

import "time"

type Address struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	FullName     string    `json:"full_name"`
	Phone        string    `json:"phone"`
	AddressLine1 string    `json:"address_line_1"`
	AddressLine2 string    `json:"address_line_2"`
	City         string    `json:"city"`
	Province     string    `json:"province"`
	PostalCode   string    `json:"postal_code"`
	Country      string    `json:"country"`
	IsDefault    bool      `json:"is_default"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AddressRequest struct {
	FullName     string `json:"full_name"`
	Phone        string `json:"phone"`
	AddressLine1 string `json:"address_line_1"`
	AddressLine2 string `json:"address_line_2"`
	City         string `json:"city"`
	Province     string `json:"province"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
	IsDefault    bool   `json:"is_default"`
}
