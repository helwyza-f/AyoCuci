package model

import (
    "time"
)

type Pengeluaran struct {
    ID                    uint                  `gorm:"column:pengeluaran_id;primaryKey;autoIncrement" json:"pengeluaran_id"`
    OutletID              *uint                 `gorm:"column:pengeluaran_outlet;index" json:"pengeluaran_outlet"`
    Outlet                *Outlet               `gorm:"foreignKey:OutletID;references:ID" json:"outlet,omitempty"`
    KategoriPengeluaranID *uint                 `gorm:"column:pengeluaran_kategori;index" json:"pengeluaran_kategori"`
    KategoriPengeluaran   *KategoriPengeluaran  `gorm:"foreignKey:KategoriPengeluaranID;references:ID" json:"kategori_pengeluaran,omitempty"`
    Tanggal               time.Time             `gorm:"column:pengeluaran_tanggal;type:date;not null;index" json:"pengeluaran_tanggal"`
    Nominal               int64                 `gorm:"column:pengeluaran_nominal;not null" json:"pengeluaran_nominal"`
    Keterangan            string                `gorm:"column:pengeluaran_keterangan;type:text" json:"pengeluaran_keterangan"`
    Status                string                `gorm:"column:pengeluaran_status;type:varchar(20);default:'Aktif'" json:"pengeluaran_status"`
    CreatedAt             time.Time             `gorm:"column:pengeluaran_created;autoCreateTime" json:"pengeluaran_created"`
    LastUpdate            time.Time             `gorm:"column:pengeluaran_lastupdate;autoUpdateTime" json:"pengeluaran_lastupdate"`
    UserUpdate            string                `gorm:"column:pengeluaran_userupdate;type:varchar(50)" json:"pengeluaran_userupdate"`
}

func (Pengeluaran) TableName() string {
    return "ac_pengeluaran"
}

type PengeluaranInput struct {
    KategoriPengeluaranID uint   `json:"pengeluaran_kategori" binding:"required"`
    Tanggal               string `json:"pengeluaran_tanggal" binding:"required"`
    Nominal               int64  `json:"pengeluaran_nominal" binding:"required,gt=0"`
    Keterangan            string `json:"pengeluaran_keterangan"`
}

type UpdatePengeluaranInput struct {
    KategoriPengeluaranID *uint  `json:"pengeluaran_kategori"`
    Tanggal               string `json:"pengeluaran_tanggal"`
    Nominal               *int64 `json:"pengeluaran_nominal" binding:"omitempty,gt=0"`
    Keterangan            string `json:"pengeluaran_keterangan"`
    Status                string `json:"pengeluaran_status" binding:"omitempty,oneof=Aktif 'Tidak Aktif'"`
}

type PengeluaranResponse struct {
    ID                    uint      `json:"pengeluaran_id"`
    OutletID              *uint     `json:"pengeluaran_outlet,omitempty"`
    KategoriPengeluaranID *uint     `json:"pengeluaran_kategori"`
    KategoriNama          string    `json:"kategori_nama,omitempty"`
    Tanggal               string    `json:"pengeluaran_tanggal"`
    Nominal               int64     `json:"pengeluaran_nominal"`
    Keterangan            string    `json:"pengeluaran_keterangan"`
    Status                string    `json:"pengeluaran_status"`
    CreatedAt             time.Time `json:"pengeluaran_created"`
    LastUpdate            time.Time `json:"pengeluaran_lastupdate"`
}

type PengeluaranDetailResponse struct {
    ID                  uint                         `json:"pengeluaran_id"`
    OutletID            *uint                        `json:"pengeluaran_outlet,omitempty"`
    Outlet              *Outlet                      `json:"outlet,omitempty"`
    KategoriPengeluaran *KategoriPengeluaranResponse `json:"kategori_pengeluaran,omitempty"`
    Tanggal             string                       `json:"pengeluaran_tanggal"`
    Nominal             int64                        `json:"pengeluaran_nominal"`
    Keterangan          string                       `json:"pengeluaran_keterangan"`
    Status              string                       `json:"pengeluaran_status"`
    CreatedAt           time.Time                    `json:"pengeluaran_created"`
    LastUpdate          time.Time                    `json:"pengeluaran_lastupdate"`
}

func (p *Pengeluaran) ToResponse() PengeluaranResponse {
    response := PengeluaranResponse{
        ID:                    p.ID,
        OutletID:              p.OutletID,
        KategoriPengeluaranID: p.KategoriPengeluaranID,
        Tanggal:               p.Tanggal.Format("2006-01-02"), 
        Nominal:               p.Nominal,
        Keterangan:            p.Keterangan,
        Status:                p.Status,
        CreatedAt:             p.CreatedAt,
        LastUpdate:            p.LastUpdate,
    }

    if p.KategoriPengeluaran != nil {
        response.KategoriNama = p.KategoriPengeluaran.Kategori
    }

    return response
}

func (p *Pengeluaran) ToDetailResponse() PengeluaranDetailResponse {
    response := PengeluaranDetailResponse{
        ID:         p.ID,
        OutletID:   p.OutletID,
        Tanggal:    p.Tanggal.Format("2006-01-02"), 
        Nominal:    p.Nominal,
        Keterangan: p.Keterangan,
        Status:     p.Status,
        CreatedAt:  p.CreatedAt,
        LastUpdate: p.LastUpdate,
    }

    if p.Outlet != nil {
        response.Outlet = p.Outlet
    }

    if p.KategoriPengeluaran != nil {
        kategoriResponse := p.KategoriPengeluaran.ToResponse()
        response.KategoriPengeluaran = &kategoriResponse
    }

    return response
}

type PengeluaranSummary struct {
    TotalPengeluaran int64                    `json:"total_pengeluaran"`
    JumlahTransaksi  int64                    `json:"jumlah_transaksi"`
    PerKategori      []PengeluaranPerKategori `json:"per_kategori"`
}

type PengeluaranPerKategori struct {
    KategoriID   uint   `json:"kategori_id"`
    KategoriNama string `json:"kategori_nama"`
    Total        int64  `json:"total"`
    Jumlah       int64  `json:"jumlah"`
}