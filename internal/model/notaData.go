package model

import (
    "time"
    "gorm.io/gorm"
)

type NotaItem struct {
    ProductName string  `json:"product_name"`
    Quantity    float64 `json:"quantity"`
    Unit        string  `json:"unit"`
    Price       float64 `json:"price"`
    Subtotal    float64 `json:"subtotal"`
    Description string  `json:"description,omitempty"`
}

type NotaData struct {
    ID              uint           `json:"id" gorm:"primaryKey"`
    OutletID        uint           `json:"outlet_id" gorm:"not null;index"`
    TransactionID   string         `json:"transaction_id" gorm:"type:varchar(100);uniqueIndex"`
    TransactionDate time.Time      `json:"transaction_date" gorm:"not null"`
    CustomerName    string         `json:"customer_name" gorm:"type:varchar(255)"`
    CustomerPhone   string         `json:"customer_phone" gorm:"type:varchar(20)"`
    CashierName     string         `json:"cashier_name" gorm:"type:varchar(255)"`
    Items           []NotaItem     `json:"items" gorm:"-"`
    ItemsJSON       string         `json:"-" gorm:"type:text"` 
    Subtotal        float64        `json:"subtotal" gorm:"type:decimal(15,2)"`
    Tax             float64        `json:"tax" gorm:"type:decimal(15,2);default:0"`
    TaxPercentage   float64        `json:"tax_percentage" gorm:"type:decimal(5,2);default:0"`
    Discount        float64        `json:"discount" gorm:"type:decimal(15,2);default:0"`
    DiscountType    string         `json:"discount_type" gorm:"type:varchar(20)"` 
    ServiceCharge   float64        `json:"service_charge" gorm:"type:decimal(15,2);default:0"`
    Total           float64        `json:"total" gorm:"type:decimal(15,2);not null"`
    PaymentAmount   float64        `json:"payment_amount" gorm:"type:decimal(15,2)"`
    Change          float64        `json:"change_amount" gorm:"type:decimal(15,2);default:0"`
    PaymentMethod   string         `json:"payment_method" gorm:"type:varchar(50)"` 
    
    // Additional Info
    Notes           string         `json:"notes" gorm:"type:text"`
    QRCodeData      string         `json:"qr_code_data" gorm:"type:text"` 
    Status          string         `json:"status" gorm:"type:varchar(20);default:'completed'"` 
    PrintCount      int            `json:"print_count" gorm:"default:0"`
    LastPrintedAt   *time.Time     `json:"last_printed_at"`
    
    CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
    DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
    Outlet          Outlet         `json:"outlet,omitempty" gorm:"foreignKey:OutletID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    NotaSettings    *NotaSettings  `json:"nota_settings,omitempty" gorm:"-"` 
}

func (NotaData) TableName() string {
    return "nota_data"
}

type NotaItemDetail struct {
    ID            uint           `json:"id" gorm:"primaryKey"`
    NotaDataID    uint           `json:"nota_data_id" gorm:"not null;index"`
    ProductID     uint           `json:"product_id" gorm:"index"`
    ProductName   string         `json:"product_name" gorm:"type:varchar(255);not null"`
    Quantity      float64        `json:"quantity" gorm:"type:decimal(10,2);not null"`
    Unit          string         `json:"unit" gorm:"type:varchar(50)"`
    Price         float64        `json:"price" gorm:"type:decimal(15,2);not null"`
    Subtotal      float64        `json:"subtotal" gorm:"type:decimal(15,2);not null"`
    Description   string         `json:"description" gorm:"type:text"`
    CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
    DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
    
    NotaData      NotaData       `json:"-" gorm:"foreignKey:NotaDataID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (NotaItemDetail) TableName() string {
    return "nota_item_details"
}

type NotaGenerateInput struct {
    OutletID        uint       `json:"outlet_id" validate:"required"`
    TransactionID   string     `json:"transaction_id" validate:"required,max=100"`
    TransactionDate time.Time  `json:"transaction_date" validate:"required"`
    CustomerName    string     `json:"customer_name" validate:"omitempty,max=255"`
    CustomerPhone   string     `json:"customer_phone" validate:"omitempty,max=20"`
    CashierName     string     `json:"cashier_name" validate:"required,max=255"`
    Items           []NotaItem `json:"items" validate:"required,min=1,dive"`
    Tax             float64    `json:"tax" validate:"gte=0"`
    TaxPercentage   float64    `json:"tax_percentage" validate:"gte=0,lte=100"`
    Discount        float64    `json:"discount" validate:"gte=0"`
    DiscountType    string     `json:"discount_type" validate:"omitempty,oneof=percentage fixed"`
    ServiceCharge   float64    `json:"service_charge" validate:"gte=0"`
    PaymentAmount   float64    `json:"payment_amount" validate:"required,gt=0"`
    PaymentMethod   string     `json:"payment_method" validate:"required,oneof=cash card qris transfer ewallet"`
    Notes           string     `json:"notes" validate:"omitempty"`
}

type NotaPrintInput struct {
    NotaDataID uint `json:"nota_data_id" validate:"required"`
    Reprint    bool `json:"reprint"`
}

type NotaPreviewInput struct {
    OutletID uint                `json:"outlet_id" validate:"required"`
    Data     NotaGenerateInput   `json:"data" validate:"required"`
}

type NotaGenerateResponse struct {
    Success bool      `json:"success"`
    Message string    `json:"message"`
    Data    *NotaData `json:"data,omitempty"`
}

type NotaPrintResponse struct {
    Success   bool   `json:"success"`
    Message   string `json:"message"`
    PrintData string `json:"print_data,omitempty"` 
    Format    string `json:"format,omitempty"`     
}

type NotaPreviewResponse struct {
    Success     bool   `json:"success"`
    Message     string `json:"message"`
    PreviewHTML string `json:"preview_html,omitempty"`
    PreviewData string `json:"preview_data,omitempty"`
}


type NotaPrintFormat struct {
    Settings        *NotaSettings `json:"settings"`
    Data            *NotaData     `json:"data"`
    FormattedItems  []string      `json:"formatted_items"`
    FormattedTotal  string        `json:"formatted_total"`
    QRCodeBase64    string        `json:"qr_code_base64,omitempty"`
    LogoBase64      string        `json:"logo_base64,omitempty"`
    PrintWidth      int           `json:"print_width"` 
}