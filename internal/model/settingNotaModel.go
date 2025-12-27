package model

import (
    "time"
    "gorm.io/gorm"
)

type NotaSettings struct {
    ID                 uint           `json:"id" gorm:"primaryKey"`
    OutletID           uint           `json:"outlet_id" gorm:"not null;uniqueIndex"`
    BusinessName       string         `json:"business_name" gorm:"type:varchar(255)"`
    Address            string         `json:"address" gorm:"type:text"`
    Phone              string         `json:"phone" gorm:"type:varchar(20)"`
    FooterNote         string         `json:"footer_note" gorm:"type:text"`
    WhatsappNote       string         `json:"whatsapp_note" gorm:"type:text"`
    ShowLogo           bool           `json:"show_logo" gorm:"default:false"`
    ShowQRCode         bool           `json:"show_qr_code" gorm:"default:true"`
    ShowBusinessName   bool           `json:"show_business_name" gorm:"default:true"`
    ShowDescription    bool           `json:"show_description" gorm:"default:true"`
    ShowFooterNote     bool           `json:"show_footer_note" gorm:"default:true"`
    ShowWhatsappFooter bool           `json:"show_whatsapp_footer" gorm:"default:true"`
    PrinterSize        int            `json:"printer_size" gorm:"default:58"`
    PrinterType        string         `json:"printer_type" gorm:"type:varchar(1);default:'A'"`
    CreatedAt          time.Time      `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt          time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
    DeletedAt          gorm.DeletedAt `json:"-" gorm:"index"`
    
    Outlet Outlet `json:"outlet,omitempty" gorm:"foreignKey:OutletID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (NotaSettings) TableName() string {
    return "nota_settings"
}

type NotaSettingsInput struct {
    BusinessName       string `json:"business_name" validate:"omitempty,max=255"`
    Address            string `json:"address" validate:"omitempty"`
    Phone              string `json:"phone" validate:"omitempty,max=20"`
    FooterNote         string `json:"footer_note" validate:"omitempty"`
    WhatsappNote       string `json:"whatsapp_note" validate:"omitempty"`
    ShowLogo           bool   `json:"show_logo"`
    ShowQRCode         bool   `json:"show_qr_code"`
    ShowBusinessName   bool   `json:"show_business_name"`
    ShowDescription    bool   `json:"show_description"`
    ShowFooterNote     bool   `json:"show_footer_note"`
    ShowWhatsappFooter bool   `json:"show_whatsapp_footer"`
    PrinterSize        int    `json:"printer_size" validate:"required,oneof=58 80"`
    PrinterType        string `json:"printer_type" validate:"required,oneof=A B"`
}

type NotaSettingsResponse struct {
    Success bool          `json:"success"`
    Message string        `json:"message"`
    Data    *NotaSettings `json:"data,omitempty"`
}

type NotaSettingsErrorResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}