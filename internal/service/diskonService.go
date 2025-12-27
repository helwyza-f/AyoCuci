package service

import (
	"BackendFramework/internal/database"
	"BackendFramework/internal/model"
	"errors"

	"gorm.io/gorm"
)

type DiskonService struct {
    db *gorm.DB
}


func NewDiskonService() *DiskonService {
    return &DiskonService{
        db: database.DbCore, 
    }
}

func NewDiskonServiceWithDB(db *gorm.DB) *DiskonService {
    return &DiskonService{db: db}
}

func (s *DiskonService) GetAllDiskon(outletID uint) ([]model.Diskon, error) {
    var diskons []model.Diskon
    
    query := s.db.Where("dis_outlet = ?", outletID)
    
    if err := query.Order("dis_created DESC").Find(&diskons).Error; err != nil {
        return nil, err
    }
    
    return diskons, nil
}

func (s *DiskonService) GetDiskonByID(id uint, outletID uint) (*model.Diskon, error) {
    var diskon model.Diskon
    
    if err := s.db.Where("dis_id = ? AND dis_outlet = ?", id, outletID).First(&diskon).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("diskon tidak ditemukan")
        }
        return nil, err
    }
    
    return &diskon, nil
}

func (s *DiskonService) CreateDiskon(input model.DiskonInput, outletID uint, username string) (*model.Diskon, error) {
    if input.Jenis == "Persen" && input.NilaiDiskon > 100 {
        return nil, errors.New("nilai diskon persen tidak boleh lebih dari 100")
    }
    
    if input.Status == "" {
        input.Status = "Aktif"
    }
    
    diskon := model.Diskon{
        OutletID:    &outletID,
        Diskon:      input.Diskon,
        Jenis:       input.Jenis,
        NilaiDiskon: input.NilaiDiskon,
        Keterangan:  input.Keterangan,
        Status:      input.Status,
        UserUpdate:  username,
    }
    
    if err := s.db.Create(&diskon).Error; err != nil {
        return nil, err
    }
    
    return &diskon, nil
}

func (s *DiskonService) UpdateDiskon(id uint, input model.UpdateDiskonInput, outletID uint, username string) (*model.Diskon, error) {
    diskon, err := s.GetDiskonByID(id, outletID)
    if err != nil {
        return nil, err
    }
    
    if input.Jenis == "Persen" && input.NilaiDiskon > 100 {
        return nil, errors.New("nilai diskon persen tidak boleh lebih dari 100")
    }
    
    if input.Diskon != "" {
        diskon.Diskon = input.Diskon
    }
    if input.Jenis != "" {
        diskon.Jenis = input.Jenis
    }
    if input.NilaiDiskon > 0 {
        diskon.NilaiDiskon = input.NilaiDiskon
    }
    if input.Keterangan != "" {
        diskon.Keterangan = input.Keterangan
    }
    if input.Status != "" {
        diskon.Status = input.Status
    }
    
    diskon.UserUpdate = username
    
    if err := s.db.Save(&diskon).Error; err != nil {
        return nil, err
    }
    
    return diskon, nil
}

func (s *DiskonService) DeleteDiskon(id uint, outletID uint) error {
    _, err := s.GetDiskonByID(id, outletID)
    if err != nil {
        return err
    }
    
    if err := s.db.Where("dis_id = ? AND dis_outlet = ?", id, outletID).Delete(&model.Diskon{}).Error; err != nil {
        return err
    }
    
    return nil
}

func (s *DiskonService) GetActiveDiskon(outletID uint) ([]model.Diskon, error) {
    var diskons []model.Diskon
    
    if err := s.db.Where("dis_outlet = ? AND dis_status = ?", outletID, "Aktif").
        Order("dis_created DESC").
        Find(&diskons).Error; err != nil {
        return nil, err
    }
    
    return diskons, nil
}


func (s *DiskonService) GetDiskonByOutlet(outletID uint) ([]model.Diskon, error) {
    var diskons []model.Diskon
    
    if err := s.db.Where("dis_outlet = ?", outletID).
        Order("dis_created DESC").
        Find(&diskons).Error; err != nil {
        return nil, err
    }
    
    return diskons, nil
}

func (s *DiskonService) ToggleStatus(id uint, outletID uint, username string) (*model.Diskon, error) {
    diskon, err := s.GetDiskonByID(id, outletID)
    if err != nil {
        return nil, err
    }
    
    if diskon.Status == "Aktif" {
        diskon.Status = "Tidak Aktif"
    } else {
        diskon.Status = "Aktif"
    }
    
    diskon.UserUpdate = username
    
    if err := s.db.Save(&diskon).Error; err != nil {
        return nil, err
    }
    
    return diskon, nil
}