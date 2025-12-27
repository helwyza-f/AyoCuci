package model

import (
    "time"
)

type KategoriPengeluaran struct {
    ID         uint      `gorm:"column:ktg_id;primaryKey;autoIncrement" json:"ktg_id"`
    OutletID   *uint     `gorm:"column:ktg_outlet;index" json:"ktg_outlet"`
    Outlet     *Outlet   `gorm:"foreignKey:OutletID;references:ID" json:"outlet,omitempty"`
    Kategori   string    `gorm:"column:ktg_nama;type:varchar(100);not null" json:"ktg_nama"`
    Status     string    `gorm:"column:ktg_status;type:varchar(20);default:'Aktif'" json:"ktg_status"`
    CreatedAt  time.Time `gorm:"column:ktg_created;autoCreateTime" json:"ktg_created"`
    LastUpdate time.Time `gorm:"column:ktg_lastupdate;autoUpdateTime" json:"ktg_lastupdate"`
    UserUpdate string    `gorm:"column:ktg_userupdate;type:varchar(50)" json:"ktg_userupdate"`
}

func (KategoriPengeluaran) TableName() string {
    return "ac_kategori_pengeluaran"
}

type KategoriPengeluaranInput struct {
    Kategori   string `json:"ktg_nama" binding:"required"`
    Status     string `json:"ktg_status" binding:"omitempty,oneof=Aktif 'Tidak Aktif'"`
}

type UpdateKategoriPengeluaranInput struct {
    Kategori   string `json:"ktg_nama"`
    Status     string `json:"ktg_status" binding:"omitempty,oneof=Aktif 'Tidak Aktif'"`
}

type KategoriPengeluaranResponse struct {
    ID         uint      `json:"ktg_id"`
    OutletID   *uint     `json:"ktg_outlet,omitempty"`
    Kategori   string    `json:"ktg_nama"`
    Status     string    `json:"ktg_status"`
    CreatedAt  time.Time `json:"ktg_created"`
    LastUpdate time.Time `json:"ktg_lastupdate"`
}

func (k *KategoriPengeluaran) ToResponse() KategoriPengeluaranResponse {
    return KategoriPengeluaranResponse{
        ID:         k.ID,
        OutletID:   k.OutletID,
        Kategori:   k.Kategori,
        Status:     k.Status,
        CreatedAt:  k.CreatedAt,
        LastUpdate: k.LastUpdate,
    }
}