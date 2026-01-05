package controller

import (
	"BackendFramework/internal/database"
	"BackendFramework/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func CreateEmployee(c *gin.Context) {
	var input struct {
		Name        string          `json:"name" binding:"required"`
		Phone       string          `json:"phone" binding:"required"`
		Email       string          `json:"email" binding:"required"`
		Password    string          `json:"password" binding:"required"`
		Permissions map[string]bool `json:"permissions"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	// Hash Password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	outletID := c.GetUint("outlet_id")

	employee := model.Employee{
		OutletID: outletID,
		Name:     input.Name,
		Phone:    input.Phone,
		Email:    input.Email,
		Password: string(hashedPassword),
		Role:     "Pegawai", // Default role

		// Mapping Permissions dari Map ke Struct
		PermMakeOrder:         input.Permissions["make_order"],
		PermCancelOrder:       input.Permissions["cancel_order"],
		PermManageExpenses:    input.Permissions["manage_expenses"],
		PermManageServices:    input.Permissions["manage_services"],
		PermViewRevenue:       input.Permissions["view_revenue"],
		PermManageEmployees:   input.Permissions["manage_employees"],
		PermReportTransaction: input.Permissions["report_transaction"],
		PermReportPerformance: input.Permissions["report_performance"],
		PermReportFinance:     input.Permissions["report_finance"],
		PermReportCustomer:    input.Permissions["report_customer"],
	}

	if err := database.DbCore.Create(&employee).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Email sudah terdaftar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": employee})
}

func GetEmployees(c *gin.Context) {
	outletID, _ := c.Get("outlet_id")
	var employees []model.Employee

	database.DbCore.Where("outlet_id = ?", outletID).Find(&employees)

	// Ubah format untuk Flutter agar menyertakan map permissions
	var response []gin.H
	for _, e := range employees {
		response = append(response, gin.H{
			"id":         e.ID,
			"name":       e.Name,
			"phone":      e.Phone,
			"email":      e.Email,
			"role":       e.Role,
			"created_at": e.CreatedAt,
			"permissions": gin.H{
				"make_order":         e.PermMakeOrder,
				"cancel_order":       e.PermCancelOrder,
				"manage_expenses":    e.PermManageExpenses,
				"manage_services":    e.PermManageServices,
				"view_revenue":       e.PermViewRevenue,
				"manage_employees":   e.PermManageEmployees,
				"report_transaction": e.PermReportTransaction,
				"report_performance": e.PermReportPerformance,
				"report_finance":     e.PermReportFinance,
				"report_customer":    e.PermReportCustomer,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}

func UpdateEmployee(c *gin.Context) {
	id := c.Param("id")
	outletID := c.GetUint("outlet_id")

	var input struct {
		Name        string          `json:"name"`
		Phone       string          `json:"phone"`
		Email       string          `json:"email"`
		Password    string          `json:"password"` // Opsional saat update
		Permissions map[string]bool `json:"permissions"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	// 1. Cari data lama terlebih dahulu
	var employee model.Employee
	if err := database.DbCore.Where("id = ? AND outlet_id = ?", id, outletID).First(&employee).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Pegawai tidak ditemukan"})
		return
	}

	// 2. Update field data pribadi
	employee.Name = input.Name
	employee.Phone = input.Phone
	employee.Email = input.Email

	// 3. Logic Password: Hanya ganti jika password baru tidak kosong
	if input.Password != "" {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		employee.Password = string(hashedPassword)
	}

	// 4. Update Mapping Permissions
	employee.PermMakeOrder = input.Permissions["make_order"]
	employee.PermCancelOrder = input.Permissions["cancel_order"]
	employee.PermManageExpenses = input.Permissions["manage_expenses"]
	employee.PermManageServices = input.Permissions["manage_services"]
	employee.PermViewRevenue = input.Permissions["view_revenue"]
	employee.PermManageEmployees = input.Permissions["manage_employees"]
	employee.PermReportTransaction = input.Permissions["report_transaction"]
	employee.PermReportPerformance = input.Permissions["report_performance"]
	employee.PermReportFinance = input.Permissions["report_finance"]
	employee.PermReportCustomer = input.Permissions["report_customer"]

	// 5. Simpan perubahan (GORM Save akan mengupdate semua field termasuk bool false)
	if err := database.DbCore.Save(&employee).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal update data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Data pegawai diperbarui"})
}

func DeleteEmployee(c *gin.Context) {
	id := c.Param("id")
	outletID := c.GetUint("outlet_id")

	// Pastikan hanya menghapus pegawai milik outlet tersebut
	result := database.DbCore.Where("id = ? AND outlet_id = ?", id, outletID).Delete(&model.Employee{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal menghapus data"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Data tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Pegawai berhasil dihapus"})
}