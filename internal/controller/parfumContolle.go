
package controller

import (
	"BackendFramework/internal/model"
	"BackendFramework/internal/service"
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
)

type ParfumController struct {
	parfumService service.ParfumService
}

func NewParfumController(parfumService service.ParfumService) *ParfumController {
	return &ParfumController{parfumService: parfumService}
}


func (pc *ParfumController) GetAllParfum(c *gin.Context) {
	outletID, exists := c.Get("outlet_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Outlet ID tidak ditemukan",
		})
		return
	}

	parfums, err := pc.parfumService.GetAllParfum(outletID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil data parfum",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data parfum berhasil diambil",
		"data":    parfums,
	})
}

func (pc *ParfumController) GetParfumByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "ID parfum tidak valid",
		})
		return
	}

	outletID, exists := c.Get("outlet_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Outlet ID tidak ditemukan",
		})
		return
	}

	parfum, err := pc.parfumService.GetParfumByID(uint(id), outletID.(uint))
	if err != nil {
		if err.Error() == "parfum tidak ditemukan" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil data parfum",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data parfum berhasil diambil",
		"data":    parfum,
	})
}

func (pc *ParfumController) CreateParfum(c *gin.Context) {
	var input model.ParfumInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Data input tidak valid",
			"error":   err.Error(),
		})
		return
	}

	outletID, exists := c.Get("outlet_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Outlet ID tidak ditemukan",
		})
		return
	}

	username, exists := c.Get("username")
	if !exists {
		username = "system"
	}

	parfum, err := pc.parfumService.CreateParfum(input, outletID.(uint), username.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal membuat parfum",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Parfum berhasil ditambahkan",
		"data":    parfum,
	})
}

func (pc *ParfumController) UpdateParfum(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "ID parfum tidak valid",
		})
		return
	}

	var input model.UpdateParfumInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Data input tidak valid",
			"error":   err.Error(),
		})
		return
	}

	outletID, exists := c.Get("outlet_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Outlet ID tidak ditemukan",
		})
		return
	}

	username, exists := c.Get("username")
	if !exists {
		username = "system"
	}

	parfum, err := pc.parfumService.UpdateParfum(uint(id), input, outletID.(uint), username.(string))
	if err != nil {
		if err.Error() == "parfum tidak ditemukan" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengupdate parfum",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Parfum berhasil diupdate",
		"data":    parfum,
	})
}

func (pc *ParfumController) DeleteParfum(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "ID parfum tidak valid",
		})
		return
	}

	outletID, exists := c.Get("outlet_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Outlet ID tidak ditemukan",
		})
		return
	}

	err = pc.parfumService.DeleteParfum(uint(id), outletID.(uint))
	if err != nil {
		if err.Error() == "parfum tidak ditemukan" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal menghapus parfum",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Parfum berhasil dihapus",
	})
}