package model

import (
	"time"
)


type Diskon struct {
	ID           uint      `gorm:"column:dis_id;primaryKey;autoIncrement" json:"dis_id"`
	OutletID     *uint     `gorm:"column:dis_outlet;index" json:"dis_outlet"`
	Diskon       string    `gorm:"column:dis_diskon;type:varchar(100);not null" json:"dis_diskon"`
	Jenis        string    `gorm:"column:dis_jenis;type:varchar(20);not null" json:"dis_jenis"` 
	NilaiDiskon  float64   `gorm:"column:dis_nilai_diskon;type:decimal(15,2);not null" json:"dis_nilai_diskon"`
	Keterangan   string    `gorm:"column:dis_keterangan;type:text" json:"dis_keterangan"`
	Status       string    `gorm:"column:dis_status;type:varchar(20);default:'Aktif'" json:"dis_status"`
	CreatedAt    time.Time `gorm:"column:dis_created;autoCreateTime" json:"dis_created"`
	LastUpdate   time.Time `gorm:"column:dis_lastupdate;autoUpdateTime" json:"dis_lastupdate"`
	UserUpdate   string    `gorm:"column:dis_userupdate;type:varchar(50)" json:"dis_userupdate"`
}


func (Diskon) TableName() string {
	return "ac_diskon"
}

type DiskonInput struct {
	Diskon      string  `json:"dis_diskon" binding:"required"`
	Jenis       string  `json:"dis_jenis" binding:"required,oneof=Nominal Persen"`
	NilaiDiskon float64 `json:"dis_nilai_diskon" binding:"required,gt=0"`
	Keterangan  string  `json:"dis_keterangan"`
	Status      string  `json:"dis_status" binding:"omitempty,oneof=Aktif 'Tidak Aktif'"`
}

type UpdateDiskonInput struct {
	Diskon      string  `json:"dis_diskon"`
	Jenis       string  `json:"dis_jenis" binding:"omitempty,oneof=Nominal Persen"`
	NilaiDiskon float64 `json:"dis_nilai_diskon" binding:"omitempty,gt=0"`
	Keterangan  string  `json:"dis_keterangan"`
	Status      string  `json:"dis_status" binding:"omitempty,oneof=Aktif 'Tidak Aktif'"`
}

type DiskonResponse struct {
	ID          uint      `json:"dis_id"`
	OutletID    *uint     `json:"dis_outlet,omitempty"`
	Diskon      string    `json:"dis_diskon"`
	Jenis       string    `json:"dis_jenis"`
	NilaiDiskon float64   `json:"dis_nilai_diskon"`
	Keterangan  string    `json:"dis_keterangan"`
	Status      string    `json:"dis_status"`
	CreatedAt   time.Time `json:"dis_created"`
	LastUpdate  time.Time `json:"dis_lastupdate"`
}

func (d *Diskon) ToResponse() DiskonResponse {
	return DiskonResponse{
		ID:          d.ID,
		OutletID:    d.OutletID,
		Diskon:      d.Diskon,
		Jenis:       d.Jenis,
		NilaiDiskon: d.NilaiDiskon,
		Keterangan:  d.Keterangan,
		Status:      d.Status,
		CreatedAt:   d.CreatedAt,
		LastUpdate:  d.LastUpdate,
	}
}