package controller

import (
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    
    "BackendFramework/internal/model"
    "BackendFramework/internal/middleware"
    "BackendFramework/internal/thirdparty"
)

// UploadFile - Upload file ke S3
func UploadFile(c *gin.Context) {
    // Get validated input if exists
    var userInput *model.FileInput
    if validatedInput, exists := c.Get("validatedInput"); exists {
        userInput = validatedInput.(*model.FileInput)
    }
    
    // Get file from form
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "code":    http.StatusBadRequest,
            "error":   "File is required",
            "details": err.Error(),
        })
        return
    }
    
    // Configuration
    const maxFileSize = 9 * 1024 * 1024 // 9 MB
    allowedExtensions := []string{".pdf", ".jpg", ".jpeg", ".png", ".doc", ".docx", ".xls", ".xlsx"}
    uploadDir := "./temp/"
    
    // Ensure temp directory exists
    if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
        middleware.LogError(err, "Failed to create temp directory")
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "code":    http.StatusInternalServerError,
            "error":   "Failed to create temporary directory",
        })
        return
    }
    
    // Get file extension
    ext := strings.ToLower(filepath.Ext(file.Filename))
    
    // Validate file
    fileStatus, errMsg := middleware.ValidateFile(maxFileSize, file.Size, ext, allowedExtensions)
    if !fileStatus {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "code":    http.StatusBadRequest,
            "error":   errMsg,
        })
        return
    }
    
    // Generate unique filename to avoid conflicts
    uniqueFilename := generateUniqueFilename(file.Filename)
    localFilePath := filepath.Join(uploadDir, uniqueFilename)
    
    // Save file locally
    if err := c.SaveUploadedFile(file, localFilePath); err != nil {
        middleware.LogError(err, "Failed to save file locally")
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "code":    http.StatusInternalServerError,
            "error":   "Failed to save file locally",
            "details": err.Error(),
        })
        return
    }
    
    // Defer cleanup of local file
    defer func() {
        if err := os.Remove(localFilePath); err != nil {
            middleware.LogError(err, "Failed to clean up local file")
        }
    }()
    
    // Generate S3 path based on current date and unique filename
    s3Path := generateS3Path(userInput, uniqueFilename, ext)
    
    // Upload file to S3
    s3Url, err := thirdparty.UploadFileBucket(localFilePath, s3Path)
    if err != nil {
        middleware.LogError(err, "Failed to upload file to S3")
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "code":    http.StatusInternalServerError,
            "error":   "Failed to upload file to cloud storage",
            "details": err.Error(),
        })
        return
    }
    
    // Prepare response
    response := gin.H{
        "success":       true,
        "code":          http.StatusOK,
        "message":       "File uploaded successfully",
        "data": gin.H{
            "s3_url":        s3Url,
            "s3_path":       s3Path,
            "filename":      file.Filename,
            "size":          file.Size,
            "extension":     ext,
            "uploaded_at":   time.Now().Format(time.RFC3339),
        },
    }
    
    // Add user input if exists
    if userInput != nil {
        response["user_input"] = userInput
    }
    
    c.JSON(http.StatusOK, response)
}

// UploadMultipleFiles - Upload multiple files
func UploadMultipleFiles(c *gin.Context) {
    form, err := c.MultipartForm()
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "code":    http.StatusBadRequest,
            "error":   "Failed to parse multipart form",
            "details": err.Error(),
        })
        return
    }
    
    files := form.File["files"]
    if len(files) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "code":    http.StatusBadRequest,
            "error":   "No files provided",
        })
        return
    }
    
    // Configuration
    const maxFileSize = 9 * 1024 * 1024 // 9 MB
    const maxFiles = 10
    allowedExtensions := []string{".pdf", ".jpg", ".jpeg", ".png", ".doc", ".docx"}
    uploadDir := "./temp/"
    
    // Check max files limit
    if len(files) > maxFiles {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "code":    http.StatusBadRequest,
            "error":   fmt.Sprintf("Maximum %d files allowed", maxFiles),
        })
        return
    }
    
    // Ensure temp directory exists
    if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
        middleware.LogError(err, "Failed to create temp directory")
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "code":    http.StatusInternalServerError,
            "error":   "Failed to create temporary directory",
        })
        return
    }
    
    var uploadedFiles []gin.H
    var failedFiles []gin.H
    
    for _, file := range files {
        // Validate file
        ext := strings.ToLower(filepath.Ext(file.Filename))
        fileStatus, errMsg := middleware.ValidateFile(maxFileSize, file.Size, ext, allowedExtensions)
        
        if !fileStatus {
            failedFiles = append(failedFiles, gin.H{
                "filename": file.Filename,
                "error":    errMsg,
            })
            continue
        }
        
        // Generate unique filename
        uniqueFilename := generateUniqueFilename(file.Filename)
        localFilePath := filepath.Join(uploadDir, uniqueFilename)
        
        // Save file locally
        if err := c.SaveUploadedFile(file, localFilePath); err != nil {
            middleware.LogError(err, "Failed to save file: "+file.Filename)
            failedFiles = append(failedFiles, gin.H{
                "filename": file.Filename,
                "error":    "Failed to save file",
            })
            continue
        }
        
        // Generate S3 path
        s3Path := generateS3Path(nil, uniqueFilename, ext)
        
        // Upload to S3
        s3Url, err := thirdparty.UploadFileBucket(localFilePath, s3Path)
        
        // Clean up local file
        os.Remove(localFilePath)
        
        if err != nil {
            middleware.LogError(err, "Failed to upload to S3: "+file.Filename)
            failedFiles = append(failedFiles, gin.H{
                "filename": file.Filename,
                "error":    "Failed to upload to cloud storage",
            })
            continue
        }
        
        uploadedFiles = append(uploadedFiles, gin.H{
            "filename":   file.Filename,
            "s3_url":     s3Url,
            "s3_path":    s3Path,
            "size":       file.Size,
            "extension":  ext,
        })
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success":        true,
        "code":           http.StatusOK,
        "message":        fmt.Sprintf("%d files uploaded successfully", len(uploadedFiles)),
        "uploaded_files": uploadedFiles,
        "failed_files":   failedFiles,
        "total_uploaded": len(uploadedFiles),
        "total_failed":   len(failedFiles),
    })
}

// DeleteFile - Delete file from S3
func DeleteFile(c *gin.Context) {
    var input struct {
        S3Path string `json:"s3_path" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "code":    http.StatusBadRequest,
            "error":   "S3 path is required",
            "details": err.Error(),
        })
        return
    }
    
    // Delete from S3 (you need to implement this in thirdparty package)
    // err := thirdparty.DeleteFileBucket(input.S3Path)
    // if err != nil {
    //     middleware.LogError(err, "Failed to delete file from S3")
    //     c.JSON(http.StatusInternalServerError, gin.H{
    //         "success": false,
    //         "code":    http.StatusInternalServerError,
    //         "error":   "Failed to delete file from cloud storage",
    //     })
    //     return
    // }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "code":    http.StatusOK,
        "message": "File deleted successfully",
    })
}

// Helper function to generate unique filename
func generateUniqueFilename(originalFilename string) string {
    ext := filepath.Ext(originalFilename)
    nameWithoutExt := strings.TrimSuffix(originalFilename, ext)
    
    // Clean filename from special characters
    nameWithoutExt = strings.ReplaceAll(nameWithoutExt, " ", "_")
    
    // Generate UUID
    uniqueID := uuid.New().String()[:8]
    
    // Generate timestamp
    timestamp := time.Now().Format("20060102_150405")
    
    return fmt.Sprintf("%s_%s_%s%s", nameWithoutExt, timestamp, uniqueID, ext)
}

// Helper function to generate S3 path
func generateS3Path(userInput *model.FileInput, filename string, ext string) string {
    now := time.Now()
    year := now.Format("2006")
    month := now.Format("01")
    
    // Default category
    category := "general"
    
    // If userInput exists and has category field, use it
    if userInput != nil {
        // Assuming FileInput has Category field
        // category = userInput.Category
    }
    
    return fmt.Sprintf("%s/%s/%s/%s", category, year, month, filename)
}