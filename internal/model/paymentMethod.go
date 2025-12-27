package model

import (
	"time"
	"gorm.io/gorm"
)


type PaymentMethod struct {
    ID              uint           `gorm:"primarykey" json:"id"`
    OutletID        uint           `json:"outlet_id"`
    Name            string         `json:"name"`
    IsActive        bool           `json:"is_active"`
    Category        string         `json:"category"` 
    BankName        *string        `json:"bank_name"`
    AccountNumber   *string        `json:"account_number"`
    AccountHolder   *string        `json:"account_holder"`
    EwalletProvider *string        `json:"ewallet_provider"`
    PhoneNumber     *string        `json:"phone_number"`
    CreatedAt       time.Time      `json:"created_at"`
    UpdatedAt       time.Time      `json:"updated_at"`
    DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (PaymentMethod) TableName() string {
    return "payment_methods"
}

type PaymentMethodInput struct {
	Name            string  `json:"name" validate:"required,min=3,max=100"`
	IsActive        bool    `json:"is_active"`
	Category        string  `json:"category" validate:"required,oneof=Cash Transfer E-Wallet"`
	BankName        *string `json:"bank_name"`
	AccountNumber   *string `json:"account_number"`
	AccountHolder   *string `json:"account_holder"`
	EwalletProvider *string `json:"ewallet_provider"`
	PhoneNumber     *string `json:"phone_number"`
}

type UpdatePaymentMethodInput struct {
	Name            *string `json:"name" validate:"omitempty,min=3,max=100"`
	IsActive        *bool   `json:"is_active"`
	Category        *string `json:"category" validate:"omitempty,oneof=Cash Transfer E-Wallet"`
	BankName        *string `json:"bank_name"`
	AccountNumber   *string `json:"account_number"`
	AccountHolder   *string `json:"account_holder"`
	EwalletProvider *string `json:"ewallet_provider"`
	PhoneNumber     *string `json:"phone_number"`
}

type PaymentMethodList struct {
	ID              uint    `json:"id"`
	OutletID        uint    `json:"outlet_id"`
	Name            string  `json:"name"`
	IsActive        bool    `json:"is_active"`
	Category        string  `json:"category"`
	BankName        *string `json:"bank_name,omitempty"`
	AccountNumber   *string `json:"account_number,omitempty"`
	AccountHolder   *string `json:"account_holder,omitempty"`
	EwalletProvider *string `json:"ewallet_provider,omitempty"`
	PhoneNumber     *string `json:"phone_number,omitempty"`
	CreatedAt       string  `json:"created_at"`
}

type PaymentMethodResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    *PaymentMethod `json:"data,omitempty"`
}

type PaymentMethodListResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    []PaymentMethodList `json:"data,omitempty"`
	Total   int64               `json:"total,omitempty"`
}