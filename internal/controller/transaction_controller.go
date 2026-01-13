package controller

import (
	"BackendFramework/internal/database"
	"BackendFramework/internal/model"
	"BackendFramework/internal/thirdparty"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Helper untuk mendapatkan waktu Jakarta secara konsisten
func getJakartaTime() time.Time {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback jika tzdata tidak tersedia di server
		return time.Now().Add(7 * time.Hour)
	}
	return time.Now().In(loc)
}

func CreateTransaction(c *gin.Context) {
	var input struct {
		CustomerID uint    `json:"customer_id" binding:"required"`
	ParfumID   uint    `json:"parfum_id"` // Dari JSON tetap uint
        DiscountID uint    `json:"discount_id"`
		TotalPrice float64 `json:"total_price" binding:"required"`
		Notes      string  `json:"notes"`
		Items      []struct {
			ServiceName  string  `json:"service_name"`
			Price        float64 `json:"price"`
			Qty          float64 `json:"qty"`
			Unit         string  `json:"unit"`
			IconPath     string  `json:"icon_path"`
			Duration     int     `json:"duration"`
			DurationUnit string  `json:"duration_unit"`
		} `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}


    var parfumID *uint
    if input.ParfumID != 0 {
        parfumID = &input.ParfumID
    }

    var discountID *uint
    if input.DiscountID != 0 {
        discountID = &input.DiscountID
    }
	// Inisialisasi waktu masuk Jakarta
	nowJakarta := getJakartaTime()
	outletID := c.GetUint("outlet_id")
	invoice := fmt.Sprintf("TRX/%d/%d", outletID, nowJakarta.Unix())
	
	tx := database.DbCore.Begin()

	transaction := model.Transaction{
		InvoiceNumber: invoice,
		OutletID:      outletID,
		CustomerID:    input.CustomerID,
		ParfumID:      parfumID,
		DiscountID:    discountID,
		TotalPrice:    input.TotalPrice,
		Notes:         input.Notes,
		CreatedAt:     nowJakarta,
		UpdatedAt:     nowJakarta,
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal buat transaksi"})
		return
	}

	// Loop Items dengan kalkulasi ready_at (jatuh tempo) Jakarta
	for _, item := range input.Items {
		readyAt := nowJakarta

		switch item.DurationUnit {
case "Hari":
			readyAt = readyAt.AddDate(0, 0, item.Duration)
		case "Jam":
			readyAt = readyAt.Add(time.Duration(item.Duration) * time.Hour)
		}

		detail := model.TransactionDetail{
			TransactionID: transaction.ID,
			ServiceName:   item.ServiceName,
			Price:         item.Price,
			Qty:           item.Qty,
			Unit:          item.Unit,
			IconPath:      item.IconPath,
			Subtotal:      item.Price * item.Qty,
			OrderStatus:   "Antrian",
			PaymentStatus: "Belum Bayar",
			ReadyAt:       &readyAt,
			CreatedAt:     nowJakarta,
		}
		tx.Create(&detail)
	}

	// Simpan Log Awal
	adminName := c.GetString("username")
	tx.Create(&model.OrderLog{
		TransactionID: transaction.ID,
		Status:        "Antrian",
		AdminName:     adminName,
		CreatedAt:     nowJakarta,
	})

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"success": true, "data": transaction})
}

func GetTransactions(c *gin.Context) {
	outletID := c.GetUint("outlet_id")
	var transactions []model.Transaction
	database.DbCore.Where("outlet_id = ?", outletID).
		Preload("Customer").Preload("Items").
		Order("created_at desc").Find(&transactions)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": transactions})
}

func GetTransactionDetail(c *gin.Context) {
    id := c.Param("id")
    outletID := c.GetUint("outlet_id")
    var transaction model.Transaction

    // Tambahkan Preload Parfum dan Discount
    err := database.DbCore.Where("id = ? AND outlet_id = ?", id, outletID).
        Preload("Customer").
        Preload("Items").
        Preload("Logs").
        Preload("Parfum").   // Preload data Parfum
        Preload("Discount"). // Preload data Diskon
        First(&transaction).Error

    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Pesanan tidak ditemukan"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"success": true, "data": transaction})
}

// UpdateStatusItem - Mendukung Pengambilan Parsial
func UpdateStatusItem(c *gin.Context) {
	itemID := c.Param("item_id")
	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DbCore.Model(&model.TransactionDetail{}).
		Where("id = ?", itemID).Update("order_status", input.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal update item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Status item diperbarui"})
}

// UpdateStatus - Update Status Global dengan Log Jakarta
// internal/controller/transaction_controller.go

func UpdateStatus(c *gin.Context) {
    id := c.Param("id")
    var input struct {
        Status    string `json:"status" binding:"required"`
        AdminName string `json:"admin_name"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var transaction model.Transaction
    database.DbCore.First(&transaction, id)

    // VALIDASI PEMBAYARAN: Jika mau Selesai, harus Lunas dulu
    if input.Status == "Selesai" && transaction.PaymentStatus != "Lunas" {
        c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "Pesanan belum lunas! Silahkan lakukan pembayaran."})
        return
    }

    nowJakarta := getJakartaTime()
    tx := database.DbCore.Begin()

    // 1. Update Header Status
    tx.Model(&transaction).Update("order_status", input.Status)

    // 2. Logika Sinkronisasi Item
    if input.Status == "Proses" {
        tx.Model(&model.TransactionDetail{}).Where("transaction_id = ?", id).Update("order_status", "Proses")
    } else if input.Status == "Selesai" {
        // Jika Selesai, pastikan SEMUA item berstatus Selesai dan Payment Lunas
        tx.Model(&model.TransactionDetail{}).Where("transaction_id = ?", id).
            Updates(map[string]interface{}{
                "order_status":   "Selesai",
                "payment_status": "Lunas",
            })
    }

    tx.Create(&model.OrderLog{TransactionID: transaction.ID, Status: input.Status, AdminName: input.AdminName, CreatedAt: nowJakarta})
    tx.Commit()
    
    c.JSON(http.StatusOK, gin.H{"success": true})
}

func ProcessPayment(c *gin.Context) {
	id := c.Param("id")
	nowJakarta := getJakartaTime()
	
	var input struct {
		PaymentMethod string `json:"payment_method" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.DbCore.Model(&model.Transaction{}).Where("id = ?", id).Updates(map[string]interface{}{
		"payment_status": "Lunas",
		"notes":          fmt.Sprintf("Bayar via %s", input.PaymentMethod),
		"updated_at":     nowJakarta,
	})
	
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Pembayaran Berhasil"})
}

func triggerFonnte(id string) error {
	var trx model.Transaction
	database.DbCore.Preload("Customer").First(&trx, id)
	
	msg := fmt.Sprintf("Halo %s,\n\nLaundry Anda dengan nota *%s* sudah selesai dan *Siap untuk diambil*.\nTotal: *IDR %v*\n\nTerima Kasih.", 
		trx.Customer.Name, trx.InvoiceNumber, trx.TotalPrice)
	
	thirdparty.SendFonnteNotification(trx.Customer.Phone, msg)
	return nil
}

func SendManualNotification(c *gin.Context) {
	id := c.Param("id")
	if err := triggerFonnte(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal mengirim notifikasi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}