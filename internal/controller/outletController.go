package controller

import (
	"BackendFramework/internal/database"
	"BackendFramework/internal/model"
	"BackendFramework/internal/service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateOutletController(c *gin.Context) {
    userIDInterface, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "error",
            "message": "Unauthorized",
        })
        return
    }

    var userID uint
    switch v := userIDInterface.(type) {
    case uint:
        userID = v
    case string:
        id, err := strconv.ParseUint(v, 10, 32)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{
                "status":  "error",
                "message": "Invalid user ID format",
            })
            return
        }
        userID = uint(id)
    default:
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "error",
            "message": "Invalid user ID type",
        })
        return
    }

    fmt.Printf("=== CREATE OUTLET REQUEST ===\n")
    fmt.Printf("User ID: %d\n", userID)

    if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
        fmt.Printf("Error parsing form: %v\n", err)
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "Gagal parse form data",
        })
        return
    }

    var input model.OutletInput
    input.NamaOutlet = c.PostForm("nama_outlet")
    input.Alamat = c.PostForm("alamat")
    input.Provinsi = c.PostForm("provinsi")
    input.Kota = c.PostForm("kota")
    input.Kecamatan = c.PostForm("kecamatan")
    input.NomorHP = c.PostForm("nomor_hp")

    fmt.Printf("Received data:\n")
    fmt.Printf("  Nama Outlet: %s\n", input.NamaOutlet)
    fmt.Printf("  Alamat: %s\n", input.Alamat)
    fmt.Printf("  Provinsi: %s\n", input.Provinsi)
    fmt.Printf("  Kota: %s\n", input.Kota)
    fmt.Printf("  Kecamatan: %s\n", input.Kecamatan)
    fmt.Printf("  Nomor HP: %s\n", input.NomorHP)

    if input.NamaOutlet == "" || input.Alamat == "" || input.NomorHP == "" || 
       input.Provinsi == "" || input.Kota == "" || input.Kecamatan == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "Semua field wajib diisi",
        })
        return
    }

    file, _ := c.FormFile("photo")
    if file != nil {
        fmt.Printf("Photo file: %s (size: %d bytes)\n", file.Filename, file.Size)
    }

    outlet, err := service.CreateOutlet(userID, input, file)
    if err != nil {
        fmt.Printf("ERROR creating outlet: %v\n", err)
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": err.Error(),
        })
        return
    }

    fmt.Printf("SUCCESS - Outlet created with ID: %d\n", outlet.ID)
    fmt.Printf("=== END CREATE OUTLET ===\n\n")

    c.JSON(http.StatusCreated, gin.H{
        "status":  "success",
        "message": "Outlet berhasil dibuat",
        "data":    outlet,
    })
}

// internal/controller/outlet_controller.go

func CreateOutlet(c *gin.Context) {
    var input model.OutletInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid"})
        return
    }

    // 1. Ambil userID dari context (di sini tipenya interface{} berisi string)
    userIDInterface, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak teridentifikasi"})
        return
    }

    // 2. Konversi interface{} ke string, lalu string ke uint
    userIDStr := userIDInterface.(string)
    userIDUint, err := strconv.ParseUint(userIDStr, 10, 32)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Format User ID salah"})
        return
    }

    newOutlet := model.Outlet{
        UserID:     uint(userIDUint), // Gunakan hasil konversi yang sudah jadi uint
        NamaOutlet: input.NamaOutlet,
        Alamat:     input.Alamat,
        NomorHP:    input.NomorHP,
        Provinsi:   input.Provinsi, // Pastikan field ini ada di model
        Kota:       input.Kota,     // Pastikan field ini ada di model
        Kecamatan:  input.Kecamatan, // Pastikan field ini ada di model
        IsAktif:    "active",
    }

    if err := database.DbCore.Create(&newOutlet).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan outlet"})
        return
    }

	err = database.DbCore.Model(&model.User{}).Where("id = ?", userIDUint).Update("outlet_id", newOutlet.ID).Error
    if err != nil {
        // Log error saja, jangan gagalkan proses karena outlet sudah terlanjur dibuat
        fmt.Println("Gagal update outlet_id di user:", err)
    }

    c.JSON(http.StatusCreated, gin.H{
        "message": "Outlet berhasil dibuat",
        "outlet_id": newOutlet.ID,
    })
}


func getUserIDFromContext(c *gin.Context) (uint, error) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		return 0, fmt.Errorf("user ID not found in context")
	}

	switch v := userIDInterface.(type) {
	case uint:
		return v, nil
	case string:
		id, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return 0, err
		}
		return uint(id), nil
	default:
		return 0, fmt.Errorf("invalid user ID type")
	}
}

func GetMyOutletsController(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	outlets, err := service.GetOutletsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Outlet berhasil diambil",
		"data":    outlets,
	})
}

func GetAllOutletsController(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	outlets, total, err := service.GetAllOutlets(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Outlet berhasil diambil",
		"data": gin.H{
			"outlets": outlets,
			"pagination": gin.H{
				"page":       page,
				"limit":      limit,
				"total":      total,
				"total_page": (total + int64(limit) - 1) / int64(limit),
			},
		},
	})
}

func GetOutletByIDController(c *gin.Context) {
	outletID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "ID outlet tidak valid",
		})
		return
	}

	outlet, err := service.GetOutletByID(uint(outletID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Outlet berhasil diambil",
		"data":    outlet,
	})
}

func UpdateOutletController(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	outletID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "ID outlet tidak valid",
		})
		return
	}

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Gagal parse form data",
		})
		return
	}

	var input model.UpdateOutletInput

	input.NamaOutlet = c.PostForm("nama_outlet")
	input.Alamat = c.PostForm("alamat")
	input.Provinsi = c.PostForm("provinsi")
	input.Kota = c.PostForm("kota")
	input.Kecamatan = c.PostForm("kecamatan")
	input.NomorHP = c.PostForm("nomor_hp")
	file, _ := c.FormFile("photo")

	outlet, err := service.UpdateOutlet(uint(outletID), userID, input, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Outlet berhasil diupdate",
		"data":    outlet,
	})
}

func DeleteOutletController(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	outletID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "ID outlet tidak valid",
		})
		return
	}

	if err := service.DeleteOutlet(uint(outletID), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Outlet berhasil dihapus",
	})
}

func ActivateOutletController(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	outletID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "ID outlet tidak valid",
		})
		return
	}

	if err := service.ActivateOutlet(uint(outletID), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Outlet berhasil diaktifkan",
	})
}

func DeactivateOutletController(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	outletID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "ID outlet tidak valid",
		})
		return
	}

	if err := service.DeactivateOutlet(uint(outletID), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Outlet berhasil dinonaktifkan",
	})
}
