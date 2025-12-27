package model

import (
	"time"
)

// OTP model untuk menyimpan kode OTP
type OTP struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	NomorHP   string    `json:"nomor_hp" gorm:"index;not null"`
	Code      string    `json:"code" gorm:"not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	IsUsed    bool      `json:"is_used" gorm:"default:false"`
	Purpose   string    `json:"purpose" gorm:"type:varchar(50);not null"` 
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (OTP) TableName() string {
	return "otps"
}


type ForgotPasswordInput struct {
	NomorHP string `json:"nomor_hp" validate:"required,min=10,max=15"`
}


type VerifyOTPInput struct {
	NomorHP string `json:"nomor_hp" validate:"required"`
	Code    string `json:"code" validate:"required,len=6"`
}


type ResetPasswordInput struct {
	NomorHP         string `json:"nomor_hp" validate:"required"`
	Code            string `json:"code" validate:"required,len=6"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}


type OTPResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}