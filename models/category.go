package models

import "time"

type Category struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	Name      string  `json:"name"`
	UserID    uint    `json:"user_id"`
	User      User    `json:"user" gorm:"foreignKey:UserID"`
	CompanyID uint    `json:"company_id"`
	Company   Company `json:"company" gorm:"foreignKey:CompanyID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
