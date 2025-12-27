package controller

import (
    "net/http"
    "strconv"
    "BackendFramework/internal/model"
	"BackendFramework/internal/service"
    "github.com/gin-gonic/gin"
)

type DiskonController struct {
    diskonService *service.DiskonService
}

func NewDiskonController(diskonService *service.DiskonService) *DiskonController {
    return &DiskonController{
        diskonService: diskonService,
    }
}

func (ctrl *DiskonController) GetAllDiskon(c *gin.Context) {

    outletID, exists := c.Get("outlet_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "error",
            "message": "Outlet ID tidak ditemukan",
        })
        return
    }
    
    diskons, err := ctrl.diskonService.GetAllDiskon(outletID.(uint))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Gagal mengambil data diskon",
            "error":   err.Error(),
        })
        return
    }
    
    // Convert ke response format
    var responses []model.DiskonResponse
    for _, diskon := range diskons {
        responses = append(responses, diskon.ToResponse())
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Data diskon berhasil diambil",
        "data":    responses,
    })
}

func (ctrl *DiskonController) GetDiskonByID(c *gin.Context) {
    idParam := c.Param("id")
    id, err := strconv.ParseUint(idParam, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "ID tidak valid",
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
    
    diskon, err := ctrl.diskonService.GetDiskonByID(uint(id), outletID.(uint))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "status":  "error",
            "message": err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Data diskon berhasil diambil",
        "data":    diskon.ToResponse(),
    })
}

func (ctrl *DiskonController) CreateDiskon(c *gin.Context) {
    var input model.DiskonInput
    
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "Input tidak valid",
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
    

    username, _ := c.Get("username")
    usernameStr := ""
    if username != nil {
        usernameStr = username.(string)
    }
    
    diskon, err := ctrl.diskonService.CreateDiskon(input, outletID.(uint), usernameStr)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Gagal membuat diskon",
            "error":   err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusCreated, gin.H{
        "status":  "success",
        "message": "Diskon berhasil dibuat",
        "data":    diskon.ToResponse(),
    })
}

func (ctrl *DiskonController) UpdateDiskon(c *gin.Context) {
    idParam := c.Param("id")
    id, err := strconv.ParseUint(idParam, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "ID tidak valid",
        })
        return
    }
    
    var input model.UpdateDiskonInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "Input tidak valid",
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
    
    username, _ := c.Get("username")
    usernameStr := ""
    if username != nil {
        usernameStr = username.(string)
    }
    
    diskon, err := ctrl.diskonService.UpdateDiskon(uint(id), input, outletID.(uint), usernameStr)
    if err != nil {
        if err.Error() == "diskon tidak ditemukan" {
            c.JSON(http.StatusNotFound, gin.H{
                "status":  "error",
                "message": err.Error(),
            })
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Gagal mengupdate diskon",
            "error":   err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Diskon berhasil diupdate",
        "data":    diskon.ToResponse(),
    })
}

func (ctrl *DiskonController) DeleteDiskon(c *gin.Context) {
    idParam := c.Param("id")
    id, err := strconv.ParseUint(idParam, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "ID tidak valid",
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
    
    if err := ctrl.diskonService.DeleteDiskon(uint(id), outletID.(uint)); err != nil {
        if err.Error() == "diskon tidak ditemukan" {
            c.JSON(http.StatusNotFound, gin.H{
                "status":  "error",
                "message": err.Error(),
            })
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Gagal menghapus diskon",
            "error":   err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Diskon berhasil dihapus",
    })
}
func (ctrl *DiskonController) GetActiveDiskon(c *gin.Context) {
    outletID, exists := c.Get("outlet_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status":  "error",
            "message": "Outlet ID tidak ditemukan",
        })
        return
    }
    
    diskons, err := ctrl.diskonService.GetActiveDiskon(outletID.(uint))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Gagal mengambil data diskon aktif",
            "error":   err.Error(),
        })
        return
    }
    
    var responses []model.DiskonResponse
    for _, diskon := range diskons {
        responses = append(responses, diskon.ToResponse())
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Data diskon aktif berhasil diambil",
        "data":    responses,
    })
}

func (ctrl *DiskonController) GetDiskonByOutlet(c *gin.Context) {
    outletIDParam := c.Param("outlet_id")
    outletID, err := strconv.ParseUint(outletIDParam, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "Outlet ID tidak valid",
        })
        return
    }
    
    diskons, err := ctrl.diskonService.GetDiskonByOutlet(uint(outletID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Gagal mengambil data diskon",
            "error":   err.Error(),
        })
        return
    }
    
    var responses []model.DiskonResponse
    for _, diskon := range diskons {
        responses = append(responses, diskon.ToResponse())
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Data diskon berhasil diambil",
        "data":    responses,
    })
}


func (ctrl *DiskonController) ToggleStatus(c *gin.Context) {
    idParam := c.Param("id")
    id, err := strconv.ParseUint(idParam, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "ID tidak valid",
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
    
    username, _ := c.Get("username")
    usernameStr := ""
    if username != nil {
        usernameStr = username.(string)
    }
    
    diskon, err := ctrl.diskonService.ToggleStatus(uint(id), outletID.(uint), usernameStr)
    if err != nil {
        if err.Error() == "diskon tidak ditemukan" {
            c.JSON(http.StatusNotFound, gin.H{
                "status":  "error",
                "message": err.Error(),
            })
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Gagal mengubah status diskon",
            "error":   err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Status diskon berhasil diubah",
        "data":    diskon.ToResponse(),
    })
}