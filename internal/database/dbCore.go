package database

import (
	"BackendFramework/internal/config"
	"BackendFramework/internal/model"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DbCore *gorm.DB

func OpenAkademik() {
	var err error

	log.Printf("DB Config - Username: '%s', Hostname: '%s', DBName: '%s'",
		config.DB_CORE_USERNAME, config.DB_CORE_HOSTNAME, config.DB_CORE_DBNAME)

	dsn := config.DB_CORE_USERNAME + ":" + config.DB_CORE_PASSWORD +
		"@tcp(" + config.DB_CORE_HOSTNAME + ")/" + config.DB_CORE_DBNAME +
		"?charset=utf8mb4&parseTime=True&loc=Local"

	DbCore, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal konek ke database: %v", err)
	}

	log.Println("Berhasil konek ke database MySQL")
	err = DbCore.AutoMigrate(
		&model.User{},
		&model.Outlet{},
		&model.OTP{},
		&model.LoginAttempt{},
		&model.Layanan{},
		&model.JenisProduk{},
		&model.Diskon{},
		&model.Parfum{},
		&model.KategoriPengeluaran{},
		&model.Pengeluaran{},
		&model.PaymentMethod{},
		&model.NotaSettings{},
		&model.NotaData{},
		&model.Karyawan{},
		&model.Customer{},
		&model.ServiceCategory{},
		&model.ServiceProduct{},
		&model.Employee{},
		&model.Parfume{},
		&model.Discount{},
		&model.Transaction{},
		&model.OrderLog{},
		&model.TransactionDetail{},

	)
	if err != nil {
		log.Fatalf("Gagal migrasi tabel: %v", err)
	}

	log.Println("Migrasi database berhasil")
}
