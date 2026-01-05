package controller

import (
	"BackendFramework/internal/database"
	"BackendFramework/internal/model"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

func CreateService(c *gin.Context) {
	var input struct {
		Name      string   `json:"name" binding:"required"`
		Processes []string `json:"processes"`
		Items     []struct {
			Name         string  `json:"name"`
			Price        float64 `json:"price"`
			Unit         string  `json:"unit"`
			Duration     int     `json:"duration"`
			DurationUnit string  `json:"duration_unit"`
			IconPath     string  `json:"icon_path"`
			Note 	   	*string  `json:"note"`
		} `json:"items"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	// Simulasi ambil outlet_id dari context (setelah middleware auth)
	outletID := c.GetUint("outlet_id")

	// Map input ke database model
	category := model.ServiceCategory{
		OutletID:  outletID,
		Name:      input.Name,
		Processes: strings.Join(input.Processes, ","),
	}

	// Mulai Transaksi Database
	tx := database.DbCore.Begin()

	if err := tx.Create(&category).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal buat kategori"})
		return
	}

	// Simpan Produk/Item Layanan
	for _, item := range input.Items {
		product := model.ServiceProduct{
			CategoryID:   category.ID,
			Name:         item.Name,
			Price:        item.Price,
			Unit:         item.Unit,
			Duration:     item.Duration,
			DurationUnit: item.DurationUnit,
			IconPath:     item.IconPath,
			Note:		  item.Note,
		}
		if err := tx.Create(&product).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal buat item"})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"success": true, "data": category})
}

func GetServices(c *gin.Context) {
	// Ambil ID outlet dari context (hasil middleware auth)
	outletID, _ := c.Get("outlet_id") 
	
	var categories []model.ServiceCategory

	// Wajib pakai Preload("Products") agar data sub-layanan tidak null
	err := database.DbCore.
		Preload("Products").
		Where("outlet_id = ?", outletID).
		Find(&categories).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false, 
			"message": "Gagal mengambil data layanan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true, 
		"data": categories,
	})
}

func DeleteService(c *gin.Context) {
	categoryID := c.Param("id")
	outletID, _ := c.Get("outlet_id")

	tx := database.DbCore.Begin()

	// 1. Hapus sub-item/products terlebih dahulu
	if err := tx.Where("category_id = ?", categoryID).Delete(&model.ServiceProduct{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal hapus item"})
		return
	}

	// 2. Hapus kategori utama
	if err := tx.Where("id = ? AND outlet_id = ?", categoryID, outletID).Delete(&model.ServiceCategory{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal hapus kategori"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Layanan berhasil dihapus"})
}

func UpdateService(c *gin.Context) {
	categoryID := c.Param("id")
	outletID, _ := c.Get("outlet_id")
	var input struct {
		Name      string   `json:"name"`
		Processes []string `json:"processes"`
		Items     []model.ServiceProduct `json:"items"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	tx := database.DbCore.Begin()

	// 1. Update Header Kategori
	if err := tx.Model(&model.ServiceCategory{}).
		Where("id = ? AND outlet_id = ?", categoryID, outletID).
		Updates(map[string]interface{}{
			"name":      input.Name,
			"processes": strings.Join(input.Processes, ","),
		}).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"success": false, "error": "Gagal update kategori"})
		return
	}

	// 2. Sinkronisasi Produk (Cara termudah: Hapus lama, masukkan baru)
	tx.Where("category_id = ?", categoryID).Delete(&model.ServiceProduct{})
	for _, item := range input.Items {
		item.CategoryID = cast.ToUint(categoryID)
		if err := tx.Create(&item).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"success": false, "error": "Gagal update item"})
			return
		}
	}

	tx.Commit()
	c.JSON(200, gin.H{"success": true})
}