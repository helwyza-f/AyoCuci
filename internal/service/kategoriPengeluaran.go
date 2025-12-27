package service

import (
    "errors"
    "BackendFramework/internal/database"
    "BackendFramework/internal/model"
    "gorm.io/gorm"
)

type KategoriPengeluaranService struct {
    db *gorm.DB
}

func NewKategoriPengeluaranService() *KategoriPengeluaranService {
    return &KategoriPengeluaranService{
        db: database.DbCore,
    }
}

func (s *KategoriPengeluaranService) Create(input model.KategoriPengeluaranInput, outletID *uint, username string) (*model.KategoriPengeluaran, error) {
    kategori := &model.KategoriPengeluaran{
        OutletID:   outletID,
        Kategori:   input.Kategori,
       
        Status:     input.Status,
        UserUpdate: username,
    }

    if kategori.Status == "" {
        kategori.Status = "Aktif"
    }

    if err := s.db.Create(kategori).Error; err != nil {
        return nil, err
    }

    return kategori, nil
}

func (s *KategoriPengeluaranService) GetAll(outletID *uint) ([]model.KategoriPengeluaran, error) {
    var kategoris []model.KategoriPengeluaran

    query := s.db.Preload("Outlet")

    if outletID != nil {
        query = query.Where("ktg_outlet = ?", *outletID)
    }

    if err := query.Order("ktg_id DESC").Find(&kategoris).Error; err != nil {
        return nil, err
    }

    return kategoris, nil
}

func (s *KategoriPengeluaranService) GetByID(id uint, outletID *uint) (*model.KategoriPengeluaran, error) {
    var kategori model.KategoriPengeluaran

    query := s.db.Preload("Outlet")

    if outletID != nil {
        query = query.Where("ktg_outlet = ?", *outletID)
    }

    if err := query.First(&kategori, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("kategori pengeluaran tidak ditemukan")
        }
        return nil, err
    }

    return &kategori, nil
}

func (s *KategoriPengeluaranService) Update(id uint, input model.UpdateKategoriPengeluaranInput, outletID *uint, username string) (*model.KategoriPengeluaran, error) {
    kategori, err := s.GetByID(id, outletID)
    if err != nil {
        return nil, err
    }

    updates := make(map[string]interface{})

    if input.Kategori != "" {
        updates["ktg_nama"] = input.Kategori
    }
    if input.Status != "" {
        updates["ktg_status"] = input.Status
    }

    updates["ktg_userupdate"] = username

    if err := s.db.Model(kategori).Updates(updates).Error; err != nil {
        return nil, err
    }

    if err := s.db.Preload("Outlet").First(kategori, id).Error; err != nil {
        return nil, err
    }

    return kategori, nil
}

func (s *KategoriPengeluaranService) Delete(id uint, outletID *uint) error {
    kategori, err := s.GetByID(id, outletID)
    if err != nil {
        return err
    }

    if err := s.db.Delete(kategori).Error; err != nil {
        return err
    }

    return nil
}

func (s *KategoriPengeluaranService) GetByStatus(status string, outletID *uint) ([]model.KategoriPengeluaran, error) {
    var kategoris []model.KategoriPengeluaran

    query := s.db.Preload("Outlet").Where("ktg_status = ?", status)

    if outletID != nil {
        query = query.Where("ktg_outlet = ?", *outletID)
    }

    if err := query.Order("ktg_id DESC").Find(&kategoris).Error; err != nil {
        return nil, err
    }

    return kategoris, nil
}