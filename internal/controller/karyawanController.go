package controller

import (
    "net/http"
    "strconv"
    "BackendFramework/internal/model"
	"BackendFramework/internal/service"
    "github.com/gin-gonic/gin"
)

type KaryawanController struct {
	Service *service.KaryawanService
}

func NewKaryawanController(service *service.KaryawanService) *KaryawanController {
	return &KaryawanController{Service: service}
}

func (c *KaryawanController) GetAllKaryawan(ctx *gin.Context) {
    var outletID *uint
    if outletParam := ctx.Query("outlet_id"); outletParam != "" {
        id, err := strconv.ParseUint(outletParam, 10, 32)
        if err == nil {
            oid := uint(id)
            outletID = &oid
        }
    }

    karyawans, err := c.Service.GetAll(outletID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "error": err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "message": "Berhasil mendapatkan data karyawan",
        "data":    karyawans,
    })
}

func (c *KaryawanController) GetKaryawanByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H {
			"error": "ID tidak valid",
		})
		return
	}

	var outletID *uint
	if outletParam := ctx.Query("outlet_id"); outletParam != ""{
		oid, err := strconv.ParseUint(outletParam, 10, 32)
        if err == nil {
            o := uint(oid)
            outletID = &o
        }
	}

	  karyawan, err := c.Service.GetByID(uint(id), outletID)
    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{
            "error": err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "message": "Berhasil mendapatkan data karyawan",
        "data":    karyawan,
    })
}

func (c *KaryawanController) CreateKaryawan(ctx *gin.Context) {
    var input model.KaryawanInput
    if err := ctx.ShouldBindJSON(&input); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "error": err.Error(),
        })
        return
    }

    var outletID *uint
    if outletParam := ctx.Query("outlet_id"); outletParam != "" {
        id, err := strconv.ParseUint(outletParam, 10, 32)
        if err == nil {
            oid := uint(id)
            outletID = &oid
        }
    }

    userUpdate := "admin" 
    if user, exists := ctx.Get("user"); exists {
        userUpdate = user.(string)
    }

    karyawan, err := c.Service.Create(&input, outletID, userUpdate)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "error": err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusCreated, gin.H{
        "message": "Karyawan berhasil ditambahkan",
        "data":    karyawan,
    })
}

func (c *KaryawanController) UpdateKaryawan(ctx *gin.Context) {
    id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "error": "ID tidak valid",
        })
        return
    }

    var input model.UpdateKaryawanInput
    if err := ctx.ShouldBindJSON(&input); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "error": err.Error(),
        })
        return
    }

    var outletID *uint
    if outletParam := ctx.Query("outlet_id"); outletParam != "" {
        oid, err := strconv.ParseUint(outletParam, 10, 32)
        if err == nil {
            o := uint(oid)
            outletID = &o
        }
    }

    userUpdate := "admin"
    if user, exists := ctx.Get("user"); exists {
        userUpdate = user.(string)
    }

    karyawan, err := c.Service.Update(uint(id), &input, outletID, userUpdate)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "error": err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "message": "Karyawan berhasil diupdate",
        "data":    karyawan,
    })
}


func (c *KaryawanController) DeleteKaryawan(ctx *gin.Context) {
    id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "error": "ID tidak valid",
        })
        return
    }

    var outletID *uint
    if outletParam := ctx.Query("outlet_id"); outletParam != "" {
        oid, err := strconv.ParseUint(outletParam, 10, 32)
        if err == nil {
            o := uint(oid)
            outletID = &o
        }
    }

    if err := c.Service.Delete(uint(id), outletID); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "error": err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "message": "Karyawan berhasil dihapus",
    })
}