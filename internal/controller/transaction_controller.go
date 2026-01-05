package controller

import (
	"BackendFramework/internal/database"
	"BackendFramework/internal/model"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateTransaction(c *gin.Context) {
	var input struct {
		CustomerID uint    `json:"customer_id" binding:"required"`
		ParfumID   uint    `json:"parfum_id"`
		DiscountID *uint   `json:"discount_id"`
		TotalPrice float64 `json:"total_price" binding:"required"`
		Notes      string  `json:"notes"`
		Items      []struct {
			ServiceName string  `json:"service_name"`
			Price       float64 `json:"price"`
			Qty         float64 `json:"qty"`
		} `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	outletID := c.GetUint("outlet_id")
	adminName := "Admin" // Nanti ambil dari JWT payload

	// Generate Nomor Invoice Sederhana: TRX-WaktuUnix
	invoice := fmt.Sprintf("TRX/%d/%d", outletID, time.Now().Unix())

	tx := database.DbCore.Begin()

	// 1. Simpan Header
	transaction := model.Transaction{
		InvoiceNumber: invoice,
		OutletID:      outletID,
		CustomerID:    input.CustomerID,
		ParfumID:      input.ParfumID,
		DiscountID:    input.DiscountID,
		TotalPrice:    input.TotalPrice,
		Notes:         input.Notes,
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal buat transaksi"})
		return
	}

	// 2. Simpan Items
	for _, item := range input.Items {
		detail := model.TransactionDetail{
			TransactionID: transaction.ID,
			ServiceName:   item.ServiceName,
			Price:         item.Price,
			Qty:           item.Qty,
			Subtotal:      item.Price * item.Qty,
		}
		tx.Create(&detail)
	}

	// 3. Simpan Log Awal (Antrian)
	log := model.OrderLog{
		TransactionID: transaction.ID,
		Status:        "Antrian",
		AdminName:     adminName,
	}
	tx.Create(&log)

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"success": true, "data": transaction})
}

func GetTransactions(c *gin.Context) {
	outletID := c.GetUint("outlet_id")
	var transactions []model.Transaction

	// Preload Customer dan Items agar muncul lengkap di UI
	database.DbCore.Where("outlet_id = ?", outletID).
		Preload("Customer").
		Preload("Items").
		Order("created_at desc").
		Find(&transactions)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": transactions})
}