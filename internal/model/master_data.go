package model

import "time"

type Parfume struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	OutletID    uint      `json:"outlet_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Discount struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	OutletID    uint      `json:"outlet_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // Nominal / Persen
	Value       float64   `json:"value"`
	Description string    `json:"description"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}