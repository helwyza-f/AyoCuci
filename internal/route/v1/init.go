package v1

import (
	"BackendFramework/internal/controller"
	"BackendFramework/internal/database"
	"BackendFramework/internal/middleware"
	"BackendFramework/internal/model"
	"BackendFramework/internal/service"

	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", controller.Login)
		auth.POST("/register", controller.RegisterUser)
		auth.POST("/google/signin", controller.GoogleSignIn)
		auth.POST("/google/register", controller.GoogleRegister)
		auth.GET("/logout/:usrId", controller.Logout)
		auth.POST("/refresh-access", middleware.LogUserActivity(), controller.RefreshAccessToken)

		// OTP
		auth.POST("/forgot-password", controller.ForgotPassword)
		auth.POST("/verify-otp", controller.VerifyOTP)
		auth.POST("/reset-password", controller.ResetPassword)
	}

	user := r.Group("/user")
	{
		user.Use(middleware.JWTAuthMiddleware(), middleware.LogUserActivity())
		user.GET("/", controller.GetAllUsers)
		user.GET("/:id", controller.GetOneUser)
		user.GET("/search/email", controller.GetUserByEmail)
		user.GET("/search/username", controller.GetUserByUsername)
		user.POST("/", controller.CreateUser)
		user.PUT("/:id", controller.UpdateUser)
		user.PATCH("/:id/status", controller.UpdateUserStatus)
		user.DELETE("/:id", controller.DeleteUser)
		user.DELETE("/:id/permanent", controller.HardDeleteUser)
	}

	outlet := r.Group("/outlet")
	{
		outlet.Use(middleware.JWTAuthMiddleware())

		outlet.POST("", controller.CreateOutletController)

		outletWithActivity := outlet.Group("")
		outletWithActivity.Use(middleware.LogUserActivity())
		{
			outletWithActivity.GET("", controller.GetAllOutletsController)
			outletWithActivity.GET("/my-outlets", controller.GetMyOutletsController)
			outletWithActivity.GET("/:id", controller.GetOutletByIDController)
			outletWithActivity.PUT("/:id", controller.UpdateOutletController)
			outletWithActivity.DELETE("/:id", controller.DeleteOutletController)
			outletWithActivity.PATCH("/:id/activate", controller.ActivateOutletController)
			outletWithActivity.PATCH("/:id/deactivate", controller.DeactivateOutletController)
		}
	}

	layananService := service.NewLayananService()
	layananController := controller.NewLayananController(layananService)

	layanan := r.Group("/layanan")
	{
		layanan.Use(middleware.JWTAuthMiddleware(), middleware.LogUserActivity())
		layanan.POST("/with-products", layananController.CreateLayananWithProducts)
		layanan.PUT("/with-products/:id", layananController.UpdateLayananWithProducts)

		layanan.POST("/simple", layananController.CreateLayananWithProducts)
		layanan.PUT("/simple/:id", layananController.UpdateLayananWithProducts)

		layanan.GET("", layananController.GetAllLayanan)
		layanan.GET("/:id", layananController.GetLayananByID)
		layanan.GET("/outlet/:outlet_id", layananController.GetLayananByOutlet)
		layanan.DELETE("/:id", layananController.DeleteLayanan)
	}

	kategoriPengeluaranService := service.NewKategoriPengeluaranService()
	kategoriPengeluaranController := controller.NewKategoriPengeluaranController(kategoriPengeluaranService)

	kategoriPengeluaran := r.Group("/kategori-pengeluaran")
	{
		kategoriPengeluaran.Use(middleware.JWTAuthMiddleware(), middleware.LogUserActivity())
		kategoriPengeluaran.GET("", kategoriPengeluaranController.GetAll)
		kategoriPengeluaran.GET("/:id", kategoriPengeluaranController.GetByID)
		kategoriPengeluaran.POST("", kategoriPengeluaranController.Create)
		kategoriPengeluaran.PUT("/:id", kategoriPengeluaranController.Update)
		kategoriPengeluaran.DELETE("/:id", kategoriPengeluaranController.Delete)
	}

	diskonService := service.NewDiskonService()
	diskonController := controller.NewDiskonController(diskonService)

	diskon := r.Group("/diskon")
	{
		diskon.Use(middleware.JWTAuthMiddleware(), middleware.LogUserActivity())
		diskon.GET("", diskonController.GetAllDiskon)
		diskon.GET("/active", diskonController.GetActiveDiskon)
		diskon.GET("/outlet/:outlet_id", diskonController.GetDiskonByOutlet)
		diskon.GET("/:id", diskonController.GetDiskonByID)
		diskon.POST("", diskonController.CreateDiskon)
		diskon.PUT("/:id", diskonController.UpdateDiskon)
		diskon.PATCH("/:id/toggle", diskonController.ToggleStatus)
		diskon.DELETE("/:id", diskonController.DeleteDiskon)
	}

	parfumService := service.NewParfumService()
	parfumController := controller.NewParfumController(parfumService)

	parfum := r.Group("/parfum")
	{
		parfum.Use(middleware.JWTAuthMiddleware(), middleware.LogUserActivity())
		parfum.GET("", parfumController.GetAllParfum)
		parfum.GET("/:id", parfumController.GetParfumByID)
		parfum.POST("", parfumController.CreateParfum)
		parfum.PUT("/:id", parfumController.UpdateParfum)
		parfum.DELETE("/:id", parfumController.DeleteParfum)
	}

	notaSettingService := service.NewNotaSettingsService()
	notaSettingController := controller.NewNotaSettingsController(notaSettingService)

	notaSettings := r.Group("/nota-settings")
	{
		notaSettings.Use(middleware.JWTAuthMiddleware(), middleware.LogUserActivity())
		notaSettings.GET("/outlet/:outlet_id", notaSettingController.GetByOutletID)
		notaSettings.POST("/outlet/:outlet_id", notaSettingController.CreateOrUpdate)
		notaSettings.DELETE("/outlet/:outlet_id", notaSettingController.Delete)
	}

	notaService := service.NewNotaService(database.DbCore)
	notaController := controller.NewNotaController(notaService)

	nota := r.Group("/nota")
	{
		nota.Use(middleware.JWTAuthMiddleware(), middleware.LogUserActivity())

		nota.POST("", notaController.GenerateNota)
		nota.GET("/:id", notaController.GetNotaByID)
		nota.GET("/transaction/:transaction_id", notaController.GetNotaByTransactionID)
		nota.GET("/outlet/:outlet_id", notaController.GetNotasByOutlet)

		nota.POST("/print", notaController.PrintNota)
		nota.POST("/preview", notaController.PreviewNota)
		nota.PATCH("/:id/void", notaController.VoidNota)
	}

	karyawanService := service.NewKaryawanService(database.DbCore)
	karyawanController := controller.NewKaryawanController(karyawanService)

	karyawan := r.Group("/karyawan")
	{
		karyawan.Use(middleware.JWTAuthMiddleware(), middleware.LogUserActivity())

		karyawan.GET("", karyawanController.GetAllKaryawan)
		karyawan.GET("/:id", karyawanController.GetKaryawanByID)
		karyawan.POST("", karyawanController.CreateKaryawan)
		karyawan.PUT("/:id", karyawanController.UpdateKaryawan)
		karyawan.DELETE("/:id", karyawanController.DeleteKaryawan)
	}
	paymentMethodService := service.NewPaymentMethodService()
	paymentMethodController := controller.NewPaymentMethodController(paymentMethodService)

	paymentMethods := r.Group("/payment-methods")
	{
		paymentMethods.Use(middleware.JWTAuthMiddleware(), middleware.LogUserActivity())
		paymentMethods.GET("/outlet/:outlet_id", paymentMethodController.GetAllPaymentMethods)
		paymentMethods.POST("/outlet/:outlet_id", paymentMethodController.CreatePaymentMethod)
		paymentMethods.GET("/:id", paymentMethodController.GetPaymentMethodByID)
		paymentMethods.PUT("/:id", paymentMethodController.UpdatePaymentMethod)
		paymentMethods.DELETE("/:id", paymentMethodController.DeletePaymentMethod)
		paymentMethods.PATCH("/:id/toggle-active", paymentMethodController.ToggleActiveStatus)
	}

	customerService := service.NewCustomerService(database.DbCore)
	customerController := controller.NewCustomerController(customerService)

	customer := r.Group("/customers")
	{
		customer.Use(middleware.JWTAuthMiddleware(), middleware.LogUserActivity())

		customer.GET("", customerController.GetAllCustomers)
		customer.GET("/:id", customerController.GetCustomerByID)
		customer.POST("", customerController.CreateCustomer)
		customer.PUT("/:id", customerController.UpdateCustomer)
		customer.DELETE("/:id", customerController.DeleteCustomer)
	}

	misc := r.Group("/misc")
	{
		misc.Use(middleware.JWTAuthMiddleware(), middleware.LogUserActivity())

		fileInput := &model.FileInput{}
		misc.POST("/upload-data-s3-local", middleware.InputValidator(fileInput), controller.UploadFile)
		misc.GET("/generate-pdf", controller.TryGeneratePdf)
		misc.GET("/send-mail", controller.SendMail)
		misc.GET("/generate-excel", controller.GenerateExcel)
		misc.POST("/read-excel", controller.ReadExcel)
		misc.GET("/test-ping", controller.PingMongo)
	}
}
