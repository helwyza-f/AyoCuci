
package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"BackendFramework/internal/model"
	"gorm.io/gorm"
)

type OTPService struct {
	db             *gorm.DB
	infobipService *InfobipService
}

func NewOTPService(db *gorm.DB) *OTPService {
	return &OTPService{
		db:             db,
		infobipService: NewInfobipService(),
	}
}

// membuat kode OTP 6 digit random
func (s *OTPService) GenerateOTPCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}


func (s *OTPService) SendOTP(nomorHP, purpose string) error {
	
	var user model.User
	if err := s.db.Where("nomor_hp = ?", nomorHP).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("nomor HP tidak terdaftar")
		}
		return err
	}

	s.db.Where("nomor_hp = ? AND purpose = ? AND is_used = false", nomorHP, purpose).Delete(&model.OTP{})

	
	code, err := s.GenerateOTPCode()
	if err != nil {
		return errors.New("gagal generate OTP code")
	}

	// OTP ke database
	otp := model.OTP{
		NomorHP:   nomorHP,
		Code:      code,
		ExpiresAt: time.Now().Add(5 * time.Minute), 
		IsUsed:    false,
		Purpose:   purpose,
	}

	if err := s.db.Create(&otp).Error; err != nil {
		return errors.New("gagal menyimpan OTP")
	}

	// Kirim OTP via WhatsApp menggunakan Infobip
	if err := s.infobipService.SendWhatsAppOTP(nomorHP, code); err != nil {
		return fmt.Errorf("gagal mengirim OTP: %v", err)
	}

	return nil
}

//  memverifikasi kode OTP
func (s *OTPService) VerifyOTP(nomorHP, code, purpose string) (bool, error) {
	var otp model.OTP
	err := s.db.Where(
		"nomor_hp = ? AND code = ? AND purpose = ? AND is_used = false AND expires_at > ?",
		nomorHP, code, purpose, time.Now(),
	).First(&otp).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("kode OTP tidak valid atau sudah kadaluarsa")
		}
		return false, err
	}

	otp.IsUsed = true
	if err := s.db.Save(&otp).Error; err != nil {
		return false, err
	}

	return true, nil
}


func (s *OTPService) CleanupExpiredOTP() error {
	return s.db.Where("expires_at < ? OR is_used = true", time.Now().Add(-24*time.Hour)).Delete(&model.OTP{}).Error
}