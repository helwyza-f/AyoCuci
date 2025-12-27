package controller

import (
    "BackendFramework/internal/model"
    "BackendFramework/internal/service"
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
)

type NotaSettingsController struct {
    service  service.NotaSettingsService
    validate *validator.Validate
}

func NewNotaSettingsController(service service.NotaSettingsService) *NotaSettingsController {
    return &NotaSettingsController{
        service:  service,
        validate: validator.New(),
    }
}

func (c *NotaSettingsController) GetByOutletID(ctx *gin.Context) {
    outletIDStr := ctx.Param("outlet_id")
    outletID, err := strconv.ParseUint(outletIDStr, 10, 32)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, model.NotaSettingsErrorResponse{
            Success: false,
            Message: "Invalid outlet ID",
        })
        return
    }
    
    settings, err := c.service.GetByOutletID(uint(outletID))
    if err != nil {
        if err.Error() == "nota settings not found" {
            ctx.JSON(http.StatusNotFound, model.NotaSettingsErrorResponse{
                Success: false,
                Message: err.Error(),
            })
            return
        }
        
        ctx.JSON(http.StatusInternalServerError, model.NotaSettingsErrorResponse{
            Success: false,
            Message: "Failed to get nota settings",
        })
        return
    }
    
    ctx.JSON(http.StatusOK, model.NotaSettingsResponse{
        Success: true,
        Message: "Nota settings retrieved successfully",
        Data:    settings,
    })
}


func (c *NotaSettingsController) CreateOrUpdate(ctx *gin.Context) {
    outletIDStr := ctx.Param("outlet_id")
    outletID, err := strconv.ParseUint(outletIDStr, 10, 32)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, model.NotaSettingsErrorResponse{
            Success: false,
            Message: "Invalid outlet ID",
        })
        return
    }
    
    var input model.NotaSettingsInput
    if err := ctx.ShouldBindJSON(&input); err != nil {
        ctx.JSON(http.StatusBadRequest, model.NotaSettingsErrorResponse{
            Success: false,
            Message: "Invalid request body",
        })
        return
    }
    
    // Validasi input
    if err := c.validate.Struct(input); err != nil {
        ctx.JSON(http.StatusBadRequest, model.NotaSettingsErrorResponse{
            Success: false,
            Message: err.Error(),
        })
        return
    }
    
    // Cek apakah settings sudah ada sebelumnya
    existingSettings, _ := c.service.GetByOutletID(uint(outletID))
    isNewRecord := existingSettings == nil
    
    settings, err := c.service.CreateOrUpdate(uint(outletID), &input)
    if err != nil {
        if err.Error() == "outlet not found" {
            ctx.JSON(http.StatusNotFound, model.NotaSettingsErrorResponse{
                Success: false,
                Message: err.Error(),
            })
            return
        }
        
        ctx.JSON(http.StatusInternalServerError, model.NotaSettingsErrorResponse{
            Success: false,
            Message: "Failed to save nota settings",
        })
        return
    }
    
    statusCode := http.StatusOK
    message := "Nota settings updated successfully"
    
    if isNewRecord {
        statusCode = http.StatusCreated
        message = "Nota settings created successfully"
    }
    
    ctx.JSON(statusCode, model.NotaSettingsResponse{
        Success: true,
        Message: message,
        Data:    settings,
    })
}


func (c *NotaSettingsController) Delete(ctx *gin.Context) {
    outletIDStr := ctx.Param("outlet_id")
    outletID, err := strconv.ParseUint(outletIDStr, 10, 32)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, model.NotaSettingsErrorResponse{
            Success: false,
            Message: "Invalid outlet ID",
        })
        return
    }
    
    err = c.service.Delete(uint(outletID))
    if err != nil {
        if err.Error() == "nota settings not found" {
            ctx.JSON(http.StatusNotFound, model.NotaSettingsErrorResponse{
                Success: false,
                Message: err.Error(),
            })
            return
        }
        
        ctx.JSON(http.StatusInternalServerError, model.NotaSettingsErrorResponse{
            Success: false,
            Message: "Failed to delete nota settings",
        })
        return
    }
    
    ctx.JSON(http.StatusOK, model.NotaSettingsErrorResponse{
        Success: true,
        Message: "Nota settings deleted successfully",
    })
}