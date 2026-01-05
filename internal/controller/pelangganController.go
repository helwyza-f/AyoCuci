package controller

// import (
//     "net/http"
//     "strconv"
// 	"BackendFramework/internal/model"
// 	"BackendFramework/internal/service"
//     "github.com/gin-gonic/gin"
// )

// type CustomerController struct {
//     CustomerService *service.CustomerService
// }

// func NewCustomerController(customerService *service.CustomerService) *CustomerController {
//     return &CustomerController{CustomerService: customerService}
// }

// func (ctrl *CustomerController) GetAllCustomers(c *gin.Context) {
//     outletIDStr := c.Query("outlet_id")
//     if outletIDStr == "" {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "success": false,
//             "error":   "outlet_id is required",
//         })
//         return
//     }

//     outletID, err := strconv.ParseUint(outletIDStr, 10, 32)
//     if err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "success": false,
//             "error":   "invalid outlet_id",
//         })
//         return
//     }
//     searchQuery := c.Query("search")
//     var customers []model.CustomerResponse

//     if searchQuery != "" {
//         customers, err = ctrl.CustomerService.SearchCustomers(uint(outletID), searchQuery)
//     } else {
//         customers, err = ctrl.CustomerService.GetAllCustomers(uint(outletID))
//     }

//     if err != nil {
//         c.JSON(http.StatusInternalServerError, gin.H{
//             "success": false,
//             "error":   err.Error(),
//         })
//         return
//     }

//     c.JSON(http.StatusOK, gin.H{
//         "success": true,
//         "data":    customers,
//     })
// }

// func (ctrl *CustomerController) GetCustomerByID(c *gin.Context) {
//     idStr := c.Param("id")
//     outletIDStr := c.Query("outlet_id")

//     if outletIDStr == "" {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "success": false,
//             "error":   "outlet_id is required",
//         })
//         return
//     }

//     id, err := strconv.ParseUint(idStr, 10, 32)
//     if err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "success": false,
//             "error":   "invalid customer id",
//         })
//         return
//     }

//     outletID, err := strconv.ParseUint(outletIDStr, 10, 32)
//     if err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "success": false,
//             "error":   "invalid outlet_id",
//         })
//         return
//     }

//     customer, err := ctrl.CustomerService.GetCustomerByID(uint(id), uint(outletID))
//     if err != nil {
//         c.JSON(http.StatusNotFound, gin.H{
//             "success": false,
//             "error":   err.Error(),
//         })
//         return
//     }

//     c.JSON(http.StatusOK, gin.H{
//         "success": true,
//         "data":    customer,
//     })
// }

// func (ctrl *CustomerController) CreateCustomer(c *gin.Context) {
//     outletIDStr := c.Query("outlet_id")
//     if outletIDStr == "" {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "success": false,
//             "error":   "outlet_id is required",
//         })
//         return
//     }

//     outletID, err := strconv.ParseUint(outletIDStr, 10, 32)
//     if err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "success": false,
//             "error":   "invalid outlet_id",
//         })
//         return
//     }

//     var input model.CustomerInput
//     if err := c.ShouldBindJSON(&input); err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "success": false,
//             "error":   err.Error(),
//         })
//         return
//     }
//     userEmail, _ := c.Get("user_email")
//     userUpdate := ""
//     if email, ok := userEmail.(string); ok {
//         userUpdate = email
//     }

//     customer, err := ctrl.CustomerService.CreateCustomer(input, uint(outletID), userUpdate)
//     if err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "success": false,
//             "error":   err.Error(),
//         })
//         return
//     }

//     c.JSON(http.StatusCreated, gin.H{
//         "success": true,
//         "message": "Customer berhasil ditambahkan",
//         "data":    customer,
//     })
// }

// func (ctrl *CustomerController) UpdateCustomer(c *gin.Context) {
//     idStr := c.Param("id")
//     outletIDStr := c.Query("outlet_id")

//     if outletIDStr == "" {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "success": false,
//             "error":   "outlet_id is required",
//         })
//         return
//     }

//     id, err := strconv.ParseUint(idStr, 10, 32)
//     if err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "success": false,
//             "error":   "invalid customer id",
//         })
//         return
//     }

//     outletID, err := strconv.ParseUint(outletIDStr, 10, 32)
//     if err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "success": false,
//             "error":   "invalid outlet_id",
//         })
//         return
//     }

//     var input model.UpdateCustomerInput
//     if err := c.ShouldBindJSON(&input); err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "success": false,
//             "error":   err.Error(),
//         })
//         return
//     }
//     userEmail, _ := c.Get("user_email")
//     userUpdate := ""
//     if email, ok := userEmail.(string); ok {
//         userUpdate = email
//     }

//     customer, err := ctrl.CustomerService.UpdateCustomer(uint(id), uint(outletID), input, userUpdate)
//     if err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "success": false,
//             "error":   err.Error(),
//         })
//         return
//     }

//     c.JSON(http.StatusOK, gin.H{
//         "success": true,
//         "message": "Customer berhasil diupdate",
//         "data":    customer,
//     })
// }

// func (ctrl *CustomerController) DeleteCustomer(c *gin.Context) {
//     idStr := c.Param("id")
//     outletIDStr := c.Query("outlet_id")

//     if outletIDStr == "" {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "success": false,
//             "error":   "outlet_id is required",
//         })
//         return
//     }

//     id, err := strconv.ParseUint(idStr, 10, 32)
//     if err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "success": false,
//             "error":   "invalid customer id",
//         })
//         return
//     }

//     outletID, err := strconv.ParseUint(outletIDStr, 10, 32)
//     if err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "success": false,
//             "error":   "invalid outlet_id",
//         })
//         return
//     }

//     if err := ctrl.CustomerService.DeleteCustomer(uint(id), uint(outletID)); err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{
//             "success": false,
//             "error":   err.Error(),
//         })
//         return
//     }

//     c.JSON(http.StatusOK, gin.H{
//         "success": true,
//         "message": "Customer berhasil dihapus",
//     })
// }