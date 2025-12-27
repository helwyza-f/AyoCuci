package service

import (
	"fmt"
	"time"

	"BackendFramework/internal/database"
	"BackendFramework/internal/model"
)



const (
	MaxLoginAttempts    = 5               
	LockoutDuration     = 5 * time.Minute 
	AttemptWindowPeriod = 15 * time.Minute 
)



type LoginLimiterService struct{}

func NewLoginLimiterService() *LoginLimiterService {
	return &LoginLimiterService{}
}

// ========== MAIN FUNCTIONS ==========

// CanAttemptLogin mengecek apakah user boleh mencoba login
// Returns: (boleh_login, sisa_waktu_lock, jumlah_gagal, error)
func (s *LoginLimiterService) CanAttemptLogin(email string) (bool, time.Duration, int, error) {
	windowStart := time.Now().Add(-AttemptWindowPeriod)

	// Hitung jumlah percobaan gagal dalam window
	var failedCount int64
	err := database.DbCore.Model(&model.LoginAttempt{}).
		Where("email = ? AND success = ? AND attempt_time > ?", email, false, windowStart).
		Count(&failedCount).Error

	if err != nil {
		return false, 0, 0, err
	}

	// Jika belum mencapai batas, boleh login
	if failedCount < MaxLoginAttempts {
		return true, 0, int(failedCount), nil
	}

	// Sudah mencapai batas, cek apakah masih dalam periode lock
	var lastAttempt model.LoginAttempt
	err = database.DbCore.
		Where("email = ? AND success = ?", email, false).
		Order("attempt_time DESC").
		First(&lastAttempt).Error

	if err != nil {
		return false, 0, int(failedCount), err
	}

	// Hitung kapan lock berakhir
	lockExpiry := lastAttempt.AttemptTime.Add(LockoutDuration)

	if time.Now().Before(lockExpiry) {
		// Masih dalam periode lock
		remaining := lockExpiry.Sub(time.Now())
		return false, remaining, int(failedCount), nil
	}

	// Lock sudah expired, boleh login lagi
	return true, 0, int(failedCount), nil
}

// RecordLoginAttempt mencatat percobaan login ke database
func (s *LoginLimiterService) RecordLoginAttempt(email, ipAddress string, success bool) error {
	attempt := model.LoginAttempt{
		Email:       email,
		IPAddress:   ipAddress,
		AttemptTime: time.Now(),
		Success:     success,
		CreatedAt:   time.Now(),
	}
	return database.DbCore.Create(&attempt).Error
}

// GetRemainingAttempts mengembalikan sisa percobaan login
func (s *LoginLimiterService) GetRemainingAttempts(email string) int {
	windowStart := time.Now().Add(-AttemptWindowPeriod)

	var failedCount int64
	database.DbCore.Model(&model.LoginAttempt{}).
		Where("email = ? AND success = ? AND attempt_time > ?", email, false, windowStart).
		Count(&failedCount)

	remaining := MaxLoginAttempts - int(failedCount)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ClearLoginAttempts menghapus semua record percobaan login untuk email tertentu
// Biasanya dipanggil setelah login sukses (opsional)
func (s *LoginLimiterService) ClearLoginAttempts(email string) error {
	return database.DbCore.
		Where("email = ?", email).
		Delete(&model.LoginAttempt{}).Error
}

// ClearFailedAttempts menghapus hanya percobaan gagal untuk email tertentu
func (s *LoginLimiterService) ClearFailedAttempts(email string) error {
	return database.DbCore.
		Where("email = ? AND success = ?", email, false).
		Delete(&model.LoginAttempt{}).Error
}

// ========== UTILITY FUNCTIONS ==========

// FormatLockMessage membuat pesan lock yang user-friendly
func FormatLockMessage(remaining time.Duration) string {
	minutes := int(remaining.Minutes())
	seconds := int(remaining.Seconds()) % 60

	if minutes > 0 {
		return fmt.Sprintf("Akun terkunci. Coba lagi dalam %d menit %d detik", minutes, seconds)
	}
	return fmt.Sprintf("Akun terkunci. Coba lagi dalam %d detik", seconds)
}

// ========== CLEANUP FUNCTIONS (untuk Cron Job) ==========

// CleanupOldAttempts membersihkan data login attempts yang lebih dari 24 jam
// Jalankan dengan cron job setiap hari
func (s *LoginLimiterService) CleanupOldAttempts() error {
	cutoff := time.Now().Add(-24 * time.Hour)
	result := database.DbCore.
		Where("attempt_time < ?", cutoff).
		Delete(&model.LoginAttempt{})
	
	if result.Error != nil {
		return result.Error
	}
	
	fmt.Printf("[Cleanup] Deleted %d old login attempts\n", result.RowsAffected)
	return nil
}

// CleanupSuccessfulAttempts membersihkan semua login sukses yang lebih dari 1 jam
// Karena kita hanya butuh track yang gagal
func (s *LoginLimiterService) CleanupSuccessfulAttempts() error {
	cutoff := time.Now().Add(-1 * time.Hour)
	return database.DbCore.
		Where("success = ? AND attempt_time < ?", true, cutoff).
		Delete(&model.LoginAttempt{}).Error
}

// ========== ADMIN FUNCTIONS ==========

// UnlockAccount membuka lock akun secara manual (untuk admin)
func (s *LoginLimiterService) UnlockAccount(email string) error {
	return s.ClearFailedAttempts(email)
}

// GetLoginHistory mengambil history login untuk email tertentu
func (s *LoginLimiterService) GetLoginHistory(email string, limit int) ([]model.LoginAttempt, error) {
	var attempts []model.LoginAttempt
	err := database.DbCore.
		Where("email = ?", email).
		Order("attempt_time DESC").
		Limit(limit).
		Find(&attempts).Error
	
	return attempts, err
}

// GetFailedAttemptsCount mengambil jumlah percobaan gagal dalam periode tertentu
func (s *LoginLimiterService) GetFailedAttemptsCount(email string, since time.Time) (int64, error) {
	var count int64
	err := database.DbCore.Model(&model.LoginAttempt{}).
		Where("email = ? AND success = ? AND attempt_time > ?", email, false, since).
		Count(&count).Error
	
	return count, err
}