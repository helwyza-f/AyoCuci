package controller

import (
	"BackendFramework/internal/model"
	"BackendFramework/internal/service"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

type LayananController struct {
	Service   *service.LayananService
	Validator *validator.Validate
}

func NewLayananController(service *service.LayananService) *LayananController {
	return &LayananController{
		Service:   service,
		Validator: validator.New(),
	}
}

func (c *LayananController) getUserID(ctx *gin.Context) (uint, error) {
	userIDInterface, exists := ctx.Get("userID")
	if !exists {
		return 0, fmt.Errorf("user ID not found in context")
	}

	switch v := userIDInterface.(type) {
	case uint:
		return v, nil
	case int:
		return uint(v), nil
	case float64:
		return uint(v), nil
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, err
		}
		return uint(parsed), nil
	default:
		return 0, fmt.Errorf("invalid user ID type")
	}
}

func (c *LayananController) parseFormData(ctx *gin.Context) (*model.CreateLayananWithProductsInput, error) {
	outletIDStr := ctx.PostForm("ln_outlet")
	namaLayanan := ctx.PostForm("ln_layanan")
	prioritasStr := ctx.PostForm("ln_prioritas")

	if outletIDStr == "" || namaLayanan == "" {
		return nil, fmt.Errorf("outlet ID dan nama layanan wajib diisi")
	}

	outletID, err := strconv.ParseUint(outletIDStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("outlet ID tidak valid")
	}

	prioritas := 0
	if prioritasStr != "" {
		prioritas, _ = strconv.Atoi(prioritasStr)
	}

	return &model.CreateLayananWithProductsInput{
		OutletID:    uint(outletID),
		NamaLayanan: namaLayanan,
		Prioritas:   prioritas,
		Cuci:        ctx.PostForm("ln_cuci"),
		Kering:      ctx.PostForm("ln_kering"),
		Setrika:     ctx.PostForm("ln_setrika"),
	}, nil
}

func (c *LayananController) parseProducts(jenisProdukJSON string) ([]model.CreateJenisProdukInput, error) {
	if jenisProdukJSON == "" {
		return nil, fmt.Errorf("jenis_produk tidak boleh kosong")
	}

	var jenisProdukArray []map[string]interface{}
	if err := json.Unmarshal([]byte(jenisProdukJSON), &jenisProdukArray); err != nil {
		return nil, fmt.Errorf("format JSON tidak valid: %v", err)
	}

	products := make([]model.CreateJenisProdukInput, 0, len(jenisProdukArray))

	for i, produkData := range jenisProdukArray {
		nama, ok := produkData["jp_nama"].(string)
		if !ok || nama == "" {
			return nil, fmt.Errorf("produk %d: nama tidak valid", i)
		}

		produkInput := model.CreateJenisProdukInput{
			Nama: nama,
		}

		// Parse optional fields
		if satuan, ok := produkData["jp_satuan"].(string); ok && satuan != "" {
			produkInput.Satuan = &satuan
		}

		if harga, ok := produkData["jp_harga_per"].(float64); ok {
			hargaInt := int(harga)
			produkInput.HargaPer = &hargaInt
		} else if hargaStr, ok := produkData["jp_harga_per"].(string); ok && hargaStr != "" {
			if hargaInt, err := strconv.Atoi(hargaStr); err == nil {
				produkInput.HargaPer = &hargaInt
			}
		}

		if lama, ok := produkData["jp_lama_pengerjaan"].(float64); ok {
			lamaInt := int(lama)
			produkInput.LamaPengerjaan = &lamaInt
		} else if lamaStr, ok := produkData["jp_lama_pengerjaan"].(string); ok && lamaStr != "" {
			if lamaInt, err := strconv.Atoi(lamaStr); err == nil {
				produkInput.LamaPengerjaan = &lamaInt
			}
		}

		if satuanWaktu, ok := produkData["jp_satuan_waktu"].(string); ok && satuanWaktu != "" {
			produkInput.SatuanWaktu = &satuanWaktu
		}

		if keterangan, ok := produkData["jp_keterangan"].(string); ok && keterangan != "" {
			produkInput.Keterangan = &keterangan
		}

		if iconPath, ok := produkData["jp_icon_path"].(string); ok && iconPath != "" {
			produkInput.IconPath = &iconPath
		}

		products = append(products, produkInput)
	}

	return products, nil
}

func (c *LayananController) mapUploadedFiles(ctx *gin.Context) (map[string]*multipart.FileHeader, error) {
	form, err := ctx.MultipartForm()
	if err != nil || form == nil {
		return make(map[string]*multipart.FileHeader), nil
	}

	files := make(map[string]*multipart.FileHeader)

	for fieldName, fileHeaders := range form.File {
		if len(fileHeaders) > 0 && strings.HasPrefix(fieldName, "produk_") {
			files[fieldName] = fileHeaders[0]
		}
	}

	return files, nil
}

// ========== CRUD ENDPOINTS ==========

func (c *LayananController) CreateLayananWithProducts(ctx *gin.Context) {
	// Get user ID
	userID, err := c.getUserID(ctx)
	if err != nil || userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Unauthorized",
		})
		return
	}

	// Parse form data
	input, err := c.parseFormData(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Parse products
	jenisProdukJSON := ctx.PostForm("jenis_produk")
	products, err := c.parseProducts(jenisProdukJSON)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	input.JenisProduk = products

	// Map uploaded files
	files, err := c.mapUploadedFiles(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Gagal memproses file upload",
		})
		return
	}

	// Debug log
	fmt.Printf("📁 Total files received: %d\n", len(files))
	for key, file := range files {
		fmt.Printf("  - %s: %s (%d bytes)\n", key, file.Filename, file.Size)
	}

	// Validate input
	if err := c.Validator.Struct(input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Validation error",
			"error":   err.Error(),
		})
		return
	}

	// Create layanan
	layanan, err := c.Service.CreateLayananWithProducts(input, userID, files)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal menambahkan layanan",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, model.LayananWithProductsResponse{
		Success: true,
		Message: "Layanan berhasil ditambahkan",
		Data:    layanan,
	})
}

func (c *LayananController) UpdateLayananWithProducts(ctx *gin.Context) {
	layananID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid layanan ID",
		})
		return
	}
	userID, err := c.getUserID(ctx)
	if err != nil || userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Unauthorized",
		})
		return
	}

	namaLayanan := ctx.PostForm("ln_layanan")
	prioritasStr := ctx.PostForm("ln_prioritas")

	if namaLayanan == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Nama layanan wajib diisi",
		})
		return
	}

	prioritas := 0
	if prioritasStr != "" {
		prioritas, _ = strconv.Atoi(prioritasStr)
	}

	input := &model.UpdateLayananWithProductsInput{
		NamaLayanan: namaLayanan,
		Prioritas:   &prioritas,
		Cuci:        ctx.PostForm("ln_cuci"),
		Kering:      ctx.PostForm("ln_kering"),
		Setrika:     ctx.PostForm("ln_setrika"),
	}

	jenisProdukJSON := ctx.PostForm("jenis_produk")
	jenisProdukJSON = strings.TrimSpace(jenisProdukJSON)

	fmt.Printf("\n === PARSING PRODUCTS JSON ===\n")
	fmt.Printf("JSON Length: %d characters\n", len(jenisProdukJSON))

	if jenisProdukJSON == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "jenis_produk tidak boleh kosong",
		})
		return
	}

	var jenisProdukArray []map[string]interface{}
	if err := json.Unmarshal([]byte(jenisProdukJSON), &jenisProdukArray); err != nil {
		fmt.Printf(" JSON Parse Error: %v\n", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("Format JSON tidak valid: %v", err),
		})
		return
	}

	fmt.Printf(" Parsed %d products from JSON\n\n", len(jenisProdukArray))

	products := make([]model.UpdateJenisProdukInput, 0, len(jenisProdukArray))

	for i, produkData := range jenisProdukArray {
		fmt.Printf(" Product %d:\n", i)
		produkInput := model.UpdateJenisProdukInput{}
		var productID *uint
		if id, ok := produkData["jp_id"].(float64); ok && id > 0 {
			idUint := uint(id)
			productID = &idUint
			fmt.Printf("   ID (float64): %d\n", idUint)
		} else if id, ok := produkData["jp_id"].(int); ok && id > 0 {
			idUint := uint(id)
			productID = &idUint
			fmt.Printf("   ID (int): %d\n", idUint)
		} else if id, ok := produkData["jp_id"].(string); ok && id != "" {
			if idInt, err := strconv.Atoi(id); err == nil && idInt > 0 {
				idUint := uint(idInt)
				productID = &idUint
				fmt.Printf("  ID (string): %d\n", idUint)
			}
		} else if id, ok := produkData["id"].(float64); ok && id > 0 {
			idUint := uint(id)
			productID = &idUint
			fmt.Printf("   ID (fallback 'id'): %d\n", idUint)
		}

		if productID != nil {
			produkInput.ID = productID
			fmt.Printf("   MODE: UPDATE (ID=%d)\n", *productID)
		} else {
			fmt.Printf("   MODE: CREATE (No ID found)\n")
		}
		if nama, ok := produkData["jp_nama"].(string); ok && nama != "" {
			produkInput.Nama = nama
			fmt.Printf("   Name: %s\n", nama)
		} else {
			fmt.Printf("  ERROR: Nama tidak valid\n")
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": fmt.Sprintf("Produk %d: nama tidak valid", i),
			})
			return
		}
		if satuan, ok := produkData["jp_satuan"].(string); ok {
			produkInput.Satuan = &satuan
		}

		if harga, ok := produkData["jp_harga_per"].(float64); ok {
			hargaInt := int(harga)
			produkInput.HargaPer = &hargaInt
		}

		if lama, ok := produkData["jp_lama_pengerjaan"].(float64); ok {
			lamaInt := int(lama)
			produkInput.LamaPengerjaan = &lamaInt
		}

		if satuanWaktu, ok := produkData["jp_satuan_waktu"].(string); ok {
			produkInput.SatuanWaktu = &satuanWaktu
		}

		if keterangan, ok := produkData["jp_keterangan"].(string); ok {
			produkInput.Keterangan = &keterangan
		}

		if iconPath, ok := produkData["jp_icon_path"].(string); ok && iconPath != "" {
			produkInput.IconPath = &iconPath
			fmt.Printf("  Icon: %s\n", iconPath)
		}

		if imageURL, ok := produkData["jp_image_url"].(string); ok && imageURL != "" {
			fmt.Printf("    Image URL: %s\n", imageURL)
		}
		if shouldDelete, ok := produkData["should_delete"].(bool); ok {
			produkInput.ShouldDelete = shouldDelete
			if shouldDelete {
				fmt.Printf("    MARKED FOR DELETION\n")
			}
		}

		products = append(products, produkInput)
		fmt.Printf("\n")
	}

	input.JenisProduk = products

	files, err := c.mapUploadedFiles(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Gagal memproses file upload",
		})
		return
	}

	fmt.Printf(" === FILES UPLOADED ===\n")
	fmt.Printf("Total files: %d\n", len(files))
	for key, file := range files {
		fmt.Printf("  - %s: %s (%d bytes)\n", key, file.Filename, file.Size)
	}
	fmt.Printf("=========================\n\n")

	fmt.Printf(" Calling service to update layanan ID: %d\n\n", layananID)

	layanan, err := c.Service.UpdateLayananWithProducts(uint(layananID), input, userID, files)
	if err != nil {
		fmt.Printf(" Update failed: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mengupdate layanan",
			"error":   err.Error(),
		})
		return
	}

	fmt.Printf("✅ Layanan updated successfully!\n")
	fmt.Printf("   - ID: %d\n", layanan.ID)
	fmt.Printf("   - Name: %s\n", layanan.NamaLayanan)
	fmt.Printf("   - Products: %d\n\n", len(layanan.JenisProduk))

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Layanan berhasil diambil",
		"data":    layanan,
	})
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ========== QUERY ENDPOINTS ==========

func (c *LayananController) GetAllLayanan(ctx *gin.Context) {
	outletID, _ := strconv.Atoi(ctx.DefaultQuery("outlet_id", "0"))
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	layananList, total, err := c.Service.GetAllLayanan(uint(outletID), page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mengambil data layanan",
			"error":   err.Error(),
		})
		return
	}

	totalPages := (int(total) + limit - 1) / limit

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Layanan berhasil diambil",
		"data":    layananList,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func (c *LayananController) GetLayananByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid layanan ID",
		})
		return
	}

	layanan, err := c.Service.GetLayananByID(uint(id))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "tidak ditemukan") {
			statusCode = http.StatusNotFound
		}
		ctx.JSON(statusCode, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Layanan berhasil diambil",
		"data":    layanan,
	})
}

func (c *LayananController) GetLayananByOutlet(ctx *gin.Context) {
	outletID, err := strconv.Atoi(ctx.Param("outlet_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid outlet ID",
		})
		return
	}

	layananList, err := c.Service.GetLayananByOutlet(uint(outletID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mengambil data layanan",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Layanan berhasil diambil",
		"data":    layananList,
	})
}

func (c *LayananController) DeleteLayanan(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid layanan ID",
		})
		return
	}

	err = c.Service.DeleteLayanan(uint(id))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "tidak ditemukan") {
			statusCode = http.StatusNotFound
		}
		ctx.JSON(statusCode, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Layanan berhasil dihapus",
	})
}
