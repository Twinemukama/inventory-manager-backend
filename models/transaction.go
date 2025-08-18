package models

import "time"

type TransactionType string

const (
	TransactionIn  TransactionType = "IN"
	TransactionOut TransactionType = "OUT"
)

type Transaction struct {
	ID        uint            `json:"id" gorm:"primaryKey"`
	ItemID    uint            `json:"item_id"`
	Quantity  int             `json:"quantity"`
	Type      TransactionType `json:"type"` // IN or OUT
	Note      string          `json:"note"`
	UserID    uint            `json:"user_id"`
	CreatedAt time.Time
}
