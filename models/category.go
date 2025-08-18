package models

import (
	"time"
)

type Category struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Name      string `json:"name" gorm:"unique"`
	UserID    uint   `json:"user_id"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
