package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"BackendFramework/internal/database"
	"BackendFramework/internal/middleware"
	"BackendFramework/internal/model"
	"BackendFramework/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var validate = validator.New()

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// RegisterUser handles standard email registration
func RegisterUser(c *gin.Context) {
	var input model.RegisterWithOutletInput

	// ShouldBindJSON akan menangkap data user DAN data outlet sekaligus
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, model.ErrorResponse{
            Success: false,
            Error:   "Invalid request body",
            Details: err.Error(),
        })
        return
    }

	// Validate sekarang juga mengecek isi field outlet (validate:"required")
    if err := validate.Struct(input); err != nil {
        c.JSON(http.StatusBadRequest, model.ErrorResponse{
            Success: false,
            Error:   "Validation failed",
            Details: err.Error(),
        })
        return
    }

	if input.Password != input.ConfirmPassword {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "Password and confirm password do not match",
		})
		return
	}

	user, err := service.RegisterWithOutlet(&input)
	if err != nil {
		handleServiceError(c, err, "Failed to register user")
		return
	}

	userID := fmt.Sprintf("%d", user.ID)
	accessToken, err := middleware.GenerateAccessToken(userID, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   "Failed to generate access token",
			Details: err.Error(),
		})
		return
	}

	refreshToken, err := middleware.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   "Failed to generate refresh token",
			Details: err.Error(),
		})
		return
	}

	// Save token data
	tokenData := bson.M{
		"user_id":               userID,
		"email":                 user.Email,
		"username":              user.NamaLengkap,
		"last_ip_address":       c.ClientIP(),
		"last_user_agent":       c.GetHeader("User-Agent"),
		"access_token":          accessToken,
		"refresh_token":         refreshToken,
		"refresh_token_expired": time.Now().Add(24 * time.Hour * 30),
		"last_login":            time.Now(),
		"is_valid_token":        "y",
		"is_remember_me":        "n",
		"login_method":          "email",
		"created_at":            time.Now(),
		"updated_at":            time.Now(),
	}

	if err := service.UpsertTokenData(userID, tokenData); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   "Failed to save token data",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":       true,
		"message":       "User registered successfully",
		"user_id":       user.ID,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// verifyGoogleAccessToken validates Google access token
func verifyGoogleAccessToken(accessToken string) (*GoogleUserInfo, error) {
	url := "https://www.googleapis.com/oauth2/v1/userinfo?access_token=" + accessToken
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to verify access token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid access token, status code: %d", resp.StatusCode)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// Validate that email is verified
	if !userInfo.VerifiedEmail {
		return nil, fmt.Errorf("email not verified by Google")
	}

	return &userInfo, nil
}

// GoogleSignIn handles Google authentication
func GoogleSignIn(c *gin.Context) {
	var input struct {
		GoogleID    string `json:"google_id" binding:"required"`
		Email       string `json:"email" binding:"required,email"`
		Name        string `json:"name"`
		PhotoUrl    string `json:"photoUrl"`
		AccessToken string `json:"accessToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.LogError(err, "[GoogleSignIn] Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Debug log
	if gin.Mode() == gin.DebugMode {
		middleware.LogError(nil, fmt.Sprintf(
			"[GoogleSignIn] Request - Email: %s, GoogleID: %s",
			input.Email,
			input.GoogleID,
		))
	}

	// Verify Google access token
	userInfo, err := verifyGoogleAccessToken(input.AccessToken)
	if err != nil {
		middleware.LogError(err, "[GoogleSignIn] Failed to verify Google access token")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid Google token",
			"error":   err.Error(),
		})
		return
	}

	// Validate email match
	if userInfo.Email != input.Email {
		middleware.LogError(nil, fmt.Sprintf(
			"[GoogleSignIn] Email mismatch - Token: %s, Input: %s",
			userInfo.Email,
			input.Email,
		))
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Email mismatch",
		})
		return
	}

	// Validate Google ID match
	if userInfo.ID != input.GoogleID {
		middleware.LogError(nil, fmt.Sprintf(
			"[GoogleSignIn] Google ID mismatch - Token: %s, Input: %s",
			userInfo.ID,
			input.GoogleID,
		))
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Google ID mismatch",
		})
		return
	}

	// Check if user exists by Google ID
	user := service.GetOneUserByGoogleID(input.GoogleID)

	if user == nil {
		// Check if email already registered with different provider
		existingUser := service.GetOneUserByEmail(input.Email)
		if existingUser != nil && existingUser.AuthProvider != "google" {
			middleware.LogError(nil, fmt.Sprintf(
				"[GoogleSignIn] Email %s already registered with %s provider",
				input.Email,
				existingUser.AuthProvider,
			))
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": fmt.Sprintf("Email sudah terdaftar dengan metode login %s", existingUser.AuthProvider),
			})
			return
		}

		// User not registered
		middleware.LogError(nil, fmt.Sprintf(
			"[GoogleSignIn] User not found - Email: %s, GoogleID: %s",
			input.Email,
			input.GoogleID,
		))
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "User not registered",
			"data": gin.H{
				"google_id":  input.GoogleID,
				"email":      input.Email,
				"name":       input.Name,
				"photoUrl":   input.PhotoUrl,
				"isNewUser":  true,
				"redirectTo": "/register/google",
			},
		})
		return
	}

	// Check user status
	if user.IsAktif != "active" {
		middleware.LogError(nil, fmt.Sprintf(
			"[GoogleSignIn] Inactive account - Email: %s",
			input.Email,
		))
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Account is inactive",
		})
		return
	}

	userID := fmt.Sprintf("%d", user.ID)

	// Get outlet ID if exists
	var outletID uint = 0
	var outlet model.Outlet
	if err := database.DbCore.Where("user_id = ?", user.ID).First(&outlet).Error; err == nil {
		outletID = outlet.ID
	}

	// Generate tokens with outletID
	accessToken, err := middleware.GenerateAccessToken(userID, outletID)
	if err != nil {
		middleware.LogError(err, "[GoogleSignIn] Failed to generate access token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to generate access token",
		})
		return
	}

	refreshToken, err := middleware.GenerateRefreshToken()
	if err != nil {
		middleware.LogError(err, "[GoogleSignIn] Failed to generate refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to generate refresh token",
		})
		return
	}

	// Track referral login
	if err := service.TrackReferralLogin(user.ID); err != nil {
		middleware.LogError(err, "[GoogleSignIn] Failed to track referral login")
	}

	// Save token to MongoDB
	tokenData := bson.M{
		"user_id":               userID,
		"email":                 user.Email,
		"username":              user.NamaLengkap,
		"last_ip_address":       c.ClientIP(),
		"last_user_agent":       c.GetHeader("User-Agent"),
		"access_token":          accessToken,
		"refresh_token":         refreshToken,
		"refresh_token_expired": time.Now().Add(24 * time.Hour * 30),
		"last_login":            time.Now(),
		"is_valid_token":        "y",
		"is_remember_me":        "n",
		"login_method":          "google",
		"updated_at":            time.Now(),
	}

	if err := service.UpsertTokenData(userID, tokenData); err != nil {
		middleware.LogError(err, "[GoogleSignIn] Failed to save token data")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to save token data",
		})
		return
	}

	// Success log
	if gin.Mode() == gin.DebugMode {
		middleware.LogError(nil, fmt.Sprintf(
			"[GoogleSignIn] Success - UserID: %s, Email: %s",
			userID,
			user.Email,
		))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login berhasil",
		"data": gin.H{
			"user": gin.H{
				"id":            user.ID,
				"username":      user.NamaLengkap,
				"email":         user.Email,
				"nomor_hp":      user.NomorHP,
				"group":         user.Group,
				"isAktif":       user.IsAktif,
				"auth_provider": user.AuthProvider,
				"referral_code": user.ReferralCode,
			},
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"isNewUser":     false,
		},
	})
}

// GoogleRegister handles Google user registration
func GoogleRegister(c *gin.Context) {
	var input model.GoogleRegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.LogError(err, "[GoogleRegister] Failed to parse request body")
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	if err := validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: err.Error(),
		})
		return
	}

	// Debug log
	if gin.Mode() == gin.DebugMode {
		middleware.LogError(nil, fmt.Sprintf(
			"[GoogleRegister] Request - Email: %s, GoogleID: %s",
			input.Email,
			input.GoogleID,
		))
	}

	// Check if user already exists
	existingByEmail := service.GetOneUserByEmail(input.Email)
	existingByGoogleID := service.GetOneUserByGoogleID(input.GoogleID)

	if existingByEmail != nil || existingByGoogleID != nil {
		middleware.LogError(nil, fmt.Sprintf(
			"[GoogleRegister] User already exists - Email: %s, GoogleID: %s",
			input.Email,
			input.GoogleID,
		))
		c.JSON(http.StatusConflict, model.ErrorResponse{
			Success: false,
			Error:   "User already registered",
		})
		return
	}

	user, err := service.GoogleRegister(&input)
	if err != nil {
		middleware.LogError(err, "[GoogleRegister] Failed to register user")
		handleServiceError(c, err, "Failed to register user with Google")
		return
	}

	userID := fmt.Sprintf("%d", user.ID)

	// Generate tokens with outletID (0 for new registration)
	accessToken, err := middleware.GenerateAccessToken(userID, 0)
	if err != nil {
		middleware.LogError(err, "[GoogleRegister] Failed to generate access token")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   "Failed to generate access token",
			Details: err.Error(),
		})
		return
	}

	refreshToken, err := middleware.GenerateRefreshToken()
	if err != nil {
		middleware.LogError(err, "[GoogleRegister] Failed to generate refresh token")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   "Failed to generate refresh token",
			Details: err.Error(),
		})
		return
	}

	// Save tokens to MongoDB
	tokenData := bson.M{
		"user_id":               userID,
		"email":                 user.Email,
		"username":              user.NamaLengkap,
		"last_ip_address":       c.ClientIP(),
		"last_user_agent":       c.GetHeader("User-Agent"),
		"access_token":          accessToken,
		"refresh_token":         refreshToken,
		"refresh_token_expired": time.Now().Add(24 * time.Hour * 30),
		"last_login":            time.Now(),
		"is_valid_token":        "y",
		"is_remember_me":        "n",
		"login_method":          "google",
		"created_at":            time.Now(),
		"updated_at":            time.Now(),
	}

	if err := service.UpsertTokenData(userID, tokenData); err != nil {
		middleware.LogError(err, "[GoogleRegister] Failed to save token data")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   "Failed to save token data",
			Details: err.Error(),
		})
		return
	}

	// Success log
	if gin.Mode() == gin.DebugMode {
		middleware.LogError(nil, fmt.Sprintf(
			"[GoogleRegister] Success - UserID: %s, Email: %s",
			userID,
			user.Email,
		))
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Registrasi berhasil",
		"data": gin.H{
			"user": gin.H{
				"id":            user.ID,
				"username":      user.NamaLengkap,
				"email":         user.Email,
				"nomor_hp":      user.NomorHP,
				"group":         user.Group,
				"auth_provider": user.AuthProvider,
				"isAktif":       user.IsAktif,
				"referral_code": user.ReferralCode,
			},
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	})
}

func GetAllUsers(c *gin.Context) {
	users := service.GetAllUsers()

	if len(users) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []model.UserList{},
			"total":   0,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
		"total":   len(users),
	})
}

func GetOneUser(c *gin.Context) {
	userId := c.Param("id")

	user := service.GetOneUser(userId)
	if user == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Success: false,
			Error:   "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}

func GetUserByEmail(c *gin.Context) {
	email := c.Query("email")

	if email == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "Email parameter is required",
		})
		return
	}

	user := service.GetOneUserByEmail(email)
	if user == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Success: false,
			Error:   "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}

func GetUserByUsername(c *gin.Context) {
	username := c.Query("username")

	if username == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "Username parameter is required",
		})
		return
	}

	user := service.GetOneUserByUsername(username)
	if user == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Success: false,
			Error:   "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}

func CreateUser(c *gin.Context) {
	var input model.RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	if err := validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: err.Error(),
		})
		return
	}

	if input.Password != input.ConfirmPassword {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "Password and confirm password do not match",
		})
		return
	}

	if err := service.InsertUser(&input); err != nil {
		handleServiceError(c, err, "Failed to create user")
		return
	}

	c.JSON(http.StatusCreated, model.RegisterResponse{
		Success: true,
		Message: "User created successfully",
	})
}

func UpdateUser(c *gin.Context) {
	userId, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	var input model.UpdateUserInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	if err := validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: err.Error(),
		})
		return
	}

	if err := service.UpdateUser(userId, &input); err != nil {
		handleServiceError(c, err, "Failed to update user")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User updated successfully",
	})
}

func UpdateUserStatus(c *gin.Context) {
	userId, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	var input struct {
		Status string `json:"status" binding:"required,oneof=active inactive"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	if err := service.UpdateUserStatus(userId, input.Status); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, model.ErrorResponse{
				Success: false,
				Error:   "User not found",
			})
			return
		}
		if err.Error() == "invalid status value" {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Success: false,
				Error:   "Invalid status value",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   "Failed to update user status",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User status updated successfully",
	})
}

func DeleteUser(c *gin.Context) {
	userId, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	if err := service.DeleteUser(userId); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, model.ErrorResponse{
				Success: false,
				Error:   "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   "Failed to delete user",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User deleted successfully",
	})
}

func HardDeleteUser(c *gin.Context) {
	userId, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	if err := service.HardDeleteUser(userId); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, model.ErrorResponse{
				Success: false,
				Error:   "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   "Failed to permanently delete user",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User permanently deleted",
	})
}

func parseUserID(c *gin.Context) (uint, error) {
	userIdStr := c.Param("id")
	userId, err := strconv.ParseUint(userIdStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(userId), nil
}

func handleServiceError(c *gin.Context, err error, defaultMsg string) {
	switch err.Error() {
	case "email already registered":
		c.JSON(http.StatusConflict, model.ErrorResponse{
			Success: false,
			Error:   "Email sudah terdaftar",
		})
	case "username already exists":
		c.JSON(http.StatusConflict, model.ErrorResponse{
			Success: false,
			Error:   "Username sudah terdaftar",
		})
	case "phone already registered", "Nomer Handphone already exists":
		c.JSON(http.StatusConflict, model.ErrorResponse{
			Success: false,
			Error:   "Nomor HP sudah terdaftar",
		})
	case "user not found":
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Success: false,
			Error:   "User not found",
		})
	case "google account already registered":
		c.JSON(http.StatusConflict, model.ErrorResponse{
			Success: false,
			Error:   "Akun Google sudah terdaftar",
		})
	case "no fields to update":
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "Tidak ada field yang diupdate",
		})
	case "referral code not found", "kode referral tidak ditemukan":
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Success: false,
			Error:   "Kode referral tidak ditemukan",
		})
	case "invalid referral code":
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "Kode referral tidak valid",
		})
	default:
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   defaultMsg,
			Details: err.Error(),
		})
	}
}