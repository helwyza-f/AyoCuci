package controller

import (
	"BackendFramework/internal/database"
	"BackendFramework/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateCustomer menangani pendaftaran pelanggan baru
func CreateCustomer(c *gin.Context) {
	var input struct {
		Name    string  `json:"name" binding:"required"`
		Phone   string  `json:"phone" binding:"required"`
		Gender  string  `json:"gender"`
		OTP     *string `json:"otp"`
		Address string  `json:"address"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	outletID := c.GetUint("outlet_id")

	customer := model.Customer{
		OutletID: outletID,
		Name:     input.Name,
		Phone:    input.Phone,
		Gender:   input.Gender,
		OTP:      input.OTP,
		Address:  input.Address,
	}

	if err := database.DbCore.Create(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal menyimpan data pelanggan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": customer})
}

// GetCustomers mengambil semua daftar pelanggan di suatu outlet
func GetCustomers(c *gin.Context) {
	outletID, _ := c.Get("outlet_id")
	var customers []model.Customer

	err := database.DbCore.
		Where("outlet_id = ?", outletID).
		Order("id desc"). // Menampilkan data terbaru di atas
		Find(&customers).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal mengambil data pelanggan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": customers})
}

// UpdateCustomer untuk memperbarui informasi pelanggan
func UpdateCustomer(c *gin.Context) {
	customerID := c.Param("id")
	outletID, _ := c.Get("outlet_id")

	var input struct {
		Name    string  `json:"name"`
		Phone   string  `json:"phone"`
		Gender  string  `json:"gender"`
		Address string  `json:"address"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result := database.DbCore.Model(&model.Customer{}).
		Where("id = ? AND outlet_id = ?", customerID, outletID).
		Updates(model.Customer{
			Name:    input.Name,
			Phone:   input.Phone,
			Gender:  input.Gender,
			Address: input.Address,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal update data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteCustomer menghapus data pelanggan
func DeleteCustomer(c *gin.Context) {
	customerID := c.Param("id")
	outletID, _ := c.Get("outlet_id")

	err := database.DbCore.
		Where("id = ? AND outlet_id = ?", customerID, outletID).
		Delete(&model.Customer{}).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal menghapus pelanggan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Pelanggan berhasil dihapus"})
}