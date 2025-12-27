package service

import (
	"BackendFramework/internal/database"
	"BackendFramework/internal/model"
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	"gorm.io/gorm"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func normalizePhotoURL(photo string) string {
	if photo == "" {
		return ""
	}
	
	if strings.Contains(photo, "http://localhost/Mobile-PipoSmart") {
		return strings.Replace(photo, "http://localhost/Mobile-PipoSmart", "http://localhost:8080", 1)
	}
	
	if strings.HasPrefix(photo, "http://") || strings.HasPrefix(photo, "https://") {
		return photo
	}
	
	if !strings.HasPrefix(photo, "/") {
		photo = "/uploads/outlets/" + filepath.Base(photo)
	}
	
	return "http://localhost:8080" + photo
}

func CreateOutlet(userID uint, input model.OutletInput, file *multipart.FileHeader) (*model.Outlet, error) {
	var user model.User
	if err := database.DbCore.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal validasi user: %v", err)
	}
	
	var count int64
	if err := database.DbCore.Model(&model.Outlet{}).
		Where("user_id = ? AND is_aktif = ?", userID, "active").
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("gagal memeriksa outlet: %v", err)
	}

	if count > 0 {
		return nil, fmt.Errorf("user sudah memiliki outlet aktif")
	}

	var photoPath string
	var fullPath string
	if file != nil {
		uploadDir := `C:\xampp\htdocs\Mobile-PipoSmart\uploads\outlets`

		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
				return nil, fmt.Errorf("gagal membuat direktori upload: %v", err)
			}
		}

		fileExt := filepath.Ext(file.Filename)
		fileName := fmt.Sprintf("l_%d_outlet%s", time.Now().Unix(), fileExt)
		fullPath = filepath.Join(uploadDir, fileName)

		if err := saveUploadedFile(file, fullPath); err != nil {
			return nil, fmt.Errorf("gagal menyimpan gambar: %v", err)
		}

		photoPath = "/uploads/outlets/" + fileName
	}

	outlet := model.Outlet{
		UserID:     userID,
		NamaOutlet: input.NamaOutlet,
		Alamat:     input.Alamat,
		Provinsi:   input.Provinsi,
		Kota:       input.Kota,
		Kecamatan:  input.Kecamatan,
		NomorHP:    input.NomorHP,
		IsAktif:    "active",
		Photo:      photoPath,
	}

	if err := database.DbCore.Create(&outlet).Error; err != nil {
		if fullPath != "" {
			os.Remove(fullPath)
		}
		return nil, fmt.Errorf("gagal menyimpan outlet ke database: %v", err)
	}

	if err := database.DbCore.Preload("User").First(&outlet, outlet.ID).Error; err != nil {
		return nil, fmt.Errorf("gagal load user data: %v", err)
	}

	outlet.Photo = normalizePhotoURL(outlet.Photo)

	return &outlet, nil
}

func GetOutletByID(outletID uint) (*model.Outlet, error) {
	var outlet model.Outlet
	if err := database.DbCore.Preload("User").First(&outlet, outletID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("outlet tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal mengambil outlet: %v", err)
	}

	outlet.Photo = normalizePhotoURL(outlet.Photo)

	return &outlet, nil
}

func GetOutletsByUserID(userID uint) ([]model.Outlet, error) {
	var outlets []model.Outlet
	if err := database.DbCore.Preload("User").Where("user_id = ?", userID).Find(&outlets).Error; err != nil {
		return nil, fmt.Errorf("gagal mengambil outlet: %v", err)
	}

	for i := range outlets {
		outlets[i].Photo = normalizePhotoURL(outlets[i].Photo)
	}

	return outlets, nil
}

func GetAllOutlets(page, limit int) ([]model.OutletList, int64, error) {
	var outlets []model.OutletList
	var total int64

	offset := (page - 1) * limit

	if err := database.DbCore.Model(&model.Outlet{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung total outlet: %v", err)
	}

	query := `
		SELECT id, user_id, nama_outlet, alamat, nomor_hp, is_aktif, photo,
		       DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at
		FROM outlets 
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	if err := database.DbCore.Raw(query, limit, offset).Scan(&outlets).Error; err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil outlet: %v", err)
	}

	for i := range outlets {
		outlets[i].Photo = normalizePhotoURL(outlets[i].Photo)
	}

	return outlets, total, nil
}

func UpdateOutlet(outletID uint, userID uint, input model.UpdateOutletInput, file *multipart.FileHeader) (*model.Outlet, error) {
	var outlet model.Outlet
	if err := database.DbCore.First(&outlet, outletID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("outlet tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal mengambil outlet: %v", err)
	}

	if outlet.UserID != userID {
		return nil, fmt.Errorf("anda tidak memiliki akses untuk mengupdate outlet ini")
	}

	if input.NamaOutlet != "" {
		outlet.NamaOutlet = input.NamaOutlet
	}
	if input.Alamat != "" {
		outlet.Alamat = input.Alamat
	}
	if input.NomorHP != "" {
		outlet.NomorHP = input.NomorHP
	}
	if input.Provinsi != "" {
		outlet.Provinsi = input.Provinsi
	}
	if input.Kota != "" {
		outlet.Kota = input.Kota
	}
	if input.Kecamatan != "" {
		outlet.Kecamatan = input.Kecamatan
	}
	if input.IsAktif != "" {
		outlet.IsAktif = input.IsAktif
	}

	if file != nil {
		if outlet.Photo != "" {
			oldFileName := filepath.Base(outlet.Photo)
			oldFilePath := filepath.Join(`C:\xampp\htdocs\Mobile-PipoSmart\uploads\outlets`, oldFileName)
			os.Remove(oldFilePath)
		}

		uploadDir := `C:\xampp\htdocs\Mobile-PipoSmart\uploads\outlets`
		fileExt := filepath.Ext(file.Filename)
		fileName := fmt.Sprintf("l_%d_outlet%s", time.Now().Unix(), fileExt)
		fullPath := filepath.Join(uploadDir, fileName)

		if err := saveUploadedFile(file, fullPath); err != nil {
			return nil, fmt.Errorf("gagal menyimpan gambar: %v", err)
		}

		outlet.Photo = "/uploads/outlets/" + fileName
	}

	if err := database.DbCore.Save(&outlet).Error; err != nil {
		return nil, fmt.Errorf("gagal mengupdate outlet: %v", err)
	}

	if err := database.DbCore.Preload("User").First(&outlet, outlet.ID).Error; err != nil {
		return nil, fmt.Errorf("gagal load user data: %v", err)
	}

	outlet.Photo = normalizePhotoURL(outlet.Photo)

	return &outlet, nil
}

func DeleteOutlet(outletID uint, userID uint) error {
	var outlet model.Outlet
	if err := database.DbCore.First(&outlet, outletID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("outlet tidak ditemukan")
		}
		return fmt.Errorf("gagal mengambil outlet: %v", err)
	}
	
	if outlet.UserID != userID {
		return fmt.Errorf("anda tidak memiliki akses untuk menghapus outlet ini")
	}

	if err := database.DbCore.Delete(&outlet).Error; err != nil {
		return fmt.Errorf("gagal menghapus outlet: %v", err)
	}

	return nil
}

func ActivateOutlet(outletID uint, userID uint) error {
	var outlet model.Outlet
	if err := database.DbCore.First(&outlet, outletID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("outlet tidak ditemukan")
		}
		return fmt.Errorf("gagal mengambil outlet: %v", err)
	}

	if outlet.UserID != userID {
		return fmt.Errorf("anda tidak memiliki akses untuk mengaktifkan outlet ini")
	}

	outlet.IsAktif = "active"
	if err := database.DbCore.Save(&outlet).Error; err != nil {
		return fmt.Errorf("gagal mengaktifkan outlet: %v", err)
	}

	return nil
}

func DeactivateOutlet(outletID uint, userID uint) error {
	var outlet model.Outlet
	if err := database.DbCore.First(&outlet, outletID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("outlet tidak ditemukan")
		}
		return fmt.Errorf("gagal mengambil outlet: %v", err)
	}

	if outlet.UserID != userID {
		return fmt.Errorf("anda tidak memiliki akses untuk menonaktifkan outlet ini")
	}

	outlet.IsAktif = "inactive"
	if err := database.DbCore.Save(&outlet).Error; err != nil {
		return fmt.Errorf("gagal menonaktifkan outlet: %v", err)
	}

	return nil
}

func saveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	img, err := imaging.Decode(src)
	if err != nil {
		return fmt.Errorf("gagal decode gambar: %v", err)
	}
	
	resized := imaging.Resize(img, 800, 0, imaging.Lanczos)
	err = imaging.Save(resized, dst, imaging.JPEGQuality(70))
	if err != nil {
		return fmt.Errorf("gagal menyimpan gambar yang di-resize: %v", err)
	}

	return nil
}