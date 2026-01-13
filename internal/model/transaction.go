package model

import "time"

// Transaction Header
type Transaction struct {
	ID            uint                `gorm:"primaryKey" json:"id"`
	InvoiceNumber string              `gorm:"unique;not null" json:"invoice_number"`
	OutletID      uint                `json:"outlet_id"`
	CustomerID    uint                `json:"customer_id"`
	Customer      Customer            `gorm:"foreignKey:CustomerID" json:"customer"`
	ParfumID      *uint               `json:"parfum_id"` // Ubah jadi pointer agar bisa NULL
    Parfum        *Parfume            `gorm:"foreignKey:ParfumID" json:"parfum"`
	DiscountID    *uint               `json:"discount_id"`
	Discount      *Discount           `gorm:"foreignKey:DiscountID" json:"diskon"`
	TotalPrice    float64             `json:"total_price"`
	PaymentStatus string              `gorm:"default:'Belum Bayar'" json:"payment_status"`
	OrderStatus   string              `gorm:"default:'Antrian'" json:"order_status"`
	Notes         string              `json:"notes"`
	Items         []TransactionDetail `gorm:"foreignKey:TransactionID" json:"items"`
	Logs          []OrderLog          `gorm:"foreignKey:TransactionID" json:"logs"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// TransactionDetail - Refactor untuk pelacakan parsial
type TransactionDetail struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TransactionID uint      `json:"transaction_id"`
	ServiceName   string    `json:"service_name"`
	Price         float64   `json:"price"`
	Qty           float64   `json:"qty"`
	Unit          string    `json:"unit"`      // Kg / Pcs / Meter
	IconPath      string    `json:"icon_path"`
	Subtotal      float64   `json:"subtotal"`
	OrderStatus   string    `gorm:"default:'Antrian'" json:"order_status"`
	PaymentStatus string    `gorm:"default:'Belum Bayar'" json:"payment_status"`
	ReadyAt       *time.Time `json:"ready_at"`  // Estimasi Jatuh Tempo per item
	CreatedAt     time.Time `json:"created_at"` // Waktu masuk item
}

// Order Log (History Status)
type OrderLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TransactionID uint      `json:"transaction_id"`
	Status        string    `json:"status"`
	AdminName     string    `json:"admin_name"`
	CreatedAt     time.Time `json:"created_at"`
}