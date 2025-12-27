// File: internal/model/parfum.go
package model

import (
	"time"
)

type Parfum struct {
	ID           uint      `gorm:"column:prf_id;primaryKey;autoIncrement" json:"prf_id"`
	OutletID     *uint     `gorm:"column:prf_outlet;index" json:"prf_outlet"`
	Parfum       string    `gorm:"column:prf_nama;type:varchar(100);not null" json:"prf_nama"`
	Keterangan   string    `gorm:"column:prf_keterangan;type:text" json:"prf_keterangan"`
	Status       string    `gorm:"column:prf_status;type:varchar(20);default:'Aktif'" json:"prf_status"`
	CreatedAt    time.Time `gorm:"column:prf_created;autoCreateTime" json:"prf_created"`
	LastUpdate   time.Time `gorm:"column:prf_lastupdate;autoUpdateTime" json:"prf_lastupdate"`
	UserUpdate   string    `gorm:"column:prf_userupdate;type:varchar(50)" json:"prf_userupdate"`
}

func (Parfum) TableName() string {
	return "ac_parfum"
}

type ParfumInput struct {
	Parfum     string `json:"prf_nama" binding:"required"`
	Keterangan string `json:"prf_keterangan"`
	Status     string `json:"prf_status" binding:"omitempty,oneof=Aktif 'Tidak Aktif'"`
}

type UpdateParfumInput struct {
	Parfum     string `json:"prf_nama"`
	Keterangan string `json:"prf_keterangan"`
	Status     string `json:"prf_status" binding:"omitempty,oneof=Aktif 'Tidak Aktif'"`
}

type ParfumResponse struct {
	ID         uint      `json:"prf_id"`
	OutletID   *uint     `json:"prf_outlet,omitempty"`
	Parfum     string    `json:"prf_nama"`
	Keterangan string    `json:"prf_keterangan"`
	Status     string    `json:"prf_status"`
	CreatedAt  time.Time `json:"prf_created"`
	LastUpdate time.Time `json:"prf_lastupdate"`
}

func (p *Parfum) ToResponse() ParfumResponse {
	return ParfumResponse{
		ID:         p.ID,
		OutletID:   p.OutletID,
		Parfum:     p.Parfum,
		Keterangan: p.Keterangan,
		Status:     p.Status,
		CreatedAt:  p.CreatedAt,
		LastUpdate: p.LastUpdate,
	}
}