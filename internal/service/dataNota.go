package service

import (
	"BackendFramework/internal/database"
	"BackendFramework/internal/model"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type NotaService interface {
    GenerateNota(input *model.NotaGenerateInput) (*model.NotaData, error)
    GetNotaByID(id uint) (*model.NotaData, error)
    GetNotaByTransactionID(transactionID string) (*model.NotaData, error)
    GetNotasByOutlet(outletID uint, page, limit int) ([]model.NotaData, int64, error)
    PrepareNotaForPrint(notaDataID uint) (*model.NotaPrintFormat, error)
    GenerateNotaPreview(input *model.NotaPreviewInput) (*model.NotaPrintFormat, error)
    VoidNota(notaDataID uint, reason string) error
    ReprintNota(notaDataID uint) (*model.NotaPrintFormat, error)
    GetPrintHistory(notaDataID uint) (*model.NotaData, error)
}

type notaService struct {
    db *gorm.DB
}

func NewNotaService(db *gorm.DB) NotaService {
    return &notaService{
		db: database.DbCore,
	}
}

func (s *notaService) GenerateNota(input *model.NotaGenerateInput) (*model.NotaData, error) {
    var outlet model.Outlet
    if err := s.db.First(&outlet, input.OutletID).Error; err != nil {
        return nil, errors.New("outlet not found")
    }

    var existing model.NotaData
    if err := s.db.Where("transaction_id = ?", input.TransactionID).First(&existing).Error; err == nil {
        return nil, errors.New("transaction ID already exists")
    }

    subtotal := 0.0
    for _, item := range input.Items {
        item.Subtotal = item.Quantity * item.Price
        subtotal += item.Subtotal
    }

    tax := input.Tax
    if input.TaxPercentage > 0 && tax == 0 {
        tax = subtotal * (input.TaxPercentage / 100)
    }

    discount := input.Discount
    if input.DiscountType == "percentage" {
        discount = subtotal * (input.Discount / 100)
    }

    total := subtotal + tax + input.ServiceCharge - discount

    change := input.PaymentAmount - total
    if change < 0 {
        return nil, errors.New("insufficient payment amount")
    }

    itemsJSON, err := json.Marshal(input.Items)
    if err != nil {
        return nil, err
    }

    notaData := &model.NotaData{
        OutletID:        input.OutletID,
        TransactionID:   input.TransactionID,
        TransactionDate: input.TransactionDate,
        CustomerName:    input.CustomerName,
        CustomerPhone:   input.CustomerPhone,
        CashierName:     input.CashierName,
        ItemsJSON:       string(itemsJSON),
        Subtotal:        subtotal,
        Tax:             tax,
        TaxPercentage:   input.TaxPercentage,
        Discount:        discount,
        DiscountType:    input.DiscountType,
        ServiceCharge:   input.ServiceCharge,
        Total:           total,
        PaymentAmount:   input.PaymentAmount,
        Change:          change,
        PaymentMethod:   input.PaymentMethod,
        Notes:           input.Notes,
        Status:          "completed",
        PrintCount:      0,
    }

    err = s.db.Transaction(func(tx *gorm.DB) error {

        if err := tx.Create(notaData).Error; err != nil {
            return err
        }
        for _, item := range input.Items {
            itemDetail := model.NotaItemDetail{
                NotaDataID:  notaData.ID,
                ProductName: item.ProductName,
                Quantity:    item.Quantity,
                Unit:        item.Unit,
                Price:       item.Price,
                Subtotal:    item.Subtotal,
                Description: item.Description,
            }
            if err := tx.Create(&itemDetail).Error; err != nil {
                return err
            }
        }

        return nil
    })

    if err != nil {
        return nil, err
    }
    if err := json.Unmarshal([]byte(notaData.ItemsJSON), &notaData.Items); err != nil {
        return nil, err
    }

    return notaData, nil
}


func (s *notaService) GetNotaByID(id uint) (*model.NotaData, error) {
    var notaData model.NotaData
    if err := s.db.Preload("Outlet").First(&notaData, id).Error; err != nil {
        return nil, err
    }

    if notaData.ItemsJSON != "" {
        if err := json.Unmarshal([]byte(notaData.ItemsJSON), &notaData.Items); err != nil {
            return nil, err
        }
    }

    return &notaData, nil
}


func (s *notaService) GetNotaByTransactionID(transactionID string) (*model.NotaData, error) {
    var notaData model.NotaData
    if err := s.db.Preload("Outlet").Where("transaction_id = ?", transactionID).First(&notaData).Error; err != nil {
        return nil, err
    }

    if notaData.ItemsJSON != "" {
        if err := json.Unmarshal([]byte(notaData.ItemsJSON), &notaData.Items); err != nil {
            return nil, err
        }
    }

    return &notaData, nil
}

func (s *notaService) GetNotasByOutlet(outletID uint, page, limit int) ([]model.NotaData, int64, error) {
    var notas []model.NotaData
    var total int64

    offset := (page - 1) * limit

 
    if err := s.db.Model(&model.NotaData{}).Where("outlet_id = ?", outletID).Count(&total).Error; err != nil {
        return nil, 0, err
    }


    if err := s.db.Where("outlet_id = ?", outletID).
        Order("created_at DESC").
        Limit(limit).
        Offset(offset).
        Find(&notas).Error; err != nil {
        return nil, 0, err
    }

    for i := range notas {
        if notas[i].ItemsJSON != "" {
            json.Unmarshal([]byte(notas[i].ItemsJSON), &notas[i].Items)
        }
    }

    return notas, total, nil
}


func (s *notaService) PrepareNotaForPrint(notaDataID uint) (*model.NotaPrintFormat, error) {
    notaData, err := s.GetNotaByID(notaDataID)
    if err != nil {
        return nil, err
    }

    var settings model.NotaSettings
    if err := s.db.Where("outlet_id = ?", notaData.OutletID).First(&settings).Error; err != nil {
        return nil, errors.New("nota settings not found for this outlet")
    }

    printFormat := &model.NotaPrintFormat{
        Settings:   &settings,
        Data:       notaData,
        PrintWidth: s.getPrintWidth(settings.PrinterSize),
    }

    printFormat.FormattedItems = s.formatItems(notaData.Items, printFormat.PrintWidth)

    printFormat.FormattedTotal = s.formatTotal(notaData, printFormat.PrintWidth)

    now := time.Now()
    s.db.Model(&model.NotaData{}).Where("id = ?", notaDataID).Updates(map[string]interface{}{
        "print_count":     gorm.Expr("print_count + 1"),
        "last_printed_at": now,
    })

    return printFormat, nil
}

func (s *notaService) GenerateNotaPreview(input *model.NotaPreviewInput) (*model.NotaPrintFormat, error) {
    var settings model.NotaSettings
    if err := s.db.Where("outlet_id = ?", input.OutletID).First(&settings).Error; err != nil {
        return nil, errors.New("nota settings not found for this outlet")
    }

    subtotal := 0.0
    for i, item := range input.Data.Items {
        input.Data.Items[i].Subtotal = item.Quantity * item.Price
        subtotal += input.Data.Items[i].Subtotal
    }

    tax := input.Data.Tax
    if input.Data.TaxPercentage > 0 && tax == 0 {
        tax = subtotal * (input.Data.TaxPercentage / 100)
    }

    discount := input.Data.Discount
    if input.Data.DiscountType == "percentage" {
        discount = subtotal * (input.Data.Discount / 100)
    }

    total := subtotal + tax + input.Data.ServiceCharge - discount
    change := input.Data.PaymentAmount - total

    notaData := &model.NotaData{
        TransactionID:   input.Data.TransactionID,
        TransactionDate: input.Data.TransactionDate,
        CustomerName:    input.Data.CustomerName,
        CustomerPhone:   input.Data.CustomerPhone,
        CashierName:     input.Data.CashierName,
        Items:           input.Data.Items,
        Subtotal:        subtotal,
        Tax:             tax,
        TaxPercentage:   input.Data.TaxPercentage,
        Discount:        discount,
        DiscountType:    input.Data.DiscountType,
        ServiceCharge:   input.Data.ServiceCharge,
        Total:           total,
        PaymentAmount:   input.Data.PaymentAmount,
        Change:          change,
        PaymentMethod:   input.Data.PaymentMethod,
        Notes:           input.Data.Notes,
    }

    printFormat := &model.NotaPrintFormat{
        Settings:   &settings,
        Data:       notaData,
        PrintWidth: s.getPrintWidth(settings.PrinterSize),
    }

    printFormat.FormattedItems = s.formatItems(notaData.Items, printFormat.PrintWidth)
    printFormat.FormattedTotal = s.formatTotal(notaData, printFormat.PrintWidth)

    return printFormat, nil
}

func (s *notaService) VoidNota(notaDataID uint, reason string) error {
    var notaData model.NotaData
    if err := s.db.First(&notaData, notaDataID).Error; err != nil {
        return err
    }

    if notaData.Status == "void" {
        return errors.New("nota already voided")
    }

    return s.db.Model(&notaData).Updates(map[string]interface{}{
        "status": "void",
        "notes":  notaData.Notes + " | VOID: " + reason,
    }).Error
}


func (s *notaService) ReprintNota(notaDataID uint) (*model.NotaPrintFormat, error) {
    return s.PrepareNotaForPrint(notaDataID)
}


func (s *notaService) GetPrintHistory(notaDataID uint) (*model.NotaData, error) {
    return s.GetNotaByID(notaDataID)
}


func (s *notaService) getPrintWidth(printerSize int) int {
    switch printerSize {
    case 58:
        return 32 
    case 80:
        return 48 
    default:
        return 32
    }
}

func (s *notaService) formatItems(items []model.NotaItem, width int) []string {
    var formatted []string
    for _, item := range items {
        line := fmt.Sprintf("%-*s %7.2f", width-8, item.ProductName, item.Subtotal)
        formatted = append(formatted, line)
        if item.Description != "" {
            formatted = append(formatted, fmt.Sprintf("  %s", item.Description))
        }
        formatted = append(formatted, fmt.Sprintf("  %.2f %s x Rp %.2f", item.Quantity, item.Unit, item.Price))
    }
    return formatted
}

func (s *notaService) formatTotal(nota *model.NotaData, width int) string {
    separator := ""
    for i := 0; i < width; i++ {
        separator += "-"
    }

    total := separator + "\n"
    total += fmt.Sprintf("%-*s %12.2f\n", width-13, "Subtotal", nota.Subtotal)
    
    if nota.Tax > 0 {
        taxLabel := "Tax"
        if nota.TaxPercentage > 0 {
            taxLabel = fmt.Sprintf("Tax (%.0f%%)", nota.TaxPercentage)
        }
        total += fmt.Sprintf("%-*s %12.2f\n", width-13, taxLabel, nota.Tax)
    }
    
    if nota.ServiceCharge > 0 {
        total += fmt.Sprintf("%-*s %12.2f\n", width-13, "Service", nota.ServiceCharge)
    }
    
    if nota.Discount > 0 {
        discLabel := "Discount"
        if nota.DiscountType == "percentage" {
            discLabel = fmt.Sprintf("Discount (%.0f%%)", nota.Discount)
        }
        total += fmt.Sprintf("%-*s -%11.2f\n", width-13, discLabel, nota.Discount)
    }
    
    total += separator + "\n"
    total += fmt.Sprintf("%-*s %12.2f\n", width-13, "TOTAL", nota.Total)
    total += fmt.Sprintf("%-*s %12.2f\n", width-13, "Payment", nota.PaymentAmount)
    total += fmt.Sprintf("%-*s %12.2f\n", width-13, "Change", nota.Change)
    total += separator

    return total
}