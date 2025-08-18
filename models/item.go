package models

import (
	"time"
)

type Item struct {
	ID          uint    `json:"id" gorm:"primaryKey"`
	Name        string  `json:"name"`
	SKU         string  `json:"sku"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	CategoryID  uint    `json:"category_id"`
	UserID      uint    `json:"user_id"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
