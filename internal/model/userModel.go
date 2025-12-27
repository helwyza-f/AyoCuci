package model

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                  uint       `json:"id" gorm:"primaryKey"`
	NamaLengkap         string     `json:"username" gorm:"uniqueIndex;not null;size:100"`
	Email               string     `json:"email" gorm:"uniqueIndex;not null;size:255"`
	Password            string     `json:"-" gorm:"not null"`
	NomorHP             string     `json:"nomor_hp" gorm:"size:20"`
	Group               string     `json:"group" gorm:"type:varchar(20);not null;default:'karyawan'"`
	OutletID            *uint      `json:"outlet_id" gorm:"column:outlet_id;index"`
	
	IsAktif             string     `json:"isAktif" gorm:"size:20;default:'active'"`
	AgreeTerms          bool       `json:"agreeTerms" gorm:"default:false"`
	SubscribeNewsletter bool       `json:"subscribeNewsletter" gorm:"default:false"`
	ReferralCode        string     `json:"referralCode" gorm:"uniqueIndex;size:20;not null"`
	ReferredBy          *string    `json:"referredBy" gorm:"size:20;index"`
	Source              *string    `json:"source" gorm:"size:50;index"`
	GoogleID            *string    `json:"-" gorm:"uniqueIndex;size:255"`
	AuthProvider        string     `json:"auth_provider" gorm:"size:20;default:'email'"`
	LastLoginAt         *time.Time `json:"last_login_at"`
	LoginCount          int        `json:"login_count" gorm:"default:0"`

	CreatedAt time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
func (User) TableName() string {
	return "users"
}

func GenerateReferralCode() (string, error) {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	code := base64.URLEncoding.EncodeToString(bytes)
	code = strings.ToUpper(strings.ReplaceAll(code, "-", ""))
	code = strings.ReplaceAll(code, "_", "")
	code = strings.ReplaceAll(code, "=", "")

	if len(code) > 8 {
		code = code[:8]
	}

	return "REF" + code, nil
}
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ReferralCode == "" {
		for i := 0; i < 5; i++ {
			code, err := GenerateReferralCode()
			if err != nil {
				return err
			}

			var count int64
			if err := tx.Model(&User{}).Where("referral_code = ?", code).Count(&count).Error; err != nil {
				return err
			}

			if count == 0 {
				u.ReferralCode = code
				break
			}
		}
	}

	if u.ReferredBy != nil && *u.ReferredBy != "" {
		normalizedCode := strings.TrimSpace(strings.ToUpper(*u.ReferredBy))
		u.ReferredBy = &normalizedCode
	}

	return nil
}

func (u *User) AfterCreate(tx *gorm.DB) error {
	if u.ReferredBy != nil && *u.ReferredBy != "" {

	}
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("ReferralCode") {
		return gorm.ErrInvalidData
	}
	return nil
}
func (u *User) HasReferrer() bool {
	return u.ReferredBy != nil && *u.ReferredBy != ""
}


func (u *User) IsActive() bool {
	return u.IsAktif == "active"
}


func (u *User) CanUseReferral() bool {
	return u.IsActive() && u.ReferralCode != ""
}


type RegisterInput struct {
	NamaLengkap         string  `json:"username" validate:"required,min=3,max=100"`
	Email               string  `json:"email" validate:"required,email"`
	Password            string  `json:"password" validate:"required,min=8"`
	ConfirmPassword     string  `json:"confirmPassword" validate:"required"`
	NomorHP             string  `json:"nomor_hp" validate:"required,min=10,max=15"`
	Group               string  `json:"group" validate:"required,oneof=owner karyawan"`
	AgreeTerms          bool    `json:"agreeTerms" validate:"required,eq=true"`
	SubscribeNewsletter bool    `json:"subscribeNewsletter"`
	ReferralCode        *string `json:"referralCode" validate:"omitempty,min=8,max=20"`
	Source              *string `json:"source" validate:"omitempty,oneof=instagram tiktok facebook referral google youtube others"`
}

type GoogleRegisterInput struct {
	GoogleID            string  `json:"google_id" validate:"required"`
	Email               string  `json:"email" validate:"required,email"`
	NamaLengkap         string  `json:"username" validate:"required,min=3,max=100"`
	NomorHP             string  `json:"nomor_hp" validate:"omitempty,min=10,max=15"`
	Group               string  `json:"group" validate:"required,oneof=owner karyawan"`
	AgreeTerms          bool    `json:"agreeTerms" validate:"required,eq=true"`
	SubscribeNewsletter bool    `json:"subscribeNewsletter"`
	ReferralCode        *string `json:"referralCode" validate:"omitempty,min=8,max=20"`
	Source              *string `json:"source" validate:"omitempty,oneof=instagram tiktok facebook referral google youtube others"`
}

type UpdateUserInput struct {
	NamaLengkap string `json:"username" validate:"omitempty,min=3,max=100"`
	Email       string `json:"email" validate:"omitempty,email"`
	NomorHP     string `json:"nomor_hp" validate:"omitempty,min=10,max=15"`
	Group       string `json:"group" validate:"omitempty,oneof=owner karyawan"`
	Password    string `json:"password" validate:"omitempty,min=8"`
	IsAktif     string `json:"isAktif" validate:"omitempty,oneof=active inactive"`
}


type UserList struct {
	ID           uint       `json:"id"`
	NamaLengkap  string     `json:"username"`
	Email        string     `json:"email"`
	NomorHP      string     `json:"nomor_hp"`
	Group        string     `json:"group"`
	IsAktif      string     `json:"isAktif"`
	ReferralCode string     `json:"referralCode"`
	ReferredBy   *string    `json:"referredBy,omitempty"`
	Source       *string    `json:"source,omitempty"`
	LoginCount   int        `json:"login_count"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
}

type UserResponse struct {
	ID                  uint       `json:"id"`
	NamaLengkap         string     `json:"username"`
	Email               string     `json:"email"`
	NomorHP             string     `json:"nomor_hp"`
	Group               string     `json:"group"`
	IsAktif             string     `json:"isAktif"`
	ReferralCode        string     `json:"referralCode"`
	ReferredBy          *string    `json:"referredBy,omitempty"`
	Source              *string    `json:"source,omitempty"`
	AgreeTerms          bool       `json:"agreeTerms"`
	SubscribeNewsletter bool       `json:"subscribeNewsletter"`
	AuthProvider        string     `json:"auth_provider"`
	LoginCount          int        `json:"login_count"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

type RegisterResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	User    *UserResponse `json:"user,omitempty"`
	Token   string        `json:"token,omitempty"`
}

type LoginResponse struct {
	Success      bool          `json:"success"`
	Message      string        `json:"message"`
	Token        string        `json:"token"`
	RefreshToken string        `json:"refresh_token,omitempty"`
	User         *UserResponse `json:"user"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

type ReferralStatsResponse struct {
	ReferralCode      string  `json:"referral_code"`
	TotalReferrals    int64   `json:"total_referrals"`
	ActiveReferrals   int64   `json:"active_referrals"`
	InactiveReferrals int64   `json:"inactive_referrals"`
	ActiveRate        float64 `json:"active_rate"`
	TotalLogins       int64   `json:"total_logins,omitempty"`
	EngagementRate    string  `json:"engagement_rate,omitempty"`
}

type ReferralListResponse struct {
	ID          uint       `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	IsAktif     string     `json:"is_aktif"`
	LoginCount  int        `json:"login_count"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (u *User) ToUserResponse() *UserResponse {
	return &UserResponse{
		ID:                  u.ID,
		NamaLengkap:         u.NamaLengkap,
		Email:               u.Email,
		NomorHP:             u.NomorHP,
		Group:               u.Group,
		IsAktif:             u.IsAktif,
		ReferralCode:        u.ReferralCode,
		ReferredBy:          u.ReferredBy,
		Source:              u.Source,
		AgreeTerms:          u.AgreeTerms,
		SubscribeNewsletter: u.SubscribeNewsletter,
		AuthProvider:        u.AuthProvider,
		LoginCount:          u.LoginCount,
		LastLoginAt:         u.LastLoginAt,
		CreatedAt:           u.CreatedAt,
		UpdatedAt:           u.UpdatedAt,
	}
}

func (u *User) ToReferralListResponse() *ReferralListResponse {
	return &ReferralListResponse{
		ID:          u.ID,
		Username:    u.NamaLengkap,
		Email:       u.Email,
		IsAktif:     u.IsAktif,
		LoginCount:  u.LoginCount,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
	}
}

func ValidateReferralInput(code string) bool {
	if code == "" {
		return false
	}
	
	if !strings.HasPrefix(code, "REF") {
		return false
	}
	
	if len(code) < 8 || len(code) > 20 {
		return false
	}
	
	for _, char := range code {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return false
		}
	}
	
	return true
}

func NormalizeReferralCode(code string) string {
	return strings.TrimSpace(strings.ToUpper(code))
}

type FileInput struct {
	FileName string `json:"fileName" form:"fileName"`
	FileType string `json:"fileType" form:"fileType"`
	FileSize int64  `json:"fileSize" form:"fileSize"`
}