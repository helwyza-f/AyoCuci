package controller

import (
	"BackendFramework/internal/model"
	"BackendFramework/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"strconv"
)

type PaymentMethodController struct {
	service  service.PaymentMethodService
	validate *validator.Validate
}

func NewPaymentMethodController(service service.PaymentMethodService) *PaymentMethodController {
	return &PaymentMethodController{
		service:  service,
		validate: validator.New(),
	}
}

func (c *PaymentMethodController) GetAllPaymentMethods(ctx *gin.Context) {
	outletIDStr := ctx.Param("outlet_id")
	if outletIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Outlet ID is required",
		})
		return
	}

	outletID, err := strconv.ParseUint(outletIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid outlet ID format. Must be a valid number",
		})
		return
	}

	var filterActive *bool
	if activeParam := ctx.Query("active"); activeParam != "" {
		active := activeParam == "true"
		filterActive = &active
	}

	paymentMethods, err := c.service.GetAllByOutletID(uint(outletID), filterActive)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch payment methods",
			"error":   err.Error(),
		})
		return
	}

	if paymentMethods == nil {
		paymentMethods = []model.PaymentMethodList{}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Payment methods retrieved successfully",
		"data":    paymentMethods,
	})
}

func (c *PaymentMethodController) GetPaymentMethodByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid payment method ID",
		})
		return
	}

	outletIDStr := ctx.Query("outlet_id")
	if outletIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Outlet ID is required",
		})
		return
	}

	outletID, err := strconv.ParseUint(outletIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid outlet ID format",
		})
		return
	}

	paymentMethod, err := c.service.GetByID(uint(id), uint(outletID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, model.PaymentMethodResponse{
		Success: true,
		Message: "Payment method retrieved successfully",
		Data:    paymentMethod,
	})
}

func (c *PaymentMethodController) UpdatePaymentMethod(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid payment method ID",
		})
		return
	}

	outletIDStr := ctx.Query("outlet_id")
	if outletIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Outlet ID is required",
		})
		return
	}

	outletID, err := strconv.ParseUint(outletIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid outlet ID format",
		})
		return
	}

	var input model.UpdatePaymentMethodInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid input",
			"error":   err.Error(),
		})
		return
	}

	if err := c.validate.Struct(input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Validation failed",
			"error":   err.Error(),
		})
		return
	}

	paymentMethod, err := c.service.Update(uint(id), uint(outletID), input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, model.PaymentMethodResponse{
		Success: true,
		Message: "Payment method updated successfully",
		Data:    paymentMethod,
	})
}

func (c *PaymentMethodController) DeletePaymentMethod(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid payment method ID",
		})
		return
	}

	outletIDStr := ctx.Query("outlet_id")
	if outletIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Outlet ID is required",
		})
		return
	}

	outletID, err := strconv.ParseUint(outletIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid outlet ID format",
		})
		return
	}

	if err := c.service.Delete(uint(id), uint(outletID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Payment method deleted successfully",
	})
}

func (c *PaymentMethodController) ToggleActiveStatus(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid payment method ID",
		})
		return
	}

	outletIDStr := ctx.Query("outlet_id")
	if outletIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Outlet ID is required",
		})
		return
	}

	outletID, err := strconv.ParseUint(outletIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid outlet ID format",
		})
		return
	}

	var input struct {
		IsActive bool `json:"is_active"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid input",
		})
		return
	}

	if err := c.service.ToggleActive(uint(id), uint(outletID), input.IsActive); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Payment method status updated successfully",
	})
}



func (c *PaymentMethodController) CreatePaymentMethod(ctx *gin.Context) {
	outletIDStr := ctx.Param("outlet_id")
	if outletIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Outlet ID is required",
		})
		return
	}

	outletID, err := strconv.ParseUint(outletIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid outlet ID format",
		})
		return
	}

	var input model.PaymentMethodInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	paymentMethod, err := c.service.Create(uint(outletID), input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create payment method",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Payment method created successfully",
		"data":    paymentMethod,
	})
}






func (c *PaymentMethodController) parseIDs(ctx *gin.Context) (uint, uint, error) {
	outletIDStr := ctx.Param("outlet_id")
	outletID, err := strconv.ParseUint(outletIDStr, 10, 32)
	if err != nil {
		return 0, 0, err
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, 0, err
	}

	return uint(outletID), uint(id), nil
}
