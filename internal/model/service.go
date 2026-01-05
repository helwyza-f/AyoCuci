package model

import (
	"time"
)

type ServiceCategory struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	OutletID  uint             `json:"outlet_id"` // FK ke Outlet
	Name      string           `json:"name"`
	Processes string           `json:"processes"` // Kita simpan sebagai string "Cuci,Kering,Setrika"
	Products  []ServiceProduct `gorm:"foreignKey:CategoryID" json:"items"`
	CreatedAt time.Time        `json:"created_at"`
}

type ServiceProduct struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	CategoryID   uint    `json:"category_id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Unit         string  `json:"unit"`
	Duration     int     `json:"duration"`
	DurationUnit string  `json:"duration_unit"`
	IconPath     string  `json:"icon_path"`
	Note         *string `json:"note"`
}