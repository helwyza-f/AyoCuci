package model

import (
	"gorm.io/gorm"
	"time"
)

type Layanan struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	OutletID     uint           `json:"ln_outlet" gorm:"not null;index"`
	NamaLayanan  string         `json:"ln_layanan" gorm:"type:varchar(255);not null"` 
	Prioritas    int            `json:"ln_prioritas" gorm:"type:int;default:0"`
	Cuci         string         `json:"ln_cuci" gorm:"type:varchar(100)"`
	Kering       string         `json:"ln_kering" gorm:"type:varchar(100)"`
	Setrika      string         `json:"ln_setrika" gorm:"type:varchar(100)"`
	
	CreatedAt    time.Time      `json:"ln_created" gorm:"autoCreateTime"`
	LastUpdate   time.Time      `json:"ln_lastupdate" gorm:"autoUpdateTime"`
	UserUpdateID *uint          `json:"ln_userupdate" gorm:"index"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relasi
	Outlet       Outlet         `json:"outlet,omitempty" gorm:"foreignKey:OutletID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	UserUpdate   *User          `json:"user_update,omitempty" gorm:"foreignKey:UserUpdateID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	JenisProduk  []JenisProduk  `json:"jenis_produk,omitempty" gorm:"foreignKey:LayananID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Layanan) TableName() string {
	return "ac_layanan"
}

type JenisProduk struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	LayananID      uint           `json:"jp_layanan_id" gorm:"not null;index"`
	Nama           string         `json:"jp_nama" gorm:"type:varchar(255);not null"` 
	IconPath       *string        `json:"jp_icon_path" gorm:"type:varchar(500)"`
	ImageURL       *string        `json:"jp_image_url" gorm:"type:varchar(500)"`
	Satuan         *string        `json:"jp_satuan" gorm:"type:varchar(50)"`
	HargaPer       *int           `json:"jp_harga_per" gorm:"type:int"`
	LamaPengerjaan *int           `json:"jp_lama_pengerjaan" gorm:"type:int"`
	SatuanWaktu    *string        `json:"jp_satuan_waktu" gorm:"type:varchar(20)"`
	Keterangan     *string        `json:"jp_keterangan" gorm:"type:text"`
	
	CreatedAt      time.Time      `json:"jp_created" gorm:"autoCreateTime"`
	LastUpdate     time.Time      `json:"jp_lastupdate" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

func (JenisProduk) TableName() string {
	return "ac_jenis_produk"
}

type CreateLayananWithProductsInput struct {
	OutletID    uint                        `json:"ln_outlet" validate:"required"`
	NamaLayanan string                      `json:"ln_layanan" validate:"required,min=3,max=255"`
	Prioritas   int                         `json:"ln_prioritas" validate:"omitempty,min=0,max=100"`
	Cuci        string                      `json:"ln_cuci" validate:"required"`
	Kering      string                      `json:"ln_kering" validate:"required"`
	Setrika     string                      `json:"ln_setrika" validate:"required"`
	JenisProduk []CreateJenisProdukInput    `json:"jenis_produk" validate:"required,min=1"`
}


type CreateJenisProdukInput struct {
	Nama           string  `json:"jp_nama" validate:"required,min=1,max=255"`
	IconPath       *string `json:"jp_icon_path" validate:"omitempty,max=500"`
	Satuan         *string `json:"jp_satuan" validate:"omitempty"`
	HargaPer       *int    `json:"jp_harga_per" validate:"omitempty,min=0"`
	LamaPengerjaan *int    `json:"jp_lama_pengerjaan" validate:"omitempty,min=1"`
	SatuanWaktu    *string `json:"jp_satuan_waktu" validate:"omitempty"`
	Keterangan     *string `json:"jp_keterangan" validate:"omitempty"`
	ImageFileName  string  `json:"-"` 
}


type UpdateLayananWithProductsInput struct {
	NamaLayanan    string                      `json:"ln_layanan" validate:"omitempty,min=3,max=255"`
	Prioritas      *int                        `json:"ln_prioritas" validate:"omitempty,min=0,max=100"`
	Cuci           string                      `json:"ln_cuci" validate:"omitempty"`
	Kering         string                      `json:"ln_kering" validate:"omitempty"`
	Setrika        string                      `json:"ln_setrika" validate:"omitempty"`
	JenisProduk    []UpdateJenisProdukInput    `json:"jenis_produk" validate:"omitempty"`
}

type UpdateJenisProdukInput struct {
	ID             *uint   `json:"id"`
	Nama           string  `json:"jp_nama" validate:"omitempty,min=1,max=255"`
	IconPath       *string `json:"jp_icon_path" validate:"omitempty,max=500"`
	Satuan         *string `json:"jp_satuan" validate:"omitempty"`
	HargaPer       *int    `json:"jp_harga_per" validate:"omitempty,min=0"`
	LamaPengerjaan *int    `json:"jp_lama_pengerjaan" validate:"omitempty,min=1"`
	SatuanWaktu    *string `json:"jp_satuan_waktu" validate:"omitempty"`
	Keterangan     *string `json:"jp_keterangan" validate:"omitempty"`
	ImageFileName  string  `json:"-"`
	ShouldDelete   bool    `json:"delete,omitempty"` 
}


type CreateLayananInput struct {
	OutletID       uint    `json:"ln_outlet" validate:"required"`
	NamaLayanan    string  `json:"ln_layanan" validate:"required,min=3,max=255"`
	Prioritas      int     `json:"ln_prioritas" validate:"omitempty,min=0,max=100"`
	Cuci           string  `json:"ln_cuci" validate:"omitempty,max=100"`
	Kering         string  `json:"ln_kering" validate:"omitempty,max=100"`
	Setrika        string  `json:"ln_setrika" validate:"omitempty,max=100"`
	IconPath       *string `json:"ln_icon_path" validate:"omitempty,max=500"`
	ImageURL       *string `json:"ln_image_url" validate:"omitempty,max=500"`
	Satuan         *string `json:"ln_satuan" validate:"omitempty"`
	HargaPer       *int    `json:"ln_harga_per" validate:"omitempty,min=0"`
	LamaPengerjaan *int    `json:"ln_lama_pengerjaan" validate:"omitempty,min=1"`
	SatuanWaktu    *string `json:"ln_satuan_waktu" validate:"omitempty"`
	Keterangan     *string `json:"ln_keterangan" validate:"omitempty"`
}

type UpdateLayananInput struct {
	NamaLayanan    string  `json:"ln_layanan" validate:"omitempty,min=3,max=255"`
	Prioritas      *int    `json:"ln_prioritas" validate:"omitempty,min=0,max=100"`
	Cuci           string  `json:"ln_cuci" validate:"omitempty,max=100"`
	Kering         string  `json:"ln_kering" validate:"omitempty,max=100"`
	Setrika        string  `json:"ln_setrika" validate:"omitempty,max=100"`
	IconPath       *string `json:"ln_icon_path" validate:"omitempty,max=500"`
	ImageURL       *string `json:"ln_image_url" validate:"omitempty,max=500"`
	Satuan         *string `json:"ln_satuan" validate:"omitempty"`
	HargaPer       *int    `json:"ln_harga_per" validate:"omitempty,min=0"`
	LamaPengerjaan *int    `json:"ln_lama_pengerjaan" validate:"omitempty,min=1"`
	SatuanWaktu    *string `json:"ln_satuan_waktu" validate:"omitempty"`
	Keterangan     *string `json:"ln_keterangan" validate:"omitempty"`
}



type LayananList struct {
	ID             uint            `json:"id"`
	OutletID       uint            `json:"ln_outlet"`
	NamaLayanan    string          `json:"ln_layanan"`
	Prioritas      int             `json:"ln_prioritas"`
	Cuci           string          `json:"ln_cuci"`
	Kering         string          `json:"ln_kering"`
	Setrika        string          `json:"ln_setrika"`
	CreatedAt      time.Time       `json:"ln_created"`
	LastUpdate     time.Time       `json:"ln_lastupdate"`
	JenisProduk    []JenisProduk   `json:"jenis_produk,omitempty"` 
}

type LayananDetail struct {
	ID             uint            `json:"id"`
	OutletID       uint            `json:"ln_outlet"`
	NamaLayanan    string          `json:"ln_layanan"`
	Prioritas      int             `json:"ln_prioritas"`
	Cuci           string          `json:"ln_cuci"`
	Kering         string          `json:"ln_kering"`
	Setrika        string          `json:"ln_setrika"`
	CreatedAt      time.Time       `json:"ln_created"`
	LastUpdate     time.Time       `json:"ln_lastupdate"`
	UserUpdateID   *uint           `json:"ln_userupdate"`
	Outlet         *Outlet         `json:"outlet,omitempty"`
	UserUpdate     *UserList       `json:"user_update,omitempty"`
	JenisProduk    []JenisProduk   `json:"jenis_produk,omitempty"`
}

type LayananResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Data    *Layanan `json:"data,omitempty"`
}

type LayananWithProductsResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Data    *Layanan `json:"data,omitempty"`
}
func (Layanan) PreloadRelations(db *gorm.DB) *gorm.DB {
    return db.Preload("JenisProduk").Preload("Outlet")
}