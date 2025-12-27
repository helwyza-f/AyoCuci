package service

import (
	"BackendFramework/internal/database"
	"BackendFramework/internal/model"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

const BaseUploadDir = `C:\xampp\htdocs\Mobile-PipoSmart\uploads\layanan`
const RelativePhotoPath = "uploads/layanan"

type LayananService struct {
	DB *gorm.DB
}

func NewLayananService() *LayananService {
	return &LayananService{
		DB: database.DbCore,
	}
}

func (s *LayananService) CreateLayananWithProducts(input *model.CreateLayananWithProductsInput, userID uint, files map[string]*multipart.FileHeader) (*model.Layanan, error) {
	var outlet model.Outlet
	if err := s.DB.First(&outlet, input.OutletID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("outlet tidak ditemukan")
		}
		return nil, err
	}

	layanan := &model.Layanan{
		OutletID:     input.OutletID,
		NamaLayanan:  input.NamaLayanan,
		Prioritas:    input.Prioritas,
		Cuci:         input.Cuci,
		Kering:       input.Kering,
		Setrika:      input.Setrika,
		UserUpdateID: &userID,
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(layanan).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var uploadedFiles []string

	for i, produkInput := range input.JenisProduk {
		jenisProduk := &model.JenisProduk{
			LayananID:      layanan.ID,
			Nama:           produkInput.Nama,
			Satuan:         produkInput.Satuan,
			HargaPer:       produkInput.HargaPer,
			LamaPengerjaan: produkInput.LamaPengerjaan,
			SatuanWaktu:    produkInput.SatuanWaktu,
			Keterangan:     produkInput.Keterangan,
		}

		fileKey := fmt.Sprintf("produk_%d", i)

		if file, exists := files[fileKey]; exists && file != nil {
			path, err := s.handleFileUpload(userID, file)
			if err != nil {
				tx.Rollback()
				s.cleanupFiles(uploadedFiles)
				return nil, fmt.Errorf("gagal upload file produk %s: %v", produkInput.Nama, err)
			}
			jenisProduk.ImageURL = &path
			jenisProduk.IconPath = nil
			uploadedFiles = append(uploadedFiles, path)
		} else if produkInput.IconPath != nil && *produkInput.IconPath != "" {
			jenisProduk.IconPath = produkInput.IconPath
			jenisProduk.ImageURL = nil
		}

		if err := tx.Create(jenisProduk).Error; err != nil {
			tx.Rollback()
			s.cleanupFiles(uploadedFiles)
			return nil, fmt.Errorf("gagal simpan produk %s: %v", produkInput.Nama, err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		s.cleanupFiles(uploadedFiles)
		return nil, err
	}

	if err := s.DB.Preload("Outlet").
		Preload("UserUpdate").
		Preload("JenisProduk").
		First(layanan, layanan.ID).Error; err != nil {
		return nil, err
	}

	return layanan, nil
}

func (s *LayananService) UpdateLayananWithProducts(id uint, input *model.UpdateLayananWithProductsInput, userID uint, files map[string]*multipart.FileHeader) (*model.Layanan, error) {
	var layanan model.Layanan
	if err := s.DB.Preload("JenisProduk").First(&layanan, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("layanan tidak ditemukan")
		}
		return nil, err
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if input.NamaLayanan != "" {
		layanan.NamaLayanan = input.NamaLayanan
	}
	if input.Prioritas != nil {
		layanan.Prioritas = *input.Prioritas
	}
	if input.Cuci != "" {
		layanan.Cuci = input.Cuci
	}
	if input.Kering != "" {
		layanan.Kering = input.Kering
	}
	if input.Setrika != "" {
		layanan.Setrika = input.Setrika
	}
	layanan.UserUpdateID = &userID

	if err := tx.Save(&layanan).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var uploadedFiles []string
	var filesToDelete []string

	existingProductsMap := make(map[uint]*model.JenisProduk)
	for i := range layanan.JenisProduk {
		existingProductsMap[layanan.JenisProduk[i].ID] = &layanan.JenisProduk[i]
	}

	processedFromInput := make(map[uint]bool)

	for i, produkInput := range input.JenisProduk {
		if produkInput.ShouldDelete && produkInput.ID != nil {
			var produk model.JenisProduk
			if err := tx.First(&produk, *produkInput.ID).Error; err == nil {
				if produk.ImageURL != nil && *produk.ImageURL != "" {
					filesToDelete = append(filesToDelete, *produk.ImageURL)
				}
				if err := tx.Delete(&produk).Error; err != nil {
					tx.Rollback()
					s.cleanupFiles(uploadedFiles)
					return nil, fmt.Errorf("gagal hapus produk: %v", err)
				}
				processedFromInput[*produkInput.ID] = true
			}
			continue
		}

		if produkInput.Nama == "" {
			tx.Rollback()
			s.cleanupFiles(uploadedFiles)
			return nil, fmt.Errorf("produk %d: nama tidak valid", i)
		}

		if produkInput.ID != nil && *produkInput.ID > 0 {
			existingProduct, exists := existingProductsMap[*produkInput.ID]
			if !exists {
				tx.Rollback()
				s.cleanupFiles(uploadedFiles)
				return nil, fmt.Errorf("produk dengan ID %d tidak ditemukan di layanan ini", *produkInput.ID)
			}

			processedFromInput[*produkInput.ID] = true
			existingProduct.Nama = produkInput.Nama
			if produkInput.Satuan != nil {
				existingProduct.Satuan = produkInput.Satuan
			}
			if produkInput.HargaPer != nil {
				existingProduct.HargaPer = produkInput.HargaPer
			}
			if produkInput.LamaPengerjaan != nil {
				existingProduct.LamaPengerjaan = produkInput.LamaPengerjaan
			}
			if produkInput.SatuanWaktu != nil {
				existingProduct.SatuanWaktu = produkInput.SatuanWaktu
			}
			if produkInput.Keterangan != nil {
				existingProduct.Keterangan = produkInput.Keterangan
			}

			fileKey := fmt.Sprintf("produk_%d", i)

			if file, fileExists := files[fileKey]; fileExists && file != nil {
				if existingProduct.ImageURL != nil && *existingProduct.ImageURL != "" {
					filesToDelete = append(filesToDelete, *existingProduct.ImageURL)
				}

				path, err := s.handleFileUpload(userID, file)
				if err != nil {
					tx.Rollback()
					s.cleanupFiles(uploadedFiles)
					return nil, fmt.Errorf("gagal upload file: %v", err)
				}

				existingProduct.ImageURL = &path
				existingProduct.IconPath = nil
				uploadedFiles = append(uploadedFiles, path)

			} else if produkInput.IconPath != nil && *produkInput.IconPath != "" {
				shouldUpdateIcon := existingProduct.IconPath == nil || *existingProduct.IconPath != *produkInput.IconPath
				switchingToIcon := existingProduct.ImageURL != nil && *existingProduct.ImageURL != ""

				if shouldUpdateIcon || switchingToIcon {
					if existingProduct.ImageURL != nil && *existingProduct.ImageURL != "" {
						filesToDelete = append(filesToDelete, *existingProduct.ImageURL)
					}
					existingProduct.IconPath = produkInput.IconPath
					existingProduct.ImageURL = nil
				}
			}

			if err := tx.Save(existingProduct).Error; err != nil {
				tx.Rollback()
				s.cleanupFiles(uploadedFiles)
				return nil, fmt.Errorf("gagal update produk: %v", err)
			}

		} else {
			jenisProduk := &model.JenisProduk{
				LayananID:      layanan.ID,
				Nama:           produkInput.Nama,
				Satuan:         produkInput.Satuan,
				HargaPer:       produkInput.HargaPer,
				LamaPengerjaan: produkInput.LamaPengerjaan,
				SatuanWaktu:    produkInput.SatuanWaktu,
				Keterangan:     produkInput.Keterangan,
			}

			fileKey := fmt.Sprintf("produk_%d", i)

			if file, fileExists := files[fileKey]; fileExists && file != nil {
				path, err := s.handleFileUpload(userID, file)
				if err != nil {
					tx.Rollback()
					s.cleanupFiles(uploadedFiles)
					return nil, fmt.Errorf("gagal upload file: %v", err)
				}
				jenisProduk.ImageURL = &path
				jenisProduk.IconPath = nil
				uploadedFiles = append(uploadedFiles, path)
			} else if produkInput.IconPath != nil && *produkInput.IconPath != "" {
				jenisProduk.IconPath = produkInput.IconPath
				jenisProduk.ImageURL = nil
			}

			if err := tx.Create(jenisProduk).Error; err != nil {
				tx.Rollback()
				s.cleanupFiles(uploadedFiles)
				return nil, fmt.Errorf("gagal create produk: %v", err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		s.cleanupFiles(uploadedFiles)
		return nil, err
	}

	s.cleanupFiles(filesToDelete)

	if err := s.DB.Preload("Outlet").
		Preload("UserUpdate").
		Preload("JenisProduk").
		First(&layanan, id).Error; err != nil {
		return nil, err
	}

	return &layanan, nil
}

func (s *LayananService) CreateLayanan(input *model.CreateLayananInput, userID uint, file *multipart.FileHeader) (*model.Layanan, error) {
	var outlet model.Outlet
	if err := s.DB.First(&outlet, input.OutletID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("outlet tidak ditemukan")
		}
		return nil, err
	}

	layanan := &model.Layanan{
		OutletID:     input.OutletID,
		NamaLayanan:  input.NamaLayanan,
		Prioritas:    input.Prioritas,
		Cuci:         input.Cuci,
		Kering:       input.Kering,
		Setrika:      input.Setrika,
		UserUpdateID: &userID,
	}

	if err := s.DB.Create(layanan).Error; err != nil {
		return nil, err
	}

	if err := s.DB.Preload("Outlet").
		Preload("UserUpdate").
		Preload("JenisProduk").
		First(layanan, layanan.ID).Error; err != nil {
		return nil, err
	}

	return layanan, nil
}

func (s *LayananService) UpdateLayanan(id uint, input *model.UpdateLayananInput, userID uint, file *multipart.FileHeader) (*model.Layanan, error) {
	var layanan model.Layanan
	if err := s.DB.First(&layanan, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("layanan tidak ditemukan")
		}
		return nil, err
	}

	if input.NamaLayanan != "" {
		layanan.NamaLayanan = input.NamaLayanan
	}
	if input.Prioritas != nil {
		layanan.Prioritas = *input.Prioritas
	}
	if input.Cuci != "" {
		layanan.Cuci = input.Cuci
	}
	if input.Kering != "" {
		layanan.Kering = input.Kering
	}
	if input.Setrika != "" {
		layanan.Setrika = input.Setrika
	}
	layanan.UserUpdateID = &userID

	if err := s.DB.Save(&layanan).Error; err != nil {
		return nil, err
	}

	if err := s.DB.Preload("Outlet").
		Preload("UserUpdate").
		Preload("JenisProduk").
		First(&layanan, id).Error; err != nil {
		return nil, err
	}

	return &layanan, nil
}

func (s *LayananService) GetAllLayanan(outletID uint, page, limit int) ([]model.LayananList, int64, error) {
	var layanan []model.Layanan
	var total int64

	query := s.DB.Model(&model.Layanan{})
	if outletID > 0 {
		query = query.Where("outlet_id = ?", outletID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.
		Preload("JenisProduk").
		Offset(offset).
		Limit(limit).
		Order("id DESC").
		Find(&layanan).Error; err != nil {
		return nil, 0, err
	}

	result := make([]model.LayananList, len(layanan))
	for i, lay := range layanan {
		result[i] = model.LayananList{
			ID:          lay.ID,
			OutletID:    lay.OutletID,
			NamaLayanan: lay.NamaLayanan,
			Prioritas:   lay.Prioritas,
			Cuci:        lay.Cuci,
			Kering:      lay.Kering,
			Setrika:     lay.Setrika,
			CreatedAt:   lay.CreatedAt,
			LastUpdate:  lay.LastUpdate,
			JenisProduk: lay.JenisProduk,
		}
	}

	return result, total, nil
}

func (s *LayananService) GetLayananByID(id uint) (*model.LayananDetail, error) {
	var layanan model.Layanan
	if err := s.DB.Preload("Outlet").
		Preload("UserUpdate").
		Preload("JenisProduk").
		First(&layanan, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("layanan tidak ditemukan")
		}
		return nil, err
	}

	detail := &model.LayananDetail{
		ID:           layanan.ID,
		OutletID:     layanan.OutletID,
		NamaLayanan:  layanan.NamaLayanan,
		Prioritas:    layanan.Prioritas,
		Cuci:         layanan.Cuci,
		Kering:       layanan.Kering,
		Setrika:      layanan.Setrika,
		CreatedAt:    layanan.CreatedAt,
		LastUpdate:   layanan.LastUpdate,
		UserUpdateID: layanan.UserUpdateID,
		Outlet:       &layanan.Outlet,
		JenisProduk:  layanan.JenisProduk,
	}

	if layanan.UserUpdate != nil {
		detail.UserUpdate = &model.UserList{
			ID:          layanan.UserUpdate.ID,
			NamaLengkap: layanan.UserUpdate.NamaLengkap,
			Email:       layanan.UserUpdate.Email,
			NomorHP:     layanan.UserUpdate.NomorHP,
			Group:       layanan.UserUpdate.Group,
			IsAktif:     layanan.UserUpdate.IsAktif,
		}
	}

	return detail, nil
}

func (s *LayananService) GetLayananByOutlet(outletID uint) ([]model.LayananList, error) {
	var layanan []model.Layanan
	if err := s.DB.Model(&model.Layanan{}).
		Preload("JenisProduk").
		Where("outlet_id = ?", outletID).
		Order("prioritas DESC, ln_layanan ASC").
		Find(&layanan).Error; err != nil {
		return nil, err
	}

	result := make([]model.LayananList, len(layanan))
	for i, lay := range layanan {
		result[i] = model.LayananList{
			ID:          lay.ID,
			OutletID:    lay.OutletID,
			NamaLayanan: lay.NamaLayanan,
			Prioritas:   lay.Prioritas,
			Cuci:        lay.Cuci,
			Kering:      lay.Kering,
			Setrika:     lay.Setrika,
			CreatedAt:   lay.CreatedAt,
			LastUpdate:  lay.LastUpdate,
			JenisProduk: lay.JenisProduk,
		}
	}

	return result, nil
}

func (s *LayananService) DeleteLayanan(id uint) error {
	var layanan model.Layanan
	if err := s.DB.Preload("JenisProduk").First(&layanan, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("layanan tidak ditemukan")
		}
		return err
	}

	var filesToDelete []string
	for _, produk := range layanan.JenisProduk {
		if produk.ImageURL != nil && *produk.ImageURL != "" {
			filesToDelete = append(filesToDelete, *produk.ImageURL)
		}
	}

	if err := s.DB.Delete(&layanan).Error; err != nil {
		return err
	}

	s.cleanupFiles(filesToDelete)
	return nil
}

func (s *LayananService) handleFileUpload(userID uint, file *multipart.FileHeader) (string, error) {
	if _, err := os.Stat(BaseUploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(BaseUploadDir, os.ModePerm); err != nil {
			return "", err
		}
	}

	fileExt := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), fileExt)
	fullPath := filepath.Join(BaseUploadDir, fileName)

	if err := saveUploadedFileLayanan(file, fullPath); err != nil {
		return "", err
	}

	return filepath.ToSlash(filepath.Join(RelativePhotoPath, fileName)), nil
}

func (s *LayananService) getFullPath(relativePath string) string {
	fileName := filepath.Base(relativePath)
	return filepath.Join(BaseUploadDir, fileName)
}

func (s *LayananService) cleanupFiles(files []string) {
	for _, file := range files {
		fullPath := s.getFullPath(file)
		_ = os.Remove(fullPath)
	}
}

func saveUploadedFileLayanan(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}