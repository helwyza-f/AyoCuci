package route

import (
    "BackendFramework/internal/route/v1"
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "fmt"
    "os"
    "time"
)

func SetupRouter() *gin.Engine {
    if os.Getenv("ENVIRONMENT") == "production" {
        gin.SetMode(gin.ReleaseMode)
    }
    
    r := gin.Default()
    r.RedirectTrailingSlash = false
    r.RedirectFixedPath = false
    
    r.Use(cors.New(cors.Config{
        AllowOrigins: []string{"*"},
        AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders: []string{
            "Content-Type",
            "Authorization",
            "X-Outlet-ID",
            "X-Requested-With",
            "Accept",
            "Origin",
        },
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))
    
    uploadsPath := "C:/xampp/htdocs/Mobile-PipoSmart/uploads"
    assetsPath := "C:/xampp/htdocs/Mobile-PipoSmart/assets"
    
    if _, err := os.Stat(uploadsPath); os.IsNotExist(err) {
        fmt.Printf("⚠️  WARNING: uploads directory not found: %s\n", uploadsPath)
    } else {
        fmt.Println("✓ Uploads directory found")
    }
    
    if _, err := os.Stat(assetsPath); os.IsNotExist(err) {
        fmt.Printf("⚠️  WARNING: assets directory not found: %s\n", assetsPath)
    } else {
        fmt.Println("✓ Assets directory found")
    }
    fmt.Println("==========================================")
    
    // Serve static files
    r.Static("/uploads", uploadsPath)
    r.Static("/assets", assetsPath)
    
    // API routes
    v1Routes := r.Group("/v1")
    {
        v1.InitRoutes(v1Routes)
    }
    
    return r
}