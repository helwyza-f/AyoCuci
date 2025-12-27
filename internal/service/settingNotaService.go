package service

import (
    "BackendFramework/internal/model"
    "BackendFramework/internal/database"
    "errors"
    "gorm.io/gorm"
)

type NotaSettingsService interface {
    GetByOutletID(outletID uint) (*model.NotaSettings, error)
    CreateOrUpdate(outletID uint, input *model.NotaSettingsInput) (*model.NotaSettings, error)
    Delete(outletID uint) error
}

type notaSettingsService struct {
    db *gorm.DB
}

func NewNotaSettingsService() NotaSettingsService {
    return &notaSettingsService{
        db: database.DbCore, 
    }
}


func (s *notaSettingsService) GetByOutletID(outletID uint) (*model.NotaSettings, error) {
    var settings model.NotaSettings

    err := s.db.Where("outlet_id = ?", outletID).First(&settings).Error
    if err == nil {
        return &settings, nil
    }

    if !errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, err
    }
    defaultSettings := model.NotaSettings{
        OutletID:           outletID,
        BusinessName:       "",
        Address:            "",
        Phone:              "",
        FooterNote:         "",
        WhatsappNote:       "",
        ShowLogo:           false,
        ShowQRCode:         true,
        ShowBusinessName:   true,
        ShowDescription:    true,
        ShowFooterNote:     true,
        ShowWhatsappFooter: true,
        PrinterSize:        58,
        PrinterType:        "A",
    }

    if err := s.db.Create(&defaultSettings).Error; err != nil {
        return nil, err
    }

    return &defaultSettings, nil
}

func (s *notaSettingsService) CreateOrUpdate(outletID uint, input *model.NotaSettingsInput) (*model.NotaSettings, error) {
    var outlet model.Outlet
    if err := s.db.First(&outlet, outletID).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("outlet not found")
        }
        return nil, err
    }
    
    var settings model.NotaSettings
    err := s.db.Where("outlet_id = ?", outletID).First(&settings).Error
    
    if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, err
    }
    
    if errors.Is(err, gorm.ErrRecordNotFound) {
        settings = model.NotaSettings{
            OutletID:           outletID,
            BusinessName:       input.BusinessName,
            Address:            input.Address,
            Phone:              input.Phone,
            FooterNote:         input.FooterNote,
            WhatsappNote:       input.WhatsappNote,
            ShowLogo:           input.ShowLogo,
            ShowQRCode:         input.ShowQRCode,
            ShowBusinessName:   input.ShowBusinessName,
            ShowDescription:    input.ShowDescription,
            ShowFooterNote:     input.ShowFooterNote,
            ShowWhatsappFooter: input.ShowWhatsappFooter,
            PrinterSize:        input.PrinterSize,
            PrinterType:        input.PrinterType,
        }
        
        if err := s.db.Create(&settings).Error; err != nil {
            return nil, err
        }
    } else {
        // Update existing
        updates := map[string]interface{}{
            "business_name":        input.BusinessName,
            "address":              input.Address,
            "phone":                input.Phone,
            "footer_note":          input.FooterNote,
            "whatsapp_note":        input.WhatsappNote,
            "show_logo":            input.ShowLogo,
            "show_qr_code":         input.ShowQRCode,
            "show_business_name":   input.ShowBusinessName,
            "show_description":     input.ShowDescription,
            "show_footer_note":     input.ShowFooterNote,
            "show_whatsapp_footer": input.ShowWhatsappFooter,
            "printer_size":         input.PrinterSize,
            "printer_type":         input.PrinterType,
        }
        
        if err := s.db.Model(&settings).Updates(updates).Error; err != nil {
            return nil, err
        }
    }
    

    if err := s.db.Where("outlet_id = ?", outletID).First(&settings).Error; err != nil {
        return nil, err
    }
    
    return &settings, nil
}


func (s *notaSettingsService) Delete(outletID uint) error {
    var settings model.NotaSettings
    
    err := s.db.Where("outlet_id = ?", outletID).First(&settings).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return errors.New("nota settings not found")
        }
        return err
    }
    
    if err := s.db.Delete(&settings).Error; err != nil {
        return err
    }
    
    return nil
}