package controller

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-ldap/ldap/v3"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"

	"BackendFramework/internal/config"
	"BackendFramework/internal/database"
	"BackendFramework/internal/middleware"
	"BackendFramework/internal/model"
	"BackendFramework/internal/service"
)

type loginBody struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	RememberMe string `json:"remember_me"`
}

type refreshBody struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	UserID       string `json:"user_id" binding:"required"`
}

type forgotPasswordInput struct {
	NomorHP string `json:"nomor_hp" binding:"required"`
}

type verifyOTPInput struct {
	NomorHP string `json:"nomor_hp" binding:"required"`
	Code    string `json:"code" binding:"required,len=6"`
}

type resetPasswordInput struct {
	NomorHP         string `json:"nomor_hp" binding:"required"`
	Code            string `json:"code" binding:"required,len=6"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

func ldapAuth(email, password string) (bool, error) {
	if config.LDAP_SERVER == "" {
		return true, nil
	}

	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", config.LDAP_SERVER, config.LDAP_PORT))
	if err != nil {
		middleware.LogError(err, "Gagal koneksi ke LDAP server")
		return false, err
	}
	defer l.Close()

	username := email
	if strings.Contains(email, "@") {
		username = strings.Split(email, "@")[0]
	}
	userDN := fmt.Sprintf("uid=%s,%s", username, config.LDAP_BASE_DN)

	err = l.Bind(userDN, password)
	if err != nil {
		middleware.LogError(err, "LDAP bind gagal")
		return false, err
	}
	return true, nil
}

// generateTokens creates access and refresh tokens
func generateTokens(userID string, outletID uint) (string, string, error) {
	accessToken, err := middleware.GenerateAccessToken(userID, outletID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := middleware.GenerateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// saveTokenData saves token information to MongoDB
func saveTokenData(userID string, user *model.User, accessToken, refreshToken, loginMethod string, c *gin.Context) error {
	tokenData := bson.M{
		"user_id":               userID,
		"email":                 user.Email,
		"username":              user.NamaLengkap,
		"last_ip_address":       c.ClientIP(),
		"last_user_agent":       c.GetHeader("User-Agent"),
		"access_token":          accessToken,
		"refresh_token":         refreshToken,
		"refresh_token_expired": time.Now().Add(config.RefreshTokenExpiry),
		"last_login":            time.Now(),
		"is_valid_token":        "y",
		"is_remember_me":        "n",
		"login_method":          loginMethod,
		"created_at":            time.Now(),
		"updated_at":            time.Now(),
	}

	return service.UpsertTokenData(userID, tokenData)
}

func Login(c *gin.Context) {
	var body loginBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  http.StatusBadRequest,
			"error": "Data request tidak valid",
		})
		return
	}

	body.Email = strings.ToLower(strings.TrimSpace(body.Email))
	clientIP := c.ClientIP()

	limiter := service.NewLoginLimiterService()

	// Check rate limiting
	canAttempt, remainingLock, _, err := limiter.CanAttemptLogin(body.Email)
	if err != nil {
		middleware.LogError(err, "Gagal cek rate limiting")
	}

	if !canAttempt {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"code":              http.StatusTooManyRequests,
			"error":             "Terlalu banyak percobaan login",
			"message":           service.FormatLockMessage(remainingLock),
			"locked_until":      time.Now().Add(remainingLock).Unix(),
			"remaining_seconds": int(remainingLock.Seconds()),
		})
		return
	}

	// Get user
	user := service.GetOneUserByEmail(body.Email)
	if user == nil {
		limiter.RecordLoginAttempt(body.Email, clientIP, false)
		remaining := limiter.GetRemainingAttempts(body.Email)

		if remaining <= 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":              http.StatusTooManyRequests,
				"error":             "Akun terkunci",
				"message":           fmt.Sprintf("Terlalu banyak percobaan gagal. Coba lagi dalam %d menit", int(service.LockoutDuration.Minutes())),
				"locked_until":      time.Now().Add(service.LockoutDuration).Unix(),
				"remaining_seconds": int(service.LockoutDuration.Seconds()),
			})
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"code":               http.StatusUnauthorized,
			"error":              "Email atau password salah",
			"remaining_attempts": remaining,
			"message":            fmt.Sprintf("Login gagal. Sisa percobaan: %d", remaining),
		})
		return
	}

	// Validate password
	passwordValid := false
	if config.LDAP_SERVER != "" {
		ldapValid, err := ldapAuth(body.Email, body.Password)
		passwordValid = ldapValid && err == nil
	} else {
		if user.Password != "" {
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
			passwordValid = err == nil
		}
	}

	if !passwordValid {
		limiter.RecordLoginAttempt(body.Email, clientIP, false)
		remaining := limiter.GetRemainingAttempts(body.Email)

		if remaining <= 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":              http.StatusTooManyRequests,
				"error":             "Akun terkunci",
				"message":           fmt.Sprintf("Terlalu banyak percobaan gagal. Coba lagi dalam %d menit", int(service.LockoutDuration.Minutes())),
				"locked_until":      time.Now().Add(service.LockoutDuration).Unix(),
				"remaining_seconds": int(service.LockoutDuration.Seconds()),
			})
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"code":               http.StatusUnauthorized,
			"error":              "Email atau password salah",
			"remaining_attempts": remaining,
			"message":            fmt.Sprintf("Login gagal. Sisa percobaan: %d", remaining),
		})
		return
	}

	// Check user status
	if user.IsAktif != "active" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":  http.StatusForbidden,
			"error": "Akun tidak aktif",
		})
		return
	}

	limiter.RecordLoginAttempt(body.Email, clientIP, true)

	// Track referral login
	if err := service.TrackReferralLogin(user.ID); err != nil {
		middleware.LogError(err, "Failed to track referral login")
	}

	// Check outlet
	var outletCount int64
	database.DbCore.Model(&model.Outlet{}).Where("user_id = ?", user.ID).Count(&outletCount)
	hasOutlet := outletCount > 0
	var outletID uint = 0
	if hasOutlet {
		var outlet model.Outlet
		if err := database.DbCore.Where("user_id = ?", user.ID).First(&outlet).Error; err == nil {
			outletID = outlet.ID
		}
	}

	userID := fmt.Sprintf("%d", user.ID)
	
	// Generate tokens with outletID
	accessToken, refreshToken, err := generateTokens(userID, outletID)
	if err != nil {
		middleware.LogError(err, "Failed to generate tokens")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  http.StatusInternalServerError,
			"error": "Gagal generate token",
		})
		return
	}

	// Save token data with remember_me
	tokenData := bson.M{
		"user_id":               userID,
		"email":                 user.Email,
		"username":              user.NamaLengkap,
		"last_ip_address":       clientIP,
		"last_user_agent":       c.GetHeader("User-Agent"),
		"access_token":          accessToken,
		"refresh_token":         refreshToken,
		"refresh_token_expired": time.Now().Add(config.RefreshTokenExpiry),
		"last_login":            time.Now(),
		"is_valid_token":        "y",
		"is_remember_me":        body.RememberMe,
		"login_method":          "email",
		"updated_at":            time.Now(),
	}

	if err := service.UpsertTokenData(userID, tokenData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  http.StatusInternalServerError,
			"error": "Gagal menyimpan data token",
		})
		return
	}

	// Prepare response
	responseData := gin.H{
		"user_id":       userID,
		"username":      user.NamaLengkap,
		"email":         user.Email,
		"nomor_hp":      user.NomorHP,
		"group":         user.Group,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"has_outlet":    hasOutlet,
		"outlet_count":  outletCount,
		"outlet_id":     outletID,
		"referral_code": user.ReferralCode,
	}

	if user.ReferredBy != nil && *user.ReferredBy != "" {
		responseData["referred_by"] = *user.ReferredBy
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Login berhasil",
		"data":    responseData,
		"token":   accessToken,
		"user": gin.H{
			"id":            userID,
			"username":      user.NamaLengkap,
			"email":         user.Email,
			"referral_code": user.ReferralCode,
		},
	})
}

func Logout(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		userID = c.Param("usrId")
	}

	if userID == "" || userID == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  http.StatusBadRequest,
			"error": "User ID tidak ditemukan",
		})
		return
	}

	userIDStr := fmt.Sprintf("%v", userID)
	if err := service.InvalidateToken(userIDStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  http.StatusInternalServerError,
			"error": "Gagal logout user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Logout berhasil",
	})
}

func LogoutAllDevices(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":  http.StatusUnauthorized,
			"error": "Unauthorized",
		})
		return
	}

	userIDStr := fmt.Sprintf("%v", userID)
	if err := service.InvalidateAllUserTokens(userIDStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  http.StatusInternalServerError,
			"error": "Gagal logout dari semua perangkat",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Berhasil logout dari semua perangkat",
	})
}

func RefreshAccessToken(c *gin.Context) {
	var body refreshBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  http.StatusBadRequest,
			"error": "Mohon sertakan refresh_token dan user_id",
		})
		return
	}

	storedToken, err := service.ValidateRefreshToken(body.UserID, body.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":  http.StatusUnauthorized,
			"error": err.Error(),
		})
		return
	}

	extendRefresh := false
	if time.Now().After(storedToken.RefreshTokenExpiredAt) {
		if storedToken.IsRememberMe != "y" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "Refresh token kadaluarsa. Silakan login kembali",
			})
			return
		}
		extendRefresh = true
	}

	if extendRefresh {
		updateData := bson.M{"refresh_token_expired": time.Now().Add(config.RefreshTokenExpiry)}
		if err := service.UpsertTokenData(storedToken.UserId, updateData); err != nil {
			middleware.LogError(err, "Gagal memperpanjang refresh token")
		}
	}

	newAccessToken, err := middleware.GenerateAccessToken(storedToken.UserId, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  http.StatusInternalServerError,
			"error": "Gagal generate access token baru",
		})
		return
	}

	if err := service.RefreshAccessToken(storedToken.UserId, newAccessToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  http.StatusInternalServerError,
			"error": "Gagal update access token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":         http.StatusOK,
		"message":      "Access token berhasil diperbarui",
		"access_token": newAccessToken,
	})
}

func GetActiveSessions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":  http.StatusUnauthorized,
			"error": "Unauthorized",
		})
		return
	}

	userIDStr := fmt.Sprintf("%v", userID)
	tokens, err := service.GetAllActiveTokens(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  http.StatusInternalServerError,
			"error": "Gagal mengambil sesi aktif",
		})
		return
	}

	var sessions []gin.H
	for _, token := range tokens {
		sessions = append(sessions, gin.H{
			"last_login":      token.LastLogin,
			"last_ip_address": token.LastIpAddress,
			"last_user_agent": token.LastUserAgent,
			"is_remember_me":  token.IsRememberMe,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"sessions": sessions,
		"total":    len(sessions),
	})
}

func GetMyReferralStats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":  http.StatusUnauthorized,
			"error": "Unauthorized",
		})
		return
	}

	user := service.GetOneUser(fmt.Sprintf("%v", userID))
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":  http.StatusNotFound,
			"error": "User tidak ditemukan",
		})
		return
	}

	stats, err := service.GetReferralStatistics(user.ReferralCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  http.StatusInternalServerError,
			"error": "Gagal mengambil statistik referral",
		})
		return
	}

	loginStats, err := service.GetReferralLoginStats(user.ReferralCode)
	if err != nil {
		middleware.LogError(err, "Failed to get referral login stats")
		loginStats = make(map[string]interface{})
	}

	engagementMetrics, err := service.GetReferralEngagementMetrics(user.ReferralCode)
	if err != nil {
		middleware.LogError(err, "Failed to get engagement metrics")
		engagementMetrics = make(map[string]interface{})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Data referral berhasil diambil",
		"data": gin.H{
			"referral_code":      user.ReferralCode,
			"statistics":         stats,
			"login_stats":        loginStats,
			"engagement_metrics": engagementMetrics,
		},
	})
}

func GetMyReferrals(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":  http.StatusUnauthorized,
			"error": "Unauthorized",
		})
		return
	}

	user := service.GetOneUser(fmt.Sprintf("%v", userID))
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":  http.StatusNotFound,
			"error": "User tidak ditemukan",
		})
		return
	}

	referrals, err := service.GetUserReferrals(user.ReferralCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  http.StatusInternalServerError,
			"error": "Gagal mengambil data referral",
		})
		return
	}

	var referralList []gin.H
	for _, ref := range referrals {
		referralList = append(referralList, gin.H{
			"id":         ref.ID,
			"username":   ref.NamaLengkap,
			"email":      ref.Email,
			"is_aktif":   ref.IsAktif,
			"created_at": ref.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Daftar referral berhasil diambil",
		"data": gin.H{
			"referral_code": user.ReferralCode,
			"total":         len(referralList),
			"referrals":     referralList,
		},
	})
}

func ForgotPassword(c *gin.Context) {
	var input forgotPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  http.StatusBadRequest,
			"error": "Nomor HP tidak boleh kosong",
		})
		return
	}

	hpLen := len(input.NomorHP)
	if hpLen < 10 || hpLen > 15 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  http.StatusBadRequest,
			"error": "Nomor HP tidak valid",
		})
		return
	}

	user := service.GetOneUserByPhone(input.NomorHP)
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":  http.StatusNotFound,
			"error": "Nomor HP tidak terdaftar",
		})
		return
	}

	otpService := service.NewOTPService(database.DbCore)
	if err := otpService.SendOTP(input.NomorHP, "forgot_password"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  http.StatusInternalServerError,
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Kode OTP telah dikirim ke WhatsApp Anda",
	})
}

func VerifyOTP(c *gin.Context) {
	var input verifyOTPInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  http.StatusBadRequest,
			"error": "Data tidak valid",
		})
		return
	}

	otpService := service.NewOTPService(database.DbCore)
	valid, err := otpService.VerifyOTP(input.NomorHP, input.Code, "forgot_password")
	if err != nil || !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  http.StatusBadRequest,
			"error": "Kode OTP tidak valid atau sudah kadaluarsa",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Kode OTP berhasil diverifikasi",
	})
}

func ResetPassword(c *gin.Context) {
	var input resetPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  http.StatusBadRequest,
			"error": "Data tidak valid",
		})
		return
	}

	if input.NewPassword != input.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  http.StatusBadRequest,
			"error": "Password dan konfirmasi password tidak cocok",
		})
		return
	}

	otpService := service.NewOTPService(database.DbCore)
	valid, err := otpService.VerifyOTP(input.NomorHP, input.Code, "forgot_password")
	if err != nil || !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  http.StatusBadRequest,
			"error": "Kode OTP tidak valid atau sudah kadaluarsa",
		})
		return
	}

	user := service.GetOneUserByPhone(input.NomorHP)
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":  http.StatusNotFound,
			"error": "User tidak ditemukan",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  http.StatusInternalServerError,
			"error": "Gagal mengenkripsi password",
		})
		return
	}

	user.Password = string(hashedPassword)
	if err := database.DbCore.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  http.StatusInternalServerError,
			"error": "Gagal mereset password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Password berhasil direset",
	})
}

func ValidateReferralCode(c *gin.Context) {
	referralCode := c.Query("code")
	if referralCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  http.StatusBadRequest,
			"error": "Kode referral tidak boleh kosong",
		})
		return
	}

	referrer, err := service.ValidateReferralCode(referralCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"valid":   false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"valid":   true,
		"message": "Kode referral valid",
		"data": gin.H{
			"referral_code": referrer.ReferralCode,
			"referrer_name": referrer.NamaLengkap,
		},
	})
}