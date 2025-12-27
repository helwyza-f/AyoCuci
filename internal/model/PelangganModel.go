package model

import (
    "time"
    "gorm.io/gorm"
)

type Customer struct {
    ID           uint           `gorm:"column:cust_id;primaryKey;autoIncrement" json:"cust_id"`
    OutletID     *uint          `gorm:"column:cust_outlet;index" json:"cust_outlet"`
    Nama         string         `gorm:"column:cust_nama;type:varchar(100);not null" json:"cust_nama"`
    Phone        string         `gorm:"column:cust_phone;type:varchar(20);not null" json:"cust_phone"`
    Alamat       string         `gorm:"column:cust_alamat;type:text" json:"cust_alamat"`
    Gender       string         `gorm:"column:cust_gender;type:varchar(10)" json:"cust_gender"`
    TanggalLahir string         `gorm:"column:cust_tanggal_lahir;type:date" json:"cust_tanggal_lahir"`
    CreatedAt    time.Time      `gorm:"column:cust_created;autoCreateTime" json:"cust_created"`
    LastUpdate   time.Time      `gorm:"column:cust_lastupdate;autoUpdateTime" json:"cust_lastupdate"`
    UserUpdate   string         `gorm:"column:cust_userupdate;type:varchar(50)" json:"cust_userupdate"`
    DeletedAt    gorm.DeletedAt `gorm:"column:cust_deleted;index" json:"-"`
}

func (Customer) TableName() string {
    return "ac_customer"
}

type CustomerInput struct {
    Nama         string `json:"nama" binding:"required"`
    Phone        string `json:"phone" binding:"required"`
    Alamat       string `json:"alamat"`
    Gender       string `json:"gender" binding:"omitempty,oneof=Pria Wanita"`
    TanggalLahir string `json:"tanggal_lahir"`
}

type UpdateCustomerInput struct {
    Nama         string `json:"nama"`
    Phone        string `json:"phone"`
    Alamat       string `json:"alamat"`
    Gender       string `json:"gender" binding:"omitempty,oneof=Pria Wanita"`
    TanggalLahir string `json:"tanggal_lahir"`
}

type CustomerResponse struct {
    ID           uint      `json:"cust_id"`
    OutletID     *uint     `json:"cust_outlet,omitempty"`
    Nama         string    `json:"cust_nama"`
    Phone        string    `json:"cust_phone"`
    Alamat       string    `json:"cust_alamat"`
    Gender       string    `json:"cust_gender"`
    TanggalLahir string    `json:"cust_tanggal_lahir"`
    CreatedAt    time.Time `json:"cust_created"`
    LastUpdate   time.Time `json:"cust_lastupdate"`
}

func (c *Customer) ToResponse() CustomerResponse {
    return CustomerResponse{
        ID:           c.ID,
        OutletID:     c.OutletID,
        Nama:         c.Nama,
        Phone:        c.Phone,
        Alamat:       c.Alamat,
        Gender:       c.Gender,
        TanggalLahir: c.TanggalLahir,
        CreatedAt:    c.CreatedAt,
        LastUpdate:   c.LastUpdate,
    }
}