package models

import "time"

type PendingRequest struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id"`
	UserName    string    `json:"user_name"`
	UserEmail   string    `json:"user_email"`
	Type        string    `json:"type"`
	TargetID    uint      `json:"target_id"`
	TargetName  string    `json:"target_name"`
	Status      string    `json:"status"`
	Note        string    `json:"note"`
	RequestedAt time.Time `json:"requested_at" gorm:"autoCreateTime"`
}
