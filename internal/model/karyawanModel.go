package model

import (
    "encoding/json"
    "time"
    "golang.org/x/crypto/bcrypt"
)

type Karyawan struct {
    ID           uint      `gorm:"column:kar_id;primaryKey;autoIncrement" json:"kar_id"`
    OutletID     *uint     `gorm:"column:kar_outlet;index" json:"kar_outlet"`
    Nama         string    `gorm:"column:kar_nama;type:varchar(100);not null" json:"kar_nama"`
    Phone        string    `gorm:"column:kar_phone;type:varchar(20);not null" json:"kar_phone"`
    Email        string    `gorm:"column:kar_email;type:varchar(100);not null;unique" json:"kar_email"`
    Password     string    `gorm:"column:kar_password;type:varchar(255);not null" json:"-"`
    Role         string    `gorm:"column:kar_role;type:varchar(20);default:'Karyawan'" json:"kar_role"`
    IsPremium    bool      `gorm:"column:kar_is_premium;default:false" json:"kar_is_premium"`
    Permissions  string    `gorm:"column:kar_permissions;type:text" json:"kar_permissions"`
    Status       string    `gorm:"column:kar_status;type:varchar(20);default:'Aktif'" json:"kar_status"`
    JoinDate     time.Time `gorm:"column:kar_join_date;autoCreateTime" json:"kar_join_date"`
    CreatedAt    time.Time `gorm:"column:kar_created;autoCreateTime" json:"kar_created"`
    LastUpdate   time.Time `gorm:"column:kar_lastupdate;autoUpdateTime" json:"kar_lastupdate"`
    UserUpdate   string    `gorm:"column:kar_userupdate;type:varchar(50)" json:"kar_userupdate"`
}

func (Karyawan) TableName() string {
    return "ac_karyawan"
}

type KaryawanInput struct {
    Nama        string   `json:"nama" binding:"required"`
    Phone       string   `json:"phone" binding:"required"`
    Email       string   `json:"email" binding:"required,email"`
    Password    string   `json:"password" binding:"required,min=6"`
    IsPremium   bool     `json:"isPremium"`
    Permissions []string `json:"permissions"`
}

type UpdateKaryawanInput struct {
    Nama        string   `json:"nama"`
    Phone       string   `json:"phone"`
    Email       string   `json:"email" binding:"omitempty,email"`
    Password    string   `json:"password" binding:"omitempty,min=6"`
    IsPremium   bool     `json:"isPremium"`
    Permissions []string `json:"permissions"`
    Status      string   `json:"status" binding:"omitempty,oneof=Aktif 'Tidak Aktif'"`
}

type KaryawanResponse struct {
    ID          uint      `json:"kar_id"`
    OutletID    *uint     `json:"kar_outlet,omitempty"`
    Nama        string    `json:"kar_nama"`
    Phone       string    `json:"kar_phone"`
    Email       string    `json:"kar_email"`
    Role        string    `json:"kar_role"`
    IsPremium   bool      `json:"kar_is_premium"`
    Permissions []string  `json:"kar_permissions"`
    Status      string    `json:"kar_status"`
    JoinDate    time.Time `json:"kar_join_date"`
    CreatedAt   time.Time `json:"kar_created"`
    LastUpdate  time.Time `json:"kar_lastupdate"`
}

func (k *Karyawan) ToResponse() KaryawanResponse {
    permissions := []string{}
    if k.Permissions != "" {
        json.Unmarshal([]byte(k.Permissions), &permissions)
    }

    return KaryawanResponse{
        ID:          k.ID,
        OutletID:    k.OutletID,
        Nama:        k.Nama,
        Phone:       k.Phone,
        Email:       k.Email,
        Role:        k.Role,
        IsPremium:   k.IsPremium,
        Permissions: permissions,
        Status:      k.Status,
        JoinDate:    k.JoinDate,
        CreatedAt:   k.CreatedAt,
        LastUpdate:  k.LastUpdate,
    }
}

func (k *Karyawan) HashPassword(password string) error {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    k.Password = string(hashedPassword)
    return nil
}

func (k *Karyawan) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(k.Password), []byte(password))
    return err == nil
}

var AvailablePermissions = []string{
    "Membuat Order / Transaksi",
    "Menambahkan Order / Transaksi",
    "Membuat Pengaturan",
    "Mengelola Layanan / Produk",
    "Menampilkan Nilai Omzet",
    "Mengelola Data Karyawan",
    "Akses Layanan Transaksi",
    "Akses Layanan Konsep",
    "Akses Layanan Keuangan",
    "Akses Layanan Pelanggan",
}