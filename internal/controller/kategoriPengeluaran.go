package controller

import (
    "net/http"
    "strconv"
    "BackendFramework/internal/model"
    "BackendFramework/internal/service"
    "github.com/gin-gonic/gin"
)

type KategoriPengeluaranController struct {
    service *service.KategoriPengeluaranService
}

// Ubah constructor untuk menerima service sebagai parameter
func NewKategoriPengeluaranController(service *service.KategoriPengeluaranService) *KategoriPengeluaranController {
    return &KategoriPengeluaranController{
        service: service,
    }
}

// GetAll godoc
// @Summary Get all kategori pengeluaran
// @Description Get all kategori pengeluaran
// @Tags Kategori Pengeluaran
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /kategori-pengeluaran [get]
func (c *KategoriPengeluaranController) GetAll(ctx *gin.Context) {
    // Ambil outlet_id dari context (dari middleware auth)
    outletID, exists := ctx.Get("outlet_id")
    var outletIDPtr *uint
    if exists && outletID != nil {
        if id, ok := outletID.(uint); ok {
            outletIDPtr = &id
        }
    }

    kategoris, err := c.service.GetAll(outletIDPtr)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Gagal mengambil data kategori pengeluaran",
            "error":   err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Data kategori pengeluaran berhasil diambil",
        "data":    kategoris,
    })
}

// GetByID godoc
// @Summary Get kategori pengeluaran by ID
// @Description Get kategori pengeluaran by ID
// @Tags Kategori Pengeluaran
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Kategori ID"
// @Success 200 {object} map[string]interface{}
// @Router /kategori-pengeluaran/{id} [get]
func (c *KategoriPengeluaranController) GetByID(ctx *gin.Context) {
    id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "ID tidak valid",
        })
        return
    }

    // Ambil outlet_id dari context
    outletID, exists := ctx.Get("outlet_id")
    var outletIDPtr *uint
    if exists && outletID != nil {
        if id, ok := outletID.(uint); ok {
            outletIDPtr = &id
        }
    }

    kategori, err := c.service.GetByID(uint(id), outletIDPtr)
    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{
            "status":  "error",
            "message": err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Data kategori pengeluaran berhasil diambil",
        "data":    kategori,
    })
}

// Create godoc
// @Summary Create new kategori pengeluaran
// @Description Create new kategori pengeluaran
// @Tags Kategori Pengeluaran
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.KategoriPengeluaranInput true "Kategori Input"
// @Success 201 {object} map[string]interface{}
// @Router /kategori-pengeluaran [post]
func (c *KategoriPengeluaranController) Create(ctx *gin.Context) {
    var input model.KategoriPengeluaranInput
    if err := ctx.ShouldBindJSON(&input); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "Input tidak valid",
            "error":   err.Error(),
        })
        return
    }

    // Validasi input
    if input.Kategori == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "Nama kategori tidak boleh kosong",
        })
        return
    }

    // Ambil outlet_id dari context
    outletID, exists := ctx.Get("outlet_id")
    var outletIDPtr *uint
    if exists && outletID != nil {
        if id, ok := outletID.(uint); ok {
            outletIDPtr = &id
        }
    }

    // Ambil username dari context
    username, _ := ctx.Get("username")
    usernameStr, _ := username.(string)

    kategori, err := c.service.Create(input, outletIDPtr, usernameStr)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Gagal membuat kategori pengeluaran",
            "error":   err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusCreated, gin.H{
        "status":  "success",
        "message": "Kategori pengeluaran berhasil dibuat",
        "data":    kategori,
    })
}

// Update godoc
// @Summary Update kategori pengeluaran
// @Description Update kategori pengeluaran
// @Tags Kategori Pengeluaran
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Kategori ID"
// @Param body body model.UpdateKategoriPengeluaranInput true "Update Input"
// @Success 200 {object} map[string]interface{}
// @Router /kategori-pengeluaran/{id} [put]
func (c *KategoriPengeluaranController) Update(ctx *gin.Context) {
    id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "ID tidak valid",
        })
        return
    }

    var input model.UpdateKategoriPengeluaranInput
    if err := ctx.ShouldBindJSON(&input); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "Input tidak valid",
            "error":   err.Error(),
        })
        return
    }

    // Ambil outlet_id dari context
    outletID, exists := ctx.Get("outlet_id")
    var outletIDPtr *uint
    if exists && outletID != nil {
        if outletId, ok := outletID.(uint); ok {
            outletIDPtr = &outletId
        }
    }

    // Ambil username dari context
    username, _ := ctx.Get("username")
    usernameStr, _ := username.(string)

    kategori, err := c.service.Update(uint(id), input, outletIDPtr, usernameStr)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Gagal mengupdate kategori pengeluaran",
            "error":   err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Kategori pengeluaran berhasil diupdate",
        "data":    kategori,
    })
}

// Delete godoc
// @Summary Delete kategori pengeluaran
// @Description Delete kategori pengeluaran
// @Tags Kategori Pengeluaran
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Kategori ID"
// @Success 200 {object} map[string]interface{}
// @Router /kategori-pengeluaran/{id} [delete]
func (c *KategoriPengeluaranController) Delete(ctx *gin.Context) {
    id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "ID tidak valid",
        })
        return
    }

    // Ambil outlet_id dari context
    outletID, exists := ctx.Get("outlet_id")
    var outletIDPtr *uint
    if exists && outletID != nil {
        if outletId, ok := outletID.(uint); ok {
            outletIDPtr = &outletId
        }
    }

    if err := c.service.Delete(uint(id), outletIDPtr); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Gagal menghapus kategori pengeluaran",
            "error":   err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Kategori pengeluaran berhasil dihapus",
    })
}