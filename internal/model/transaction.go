package model

import "time"

// Transaction Header
type Transaction struct {
	ID            uint                `gorm:"primaryKey" json:"id"`
	InvoiceNumber string              `gorm:"unique;not null" json:"invoice_number"` // Contoh: TRX/251029001
	OutletID      uint                `json:"outlet_id"`
	CustomerID    uint                `json:"customer_id"`
	Customer      Customer            `gorm:"foreignKey:CustomerID" json:"customer"`
	ParfumID      uint                `json:"parfum_id"`
	DiscountID    *uint               `json:"discount_id"` // Pointer agar bisa null
	TotalPrice    float64             `json:"total_price"`
	PaymentStatus string              `gorm:"default:'Belum Bayar'" json:"payment_status"` // Belum Bayar / Lunas
	OrderStatus   string              `gorm:"default:'Antrian'" json:"order_status"`     // Antrian / Proses / Siap Ambil / Selesai
	Notes         string              `json:"notes"`
	Items         []TransactionDetail `gorm:"foreignKey:TransactionID" json:"items"`
	Logs          []OrderLog          `gorm:"foreignKey:TransactionID" json:"logs"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// Transaction Detail (Satu baris per layanan)
type TransactionDetail struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	TransactionID uint    `json:"transaction_id"`
	ServiceName   string  `json:"service_name"` // Simpan nama saat transaksi
	Price         float64 `json:"price"`
	Qty           float64 `json:"qty"` // Mendukung desimal (Kg)
	Subtotal      float64 `json:"subtotal"`
}

// Order Log (History Status)
type OrderLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TransactionID uint      `json:"transaction_id"`
	Status        string    `json:"status"`
	AdminName     string    `json:"admin_name"`
	CreatedAt     time.Time `json:"created_at"`
}