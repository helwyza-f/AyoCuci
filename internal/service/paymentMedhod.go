package service

import (
	"BackendFramework/internal/database"
	"BackendFramework/internal/model"
	"errors"

	"gorm.io/gorm"
)

type PaymentMethodService interface {
	GetAllByOutletID(outletID uint, filterActive *bool) ([]model.PaymentMethodList, error)
	GetByID(id uint, outletID uint) (*model.PaymentMethod, error)
	Create(outletID uint, input model.PaymentMethodInput) (*model.PaymentMethod, error)
	Update(id uint, outletID uint, input model.UpdatePaymentMethodInput) (*model.PaymentMethod, error)
	Delete(id uint, outletID uint) error
	ToggleActive(id uint, outletID uint, isActive bool) error
}

type paymentMethodService struct {
	db *gorm.DB
}

func NewPaymentMethodService() PaymentMethodService {
	return &paymentMethodService{
		db: database.DbCore, 
	}
}

func (s *paymentMethodService) GetAllByOutletID(outletID uint, filterActive *bool) ([]model.PaymentMethodList, error) {
    var paymentMethods []model.PaymentMethod
    
    query := s.db.Where("outlet_id = ?", outletID)
    
    if filterActive != nil {
        query = query.Where("is_active = ?", *filterActive)
    }
    
    if err := query.Order("created_at DESC").Find(&paymentMethods).Error; err != nil {
        return nil, err
    }
    
    result := make([]model.PaymentMethodList, 0)
    
    for _, pm := range paymentMethods {
        result = append(result, model.PaymentMethodList{
            ID:              pm.ID,
            OutletID:        pm.OutletID,
            Name:            pm.Name,
          
            IsActive:        pm.IsActive,
            Category:        pm.Category,
            BankName:        pm.BankName,
            AccountNumber:   pm.AccountNumber,
            AccountHolder:   pm.AccountHolder,
            EwalletProvider: pm.EwalletProvider,
            PhoneNumber:     pm.PhoneNumber,
            CreatedAt:       pm.CreatedAt.Format("2006-01-02 15:04:05"),
        })
    }
    
    return result, nil
}

func (s *paymentMethodService) GetByID(id uint, outletID uint) (*model.PaymentMethod, error) {
	var paymentMethod model.PaymentMethod
	
	if err := s.db.Where("id = ? AND outlet_id = ?", id, outletID).First(&paymentMethod).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment method not found")
		}
		return nil, err
	}
	
	return &paymentMethod, nil
}

func (s *paymentMethodService) Create(outletID uint, input model.PaymentMethodInput) (*model.PaymentMethod, error) {
	if err := s.validateInput(input.Category, input.BankName, input.AccountNumber, input.AccountHolder, 
		input.EwalletProvider, input.PhoneNumber); err != nil {
		return nil, err
	}
	
	paymentMethod := model.PaymentMethod{
		OutletID:        outletID,
		Name:            input.Name,
		
		IsActive:        input.IsActive,
		Category:        input.Category,
		BankName:        input.BankName,
		AccountNumber:   input.AccountNumber,
		AccountHolder:   input.AccountHolder,
		EwalletProvider: input.EwalletProvider,
		PhoneNumber:     input.PhoneNumber,
	}
	
	if err := s.db.Create(&paymentMethod).Error; err != nil {
		return nil, err
	}
	
	return &paymentMethod, nil
}

func (s *paymentMethodService) Update(id uint, outletID uint, input model.UpdatePaymentMethodInput) (*model.PaymentMethod, error) {
	paymentMethod, err := s.GetByID(id, outletID)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		paymentMethod.Name = *input.Name
	}
	if input.IsActive != nil {
		paymentMethod.IsActive = *input.IsActive
	}
	if input.Category != nil {
		paymentMethod.Category = *input.Category
	}
	
	paymentMethod.BankName = input.BankName
	paymentMethod.AccountNumber = input.AccountNumber
	paymentMethod.AccountHolder = input.AccountHolder
	paymentMethod.EwalletProvider = input.EwalletProvider
	paymentMethod.PhoneNumber = input.PhoneNumber
	
	if err := s.db.Save(&paymentMethod).Error; err != nil {
		return nil, err
	}
	
	return paymentMethod, nil
}

func (s *paymentMethodService) Delete(id uint, outletID uint) error {
	paymentMethod, err := s.GetByID(id, outletID)
	if err != nil {
		return err
	}
	
	if err := s.db.Delete(&paymentMethod).Error; err != nil {
		return err
	}
	
	return nil
}

func (s *paymentMethodService) ToggleActive(id uint, outletID uint, isActive bool) error {
	paymentMethod, err := s.GetByID(id, outletID)
	if err != nil {
		return err
	}
	
	paymentMethod.IsActive = isActive
	
	if err := s.db.Save(&paymentMethod).Error; err != nil {
		return err
	}
	
	return nil
}

func (s *paymentMethodService) validateInput(category string, bankName, accountNumber, accountHolder, 
	ewalletProvider, phoneNumber *string) error {
	
	switch category {
	case "Transfer":
		if bankName == nil || *bankName == "" {
			return errors.New("bank name is required for Transfer category")
		}
		if accountNumber == nil || *accountNumber == "" {
			return errors.New("account number is required for Transfer category")
		}
		if accountHolder == nil || *accountHolder == "" {
			return errors.New("account holder is required for Transfer category")
		}
	case "E-Wallet":
		if ewalletProvider == nil || *ewalletProvider == "" {
			return errors.New("e-wallet provider is required for E-Wallet category")
		}
		if phoneNumber == nil || *phoneNumber == "" {
			return errors.New("phone number is required for E-Wallet category")
		}
	}
	
	return nil
}