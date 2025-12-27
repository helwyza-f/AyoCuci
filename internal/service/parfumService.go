// File: internal/service/parfum_service.go
package service

import (
	"BackendFramework/internal/database"
	"BackendFramework/internal/model"
	"errors"

	"gorm.io/gorm"
)

type ParfumService interface {
	GetAllParfum(outletID uint) ([]model.ParfumResponse, error)
	GetParfumByID(id uint, outletID uint) (*model.ParfumResponse, error)
	CreateParfum(input model.ParfumInput, outletID uint, username string) (*model.ParfumResponse, error)
	UpdateParfum(id uint, input model.UpdateParfumInput, outletID uint, username string) (*model.ParfumResponse, error)
	DeleteParfum(id uint, outletID uint) error
}

type parfumService struct {
	db *gorm.DB
}

func NewParfumService() ParfumService {
	return &parfumService{
		db: database.DbCore, 
	}
}

func (s *parfumService) GetAllParfum(outletID uint) ([]model.ParfumResponse, error) {
	var parfums []model.Parfum
	
	query := s.db.Where("prf_outlet = ?", outletID)
	
	if err := query.Order("prf_id DESC").Find(&parfums).Error; err != nil {
		return nil, err
	}

	var responses []model.ParfumResponse
	for _, parfum := range parfums {
		responses = append(responses, parfum.ToResponse())
	}

	return responses, nil
}

func (s *parfumService) GetParfumByID(id uint, outletID uint) (*model.ParfumResponse, error) {
	var parfum model.Parfum
	
	if err := s.db.Where("prf_id = ? AND prf_outlet = ?", id, outletID).First(&parfum).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("parfum tidak ditemukan")
		}
		return nil, err
	}

	response := parfum.ToResponse()
	return &response, nil
}

func (s *parfumService) CreateParfum(input model.ParfumInput, outletID uint, username string) (*model.ParfumResponse, error) {
	parfum := model.Parfum{
		OutletID:   &outletID,
		Parfum:     input.Parfum,
		Keterangan: input.Keterangan,
		Status:     "Aktif",
		UserUpdate: username,
	}

	if input.Status != "" {
		parfum.Status = input.Status
	}

	if err := s.db.Create(&parfum).Error; err != nil {
		return nil, err
	}

	response := parfum.ToResponse()
	return &response, nil
}

func (s *parfumService) UpdateParfum(id uint, input model.UpdateParfumInput, outletID uint, username string) (*model.ParfumResponse, error) {
	var parfum model.Parfum
	
	if err := s.db.Where("prf_id = ? AND prf_outlet = ?", id, outletID).First(&parfum).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("parfum tidak ditemukan")
		}
		return nil, err
	}

	updates := map[string]interface{}{
		"prf_userupdate": username,
	}

	if input.Parfum != "" {
		updates["prf_nama"] = input.Parfum
	}
	if input.Keterangan != "" {
		updates["prf_keterangan"] = input.Keterangan
	}
	if input.Status != "" {
		updates["prf_status"] = input.Status
	}

	if err := s.db.Model(&parfum).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Reload data
	if err := s.db.Where("prf_id = ?", id).First(&parfum).Error; err != nil {
		return nil, err
	}

	response := parfum.ToResponse()
	return &response, nil
}

func (s *parfumService) DeleteParfum(id uint, outletID uint) error {
	var parfum model.Parfum
	
	if err := s.db.Where("prf_id = ? AND prf_outlet = ?", id, outletID).First(&parfum).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("parfum tidak ditemukan")
		}
		return err
	}

	if err := s.db.Delete(&parfum).Error; err != nil {
		return err
	}

	return nil
}