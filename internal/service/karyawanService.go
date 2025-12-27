package service

import (
	"BackendFramework/internal/model"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
)

type KaryawanService struct {
	DB *gorm.DB
}

func NewKaryawanService(db *gorm.DB) *KaryawanService {
	return &KaryawanService{DB: db}
}

func (s *KaryawanService) GetAll(outletID *uint) ([]model.KaryawanResponse, error) {
	var karyawans []model.Karyawan
	query := s.DB

	if outletID != nil {
		query = query.Where("kar_outlet = ?", *outletID)
	}

	if err := query.Order("kar_created DESC").Find(&karyawans).Error; err != nil {
		return nil, err
	}

	// Return empty array instead of nil when no data found
	responses := make([]model.KaryawanResponse, 0, len(karyawans))
	for _, k := range karyawans {
		responses = append(responses, k.ToResponse())
	}

	return responses, nil
}

func (s *KaryawanService) GetByID(id uint, outletID *uint) (*model.KaryawanResponse, error) {
	var karyawan model.Karyawan
	query := s.DB.Where("kar_id = ?", id)

	if outletID != nil {
		query = query.Where("kar_outlet = ?", *outletID)
	}

	if err := query.First(&karyawan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("karyawan tidak ditemukan")
		}
		return nil, err
	}

	response := karyawan.ToResponse()
	return &response, nil
}

func (s *KaryawanService) Create(input *model.KaryawanInput, outletID *uint, userUpdate string) (*model.KaryawanResponse, error) {
	// Validate input
	if input.Email == "" {
		return nil, errors.New("email tidak boleh kosong")
	}
	if input.Nama == "" {
		return nil, errors.New("nama tidak boleh kosong")
	}
	if input.Password == "" {
		return nil, errors.New("password tidak boleh kosong")
	}

	var existingKaryawan model.Karyawan
	if err := s.DB.Where("kar_email = ?", input.Email).First(&existingKaryawan).Error; err == nil {
		return nil, errors.New("email sudah digunakan")
	}

	permissionsJSON, err := json.Marshal(input.Permissions)
	if err != nil {
		return nil, errors.New("gagal memproses permissions")
	}

	karyawan := model.Karyawan{
		OutletID:    outletID,
		Nama:        input.Nama,
		Phone:       input.Phone,
		Email:       input.Email,
		Role:        "Karyawan",
		IsPremium:   input.IsPremium,
		Permissions: string(permissionsJSON),
		Status:      "Aktif",
		UserUpdate:  userUpdate,
	}

	if err := karyawan.HashPassword(input.Password); err != nil {
		return nil, errors.New("gagal mengenkripsi password")
	}
	if err := s.DB.Create(&karyawan).Error; err != nil {
		return nil, err
	}

	response := karyawan.ToResponse()
	return &response, nil
}

func (s *KaryawanService) Update(id uint, input *model.UpdateKaryawanInput, outletID *uint, userUpdate string) (*model.KaryawanResponse, error) {
	var karyawan model.Karyawan
	query := s.DB.Where("kar_id = ?", id)

	if outletID != nil {
		query = query.Where("kar_outlet = ?", *outletID)
	}

	if err := query.First(&karyawan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("karyawan tidak ditemukan")
		}
		return nil, err
	}
	
	if karyawan.Role == "Ownership" {
		return nil, errors.New("tidak dapat mengupdate akun ownership")
	}
	
	if input.Email != "" && input.Email != karyawan.Email {
		var existingKaryawan model.Karyawan
		if err := s.DB.Where("kar_email = ? AND kar_id != ?", input.Email, id).First(&existingKaryawan).Error; err == nil {
			return nil, errors.New("email sudah digunakan")
		}
	}
	
	if input.Nama != "" {
		karyawan.Nama = input.Nama
	}
	if input.Phone != "" {
		karyawan.Phone = input.Phone
	}
	if input.Email != "" {
		karyawan.Email = input.Email
	}
	if input.Password != "" {
		if err := karyawan.HashPassword(input.Password); err != nil {
			return nil, errors.New("gagal mengenkripsi password")
		}
	}
	if input.Status != "" {
		karyawan.Status = input.Status
	}

	karyawan.IsPremium = input.IsPremium

	if input.Permissions != nil {
		permissionsJSON, err := json.Marshal(input.Permissions)
		if err != nil {
			return nil, errors.New("gagal memproses permissions")
		}
		karyawan.Permissions = string(permissionsJSON)
	}

	karyawan.UserUpdate = userUpdate

	if err := s.DB.Save(&karyawan).Error; err != nil {
		return nil, err
	}

	response := karyawan.ToResponse()
	return &response, nil
}

func (s *KaryawanService) Delete(id uint, outletID *uint) error {
	var karyawan model.Karyawan
	query := s.DB.Where("kar_id = ?", id)

	if outletID != nil {
		query = query.Where("kar_outlet = ?", *outletID)
	}

	if err := query.First(&karyawan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("karyawan tidak ditemukan")
		}
		return err
	}

	if karyawan.Role == "Ownership" {
		return errors.New("tidak dapat menghapus akun ownership")
	}

	if err := s.DB.Delete(&karyawan).Error; err != nil {
		return err
	}

	return nil
}

func (s *KaryawanService) GetByEmail(email string) (*model.Karyawan, error) {
	var karyawan model.Karyawan
	if err := s.DB.Where("kar_email = ? AND kar_status = ?", email, "Aktif").First(&karyawan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("karyawan tidak ditemukan")
		}
		return nil, err
	}
	return &karyawan, nil
}