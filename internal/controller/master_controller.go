package controller

import (
	"BackendFramework/internal/database"
	"BackendFramework/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// --- PARFUM HANDLERS ---

func GetParfums(c *gin.Context) {
	outletID := c.GetUint("outlet_id")
	var parfums []model.Parfume
	database.DbCore.Where("outlet_id = ?", outletID).Find(&parfums)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": parfums})
}

func CreateParfum(c *gin.Context) {
	var parfum model.Parfume
	if err := c.ShouldBindJSON(&parfum); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	parfum.OutletID = c.GetUint("outlet_id")
	database.DbCore.Create(&parfum)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": parfum})
}

func UpdateParfum(c *gin.Context) { {
	id := c.Param("id")
	outletID := c.GetUint("outlet_id")
	var parfum model.Parfume

	if err := database.DbCore.Where("id = ? AND outlet_id = ?", id, outletID).First(&parfum).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Data tidak ditemukan"})
		return
	}

	c.ShouldBindJSON(&parfum)
	database.DbCore.Save(&parfum)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": parfum})
}
}

func DeleteParfum(c *gin.Context) {
	id := c.Param("id")
	outletID := c.GetUint("outlet_id")
	database.DbCore.Where("id = ? AND outlet_id = ?", id, outletID).Delete(&model.Parfume{})
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Berhasil dihapus"})
}

// --- DISCOUNT HANDLERS ---

func GetDiscounts(c *gin.Context) {
	outletID := c.GetUint("outlet_id")
	var discounts []model.Discount
	database.DbCore.Where("outlet_id = ?", outletID).Find(&discounts)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": discounts})
}

func CreateDiscount(c *gin.Context) {
	var discount model.Discount
	if err := c.ShouldBindJSON(&discount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	discount.OutletID = c.GetUint("outlet_id")
	database.DbCore.Create(&discount)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": discount})
}

func UpdateDiscount(c *gin.Context) {
	id := c.Param("id")
	outletID := c.GetUint("outlet_id")
	var discount model.Discount

	if err := database.DbCore.Where("id = ? AND outlet_id = ?", id, outletID).First(&discount).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Data tidak ditemukan"})
		return
	}

	c.ShouldBindJSON(&discount)
	database.DbCore.Save(&discount) // Pakai Save agar bool false ikut terupdate
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Berhasil diupdate"})
}

func DeleteDiscount(c *gin.Context) {
	id := c.Param("id")
	outletID := c.GetUint("outlet_id")
	database.DbCore.Where("id = ? AND outlet_id = ?", id, outletID).Delete(&model.Discount{})
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Berhasil dihapus"})
}