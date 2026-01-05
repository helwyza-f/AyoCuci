package model

import (
	"time"
)

type Customer struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OutletID  uint      `json:"outlet_id"` // FK ke Outlet
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Gender    string    `json:"gender"`  // Pria / Wanita
	OTP       *string   `json:"otp"`     // Nullable
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}