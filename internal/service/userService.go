package service

import (
	"BackendFramework/internal/database"
	"BackendFramework/internal/middleware"
	"BackendFramework/internal/model"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func GetAllUsers() []model.UserList {
	var users []model.UserList
	result := database.DbCore.
		Raw(`SELECT id, nama_lengkap AS username, email, nomor_hp, 
			 ` + "`group`" + `, is_aktif AS isAktif, referral_code AS referralCode 
			 FROM users WHERE deleted_at IS NULL`).
		Scan(&users)

	if result.Error != nil {
		middleware.LogError(result.Error, "Query Error")
		return []model.UserList{}
	}
	return users
}

func GetOneUser(userId string) *model.UserList {
	var user model.UserList
	result := database.DbCore.
		Raw(`SELECT id, nama_lengkap AS username, email, nomor_hp, 
			 ` + "`group`" + `, is_aktif AS isAktif, referral_code AS referralCode 
			 FROM users WHERE id = ? AND deleted_at IS NULL`, userId).
		Scan(&user)

	if result.Error != nil {
		middleware.LogError(result.Error, "Data Scan Error")
		return nil
	}
	if user.ID == 0 {
		return nil
	}
	return &user
}

// --- PENCARIAN USER ---

func GetOneUserByEmail(userEmail string) *model.User {
	var user model.User
	// Menggunakan Find agar tidak muncul error "record not found" di log jika email belum terdaftar
	err := database.DbCore.Where("email = ? AND deleted_at IS NULL", userEmail).Limit(1).Find(&user).Error

	if err != nil {
		middleware.LogError(err, "Database Query Error (Email)")
		return nil
	}
	if user.ID == 0 {
		return nil
	}
	return &user
}

func GetOneUserByPhone(phone string) *model.User {
	var user model.User
	err := database.DbCore.Where("nomor_hp = ? AND deleted_at IS NULL", phone).Limit(1).Find(&user).Error

	if err != nil {
		middleware.LogError(err, "Database Query Error (Nomor HP)")
		return nil
	}
	if user.ID == 0 {
		return nil
	}
	return &user
}

func GetOneUserByUsername(username string) *model.User {
	var user model.User
	// Menghilangkan log "record not found" saat pengecekan username yang tersedia
	err := database.DbCore.Where("nama_lengkap = ? AND deleted_at IS NULL", username).Limit(1).Find(&user).Error

	if err != nil {
		middleware.LogError(err, "Database Query Error (Username)")
		return nil
	}
	if user.ID == 0 {
		return nil
	}
	return &user
}

func GetOneUserByGoogleID(googleID string) *model.User {
	var user model.User
	err := database.DbCore.Where("google_id = ? AND deleted_at IS NULL", googleID).Limit(1).Find(&user).Error

	if err != nil {
		middleware.LogError(err, "Database Query Error (Google ID)")
		return nil
	}
	if user.ID == 0 {
		return nil
	}
	return &user
}

func GetOneUserByReferralCode(referralCode string) *model.User {
	var user model.User
	err := database.DbCore.Where("referral_code = ? AND deleted_at IS NULL", referralCode).Limit(1).Find(&user).Error

	if err != nil {
		middleware.LogError(err, "Database Query Error (Referral Code)")
		return nil
	}
	if user.ID == 0 {
		return nil
	}
	return &user
}

func ValidateReferralCode(referralCode string) (*model.User, error) {
	if referralCode == "" {
		return nil, errors.New("kode referral tidak boleh kosong")
	}

	referralCode = strings.TrimSpace(strings.ToUpper(referralCode))
	referrer := GetOneUserByReferralCode(referralCode)
	if referrer == nil {
		middleware.LogError(nil, fmt.Sprintf("Kode referral tidak ditemukan: %s", referralCode))
		return nil, errors.New("kode referral tidak ditemukan")
	}

	if referrer.IsAktif != "active" {
		middleware.LogError(nil, fmt.Sprintf("Kode referral dari user tidak aktif: %s (User ID: %d)", referralCode, referrer.ID))
		return nil, errors.New("kode referral tidak dapat digunakan (akun tidak aktif)")
	}

	middleware.LogError(nil, fmt.Sprintf(" Kode referral valid: %s dari user ID: %d (%s)", referralCode, referrer.ID, referrer.NamaLengkap))
	return referrer, nil
}

func CheckGoogleEmail(email string) (*model.User, bool) {
	user := GetOneUserByEmail(email)
	if user == nil {
		return nil, false
	}
	return user, true
}

func GoogleRegister(registerData *model.GoogleRegisterInput) (*model.User, error) {
	if registerData.Email == "" || registerData.GoogleID == "" {
		return nil, errors.New("email and google ID are required")
	}

	if existingUser := GetOneUserByEmail(registerData.Email); existingUser != nil {
		return nil, errors.New("email already registered")
	}

	if existingGoogleUser := GetOneUserByGoogleID(registerData.GoogleID); existingGoogleUser != nil {
		return nil, errors.New("google account already registered")
	}

	if registerData.NamaLengkap != "" {
		if existingUsername := GetOneUserByUsername(registerData.NamaLengkap); existingUsername != nil {
			return nil, errors.New("username already exists")
		}
	}

	if registerData.NomorHP != "" {
		if existingPhone := GetOneUserByPhone(registerData.NomorHP); existingPhone != nil {
			return nil, errors.New("phone already registered")
		}
	}
	var referredBy *string
	if registerData.ReferralCode != nil && *registerData.ReferralCode != "" {
		referralCode := strings.TrimSpace(strings.ToUpper(*registerData.ReferralCode))
		referrer, err := ValidateReferralCode(referralCode)
		if err != nil {
			// Return error langsung tanpa prefix tambahan
			return nil, err
		}
		referredBy = &referralCode
		middleware.LogError(nil, fmt.Sprintf("Google Register dengan referral: %s dari user ID: %d", referralCode, referrer.ID))
	}

	user := model.User{
		NamaLengkap:         registerData.NamaLengkap,
		Email:               registerData.Email,
		NomorHP:             registerData.NomorHP,
		Group:               registerData.Group,
		GoogleID:            &registerData.GoogleID,
		AuthProvider:        "google",
		IsAktif:             "active",
		AgreeTerms:          registerData.AgreeTerms,
		SubscribeNewsletter: registerData.SubscribeNewsletter,
		Password:            "",
		ReferredBy:          referredBy,
	}

	// BeforeCreate hook will auto-generate ReferralCode
	if err := database.DbCore.Create(&user).Error; err != nil {
		middleware.LogError(err, "Google Register Failed")
		return nil, err
	}

	return &user, nil
}

// ==================== REGISTER USER ====================

func RegisterUser(userData *model.RegisterInput) (*model.User, error) {
	// Validate input
	if userData.Email == "" || userData.Password == "" {
		return nil, errors.New("email dan password wajib diisi")
	}

	// Validate password confirmation
	if userData.Password != userData.ConfirmPassword {
		return nil, errors.New("password dan konfirmasi password tidak cocok")
	}

	// Cek email sudah terdaftar
	if existingUser := GetOneUserByEmail(strings.ToLower(userData.Email)); existingUser != nil {
		return nil, errors.New("email sudah terdaftar")
	}

	// Cek username sudah digunakan
	if userData.NamaLengkap != "" {
		// Jika GetOneUserByUsername mengembalikan error 'record not found' yang tidak ditangani,
		// maka pengecekan ini bisa menganggap terjadi kesalahan fatal.
		if existingUsername := GetOneUserByUsername(userData.NamaLengkap); existingUsername != nil {
			return nil, errors.New("username sudah digunakan")
		}
	}

	// Cek nomor HP sudah terdaftar
	if userData.NomorHP != "" {
		if existingPhone := GetOneUserByPhone(userData.NomorHP); existingPhone != nil {
			return nil, errors.New("nomor HP sudah terdaftar")
		}
	}

	// Validasi dan cek kode referral dengan pengecekan ketat
	var referredBy *string
	if userData.ReferralCode != nil && *userData.ReferralCode != "" {
		referralCode := strings.TrimSpace(strings.ToUpper(*userData.ReferralCode))
		
		// Gunakan fungsi ValidateReferralCode yang sudah diperkuat
		referrer, err := ValidateReferralCode(referralCode)
		if err != nil {
			// Return error langsung tanpa prefix tambahan
			// Pesan error sudah jelas dari ValidateReferralCode
			return nil, err
		}
		
		referredBy = &referralCode
		middleware.LogError(nil, fmt.Sprintf("✅ Registrasi dengan kode referral valid: %s dari user %s (ID: %d)", 
			referralCode, referrer.NamaLengkap, referrer.ID))
	}

	// Validasi source
	var source *string
	if userData.Source != nil && *userData.Source != "" {
		validSources := []string{"instagram", "tiktok", "facebook", "referral", "google", "youtube", "others"}
		sourceValue := strings.ToLower(strings.TrimSpace(*userData.Source))
		isValid := false
		for _, validSource := range validSources {
			if sourceValue == validSource {
				isValid = true
				break
			}
		}
		if isValid {
			source = &sourceValue
		} else {
			middleware.LogError(nil, fmt.Sprintf("⚠️ Source tidak valid: %s", sourceValue))
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
	if err != nil {
		middleware.LogError(err, "Password Hashing Failed")
		return nil, errors.New("gagal mengenkripsi password")
	}

	// Create user
	user := model.User{
		NamaLengkap:         userData.NamaLengkap,
		Email:               strings.ToLower(userData.Email),
		Password:            string(hashedPassword),
		NomorHP:             userData.NomorHP,
		Group:               userData.Group,
		AuthProvider:        "email",
		IsAktif:             "active",
		AgreeTerms:          userData.AgreeTerms,
		SubscribeNewsletter: userData.SubscribeNewsletter,
		ReferredBy:          referredBy,
		Source:              source,
	}

	if err := database.DbCore.Create(&user).Error; err != nil {
		middleware.LogError(err, "Register User Failed")
		return nil, errors.New("gagal mendaftarkan user")
	}

	middleware.LogError(nil, fmt.Sprintf("✅ User registered successfully - ID: %d, Email: %s, ReferralCode: %s, ReferredBy: %v", 
		user.ID, user.Email, user.ReferralCode, referredBy))

	return &user, nil
}


// internal/service/userService.go

func RegisterWithOutlet(userData *model.RegisterWithOutletInput) (*model.User, error) {
    // Mulai Transaksi Database
    tx := database.DbCore.Begin()

    // 1. Hash Password
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)

    // 2. Buat User
    user := model.User{
        NamaLengkap: userData.NamaLengkap,
        Email:       strings.ToLower(userData.Email),
        Password:    string(hashedPassword),
        NomorHP:     userData.NomorHP,
        Group:       userData.Group,
        AgreeTerms:  userData.AgreeTerms,
        Source:      userData.Source,
        ReferredBy:  userData.ReferralCode,
        IsAktif:     "active",
    }

    if err := tx.Create(&user).Error; err != nil {
        tx.Rollback()
        return nil, err
    }

    // 3. Buat Outlet secara otomatis
    outlet := model.Outlet{
        UserID:     user.ID,
        NamaOutlet: userData.Outlet.NamaOutlet,
        Alamat:     userData.Outlet.Alamat,
        Provinsi:   userData.Outlet.Provinsi,
        Kota:       userData.Outlet.Kota,
        Kecamatan:  userData.Outlet.Kecamatan,
        NomorHP:    userData.Outlet.NomorHP,
        IsAktif:    "active",
    }

    if err := tx.Create(&outlet).Error; err != nil {
        tx.Rollback()
        return nil, err
    }

    // 4. MANIPULASI: Update field outlet_id di tabel User
    // Sekarang kita hubungkan user tersebut ke outlet_id yang baru saja dibuat
    if err := tx.Model(&user).Update("outlet_id", outlet.ID).Error; err != nil {
        tx.Rollback()
        return nil, err
    }

    // Commit jika semua sukses
    tx.Commit()
    return &user, nil
}

func InsertUser(userData *model.RegisterInput) error {
	_, err := RegisterUser(userData)
	return err
}

func UpdateUser(userId uint, userData *model.UpdateUserInput) error {
	var existingUser model.User
	if err := database.DbCore.First(&existingUser, userId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		middleware.LogError(err, "Failed to fetch user")
		return err
	}

	updates := make(map[string]interface{})
	if userData.Email != "" {
		email := strings.ToLower(userData.Email)
		if email != existingUser.Email {
			if emailCheck := GetOneUserByEmail(email); emailCheck != nil && emailCheck.ID != userId {
				return errors.New("email already registered")
			}
		}
		updates["email"] = email
	}

	if userData.NamaLengkap != "" {
		if userData.NamaLengkap != existingUser.NamaLengkap {
			if usernameCheck := GetOneUserByUsername(userData.NamaLengkap); usernameCheck != nil && usernameCheck.ID != userId {
				return errors.New("username already exists")
			}
		}
		updates["nama_lengkap"] = userData.NamaLengkap
	}
	if userData.NomorHP != "" {
		if userData.NomorHP != existingUser.NomorHP {
			if phoneCheck := GetOneUserByPhone(userData.NomorHP); phoneCheck != nil && phoneCheck.ID != userId {
				return errors.New("phone already registered")
			}
		}
		updates["nomor_hp"] = userData.NomorHP
	}

	if userData.Group != "" {
		updates["group"] = userData.Group
	}
	if userData.IsAktif != "" {
		if userData.IsAktif != "active" && userData.IsAktif != "inactive" {
			return errors.New("invalid status value")
		}
		updates["is_aktif"] = userData.IsAktif
	}

	if userData.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
		if err != nil {
			middleware.LogError(err, "Password Hashing Failed")
			return err
		}
		updates["password"] = string(hashed)
	}
	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	middleware.LogError(nil, fmt.Sprintf("Updating user %d with fields: %+v", userId, updates))

	result := database.DbCore.Model(&model.User{}).Where("id = ?", userId).Updates(updates)
	if result.Error != nil {
		middleware.LogError(result.Error, "Update User Failed")
		return result.Error
	}

	if result.RowsAffected == 0 {
		middleware.LogError(nil, fmt.Sprintf("No rows affected for user %d", userId))
		return errors.New("no rows updated")
	}

	middleware.LogError(nil, fmt.Sprintf("Successfully updated %d rows for user %d", result.RowsAffected, userId))

	return nil
}

func UpdateUserStatus(userId uint, status string) error {
	if status != "active" && status != "inactive" {
		return errors.New("invalid status value")
	}

	result := database.DbCore.Model(&model.User{}).Where("id = ?", userId).Update("is_aktif", status)
	if result.Error != nil {
		middleware.LogError(result.Error, "Update Status Failed")
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func DeleteUser(userId uint) error {
	result := database.DbCore.Delete(&model.User{}, userId)
	if result.Error != nil {
		middleware.LogError(result.Error, "Delete Data Failed")
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func HardDeleteUser(userId uint) error {
	result := database.DbCore.Unscoped().Delete(&model.User{}, userId)
	if result.Error != nil {
		middleware.LogError(result.Error, "Hard Delete Data Failed")
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func GetUserReferrals(referralCode string) ([]model.User, error) {
	var users []model.User
	err := database.DbCore.Where("referred_by = ? AND deleted_at IS NULL", referralCode).Find(&users).Error
	if err != nil {
		middleware.LogError(err, "Failed to fetch referrals")
		return nil, err
	}
	return users, nil
}

func GetReferralCount(referralCode string) (int64, error) {
	var count int64
	err := database.DbCore.Model(&model.User{}).
		Where("referred_by = ? AND deleted_at IS NULL", referralCode).
		Count(&count).Error
	if err != nil {
		middleware.LogError(err, "Failed to count referrals")
		return 0, err
	}
	return count, nil
}

func GetUsersBySource(source string) ([]model.User, error) {
	var users []model.User
	err := database.DbCore.Where("source = ? AND deleted_at IS NULL", source).Find(&users).Error
	if err != nil {
		middleware.LogError(err, "Failed to fetch users by source")
		return nil, err
	}
	return users, nil
}

func GetSourceStatistics() (map[string]int64, error) {
	var results []struct {
		Source string
		Count  int64
	}
	
	err := database.DbCore.Model(&model.User{}).
		Select("source, COUNT(*) as count").
		Where("source IS NOT NULL AND deleted_at IS NULL").
		Group("source").
		Scan(&results).Error
	
	if err != nil {
		middleware.LogError(err, "Failed to get source statistics")
		return nil, err
	}

	stats := make(map[string]int64)
	for _, result := range results {
		stats[result.Source] = result.Count
	}
	
	return stats, nil
}

// GetReferralStatistics mendapatkan statistik lengkap referral
func GetReferralStatistics(referralCode string) (map[string]interface{}, error) {
	referrals, err := GetUserReferrals(referralCode)
	if err != nil {
		return nil, err
	}

	activeCount := 0
	inactiveCount := 0
	for _, user := range referrals {
		if user.IsAktif == "active" {
			activeCount++
		} else {
			inactiveCount++
		}
	}

	stats := map[string]interface{}{
		"referral_code":  referralCode,
		"total":          len(referrals),
		"active":         activeCount,
		"inactive":       inactiveCount,
		"active_rate":    0.0,
	}

	if len(referrals) > 0 {
		stats["active_rate"] = float64(activeCount) / float64(len(referrals)) * 100
	}

	return stats, nil
}