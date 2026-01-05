package model

import (
	"time"

	"gorm.io/gorm"
)

type Outlet struct {
    ID         uint           `json:"id" gorm:"primaryKey;column:id"`
    UserID     uint           `json:"user_id" gorm:"not null;index;column:user_id"`
    Photo      string         `json:"photo" gorm:"type:varchar(255);column:photo"`
    NamaOutlet string         `json:"nama_outlet" gorm:"not null;type:varchar(100);column:nama_outlet"`
    Alamat     string         `json:"alamat" gorm:"not null;type:text;column:alamat"`
    NomorHP    string         `json:"nomor_hp" gorm:"not null;type:varchar(20);column:nomor_hp"`
    Provinsi   string         `json:"provinsi" gorm:"type:varchar(100);column:provinsi"`
    Kota       string         `json:"kota" gorm:"type:varchar(100);column:kota"`
    Kecamatan  string         `json:"kecamatan" gorm:"type:varchar(100);column:kecamatan"`
    IsAktif    string         `json:"is_aktif" gorm:"type:varchar(20);default:'active';column:is_aktif"`
    CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime;column:created_at"`
    UpdatedAt  time.Time      `json:"updated_at" gorm:"autoUpdateTime;column:updated_at"`
    DeletedAt  gorm.DeletedAt `json:"-" gorm:"index;column:deleted_at"`
    User       User           `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // ✅ Pastikan ada ini
}

func (Outlet) TableName() string {
    return "outlets"
}

type OutletInput struct {
    NamaOutlet string `json:"nama_outlet" validate:"required,min=3"`
    Alamat     string `json:"alamat" validate:"required,min=10"`
    NomorHP    string `json:"nomor_hp" validate:"required,min=10,max=15"`
    Provinsi   string `json:"provinsi" validate:"required"`
    Kota       string `json:"kota" validate:"required"`
    Kecamatan  string `json:"kecamatan" validate:"required"`
}

// type RegisterWithOutletInput struct {
//     NamaLengkap         string `json:"username" validate:"required,min=3"`
//     Email               string `json:"email" validate:"required,email"`
//     Password            string `json:"password" validate:"required,min=8"`
//     ConfirmPassword     string `json:"confirmPassword" validate:"required"`
//     NomorHP             string `json:"nomor_hp" validate:"required,min=10,max=15"`
//     Group               string `json:"group" validate:"required,oneof=owner karyawan"`
//     AgreeTerms          bool   `json:"agreeTerms" validate:"required"`
//     SubscribeNewsletter bool   `json:"subscribeNewsletter"`
    
//     Outlet OutletInput `json:"outlet" validate:"required"`
// }

type OutletList struct {
    ID         uint   `json:"id"`
    UserID     uint   `json:"user_id"`
    NamaOutlet string `json:"nama_outlet"`
    Alamat     string `json:"alamat"`
    Photo      string `json:"photo"`
    NomorHP    string `json:"nomor_hp"`
    Provinsi   string `json:"provinsi"`
    Kota       string `json:"kota"`
    Kecamatan  string `json:"kecamatan"`
    IsAktif    string `json:"is_aktif"`
    CreatedAt  string `json:"created_at"`
}

type OutletResponse struct {
    Success bool    `json:"success"`
    Message string  `json:"message"`
    Outlet  *Outlet `json:"outlet,omitempty"`
}

type OutletListResponse struct {
    Success bool         `json:"success"`
    Message string       `json:"message"`
    Data    []OutletList `json:"data,omitempty"`
    Total   int64        `json:"total,omitempty"`
}

type UpdateOutletInput struct {
    NamaOutlet string `json:"nama_outlet" validate:"omitempty,min=3"`
    Alamat     string `json:"alamat" validate:"omitempty,min=10"`
    NomorHP    string `json:"nomor_hp" validate:"omitempty,min=10,max=15"`
    Provinsi   string `json:"provinsi" validate:"omitempty"`
    Kota       string `json:"kota" validate:"omitempty"`
    Kecamatan  string `json:"kecamatan" validate:"omitempty"`
    IsAktif    string `json:"is_aktif" validate:"omitempty,oneof=active inactive"`
}