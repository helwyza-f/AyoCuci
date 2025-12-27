package controller

import (
	"BackendFramework/internal/model"
	"BackendFramework/internal/service"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"strconv"
	"strings"
)

type NotaController struct {
	notaService service.NotaService
	validate    *validator.Validate
}

func NewNotaController(notaService service.NotaService) *NotaController {
	return &NotaController{
		notaService: notaService,
		validate:    validator.New(),
	}
}

func (c *NotaController) GenerateNota(ctx *gin.Context) {
	var input model.NotaGenerateInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, model.NotaSettingsErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if err := c.validate.Struct(input); err != nil {
		ctx.JSON(http.StatusBadRequest, model.NotaSettingsErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	notaData, err := c.notaService.GenerateNota(&input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.NotaSettingsErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, model.NotaGenerateResponse{
		Success: true,
		Message: "Nota generated successfully",
		Data:    notaData,
	})
}

func (c *NotaController) GetNotaByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.NotaSettingsErrorResponse{
			Success: false,
			Message: "Invalid nota ID",
		})
		return
	}

	notaData, err := c.notaService.GetNotaByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, model.NotaSettingsErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, model.NotaGenerateResponse{
		Success: true,
		Message: "Nota retrieved successfully",
		Data:    notaData,
	})
}

func (c *NotaController) GetNotaByTransactionID(ctx *gin.Context) {
	transactionID := ctx.Param("transaction_id")

	notaData, err := c.notaService.GetNotaByTransactionID(transactionID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, model.NotaSettingsErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, model.NotaGenerateResponse{
		Success: true,
		Message: "Nota retrieved successfully",
		Data:    notaData,
	})
}

func (c *NotaController) GetNotasByOutlet(ctx *gin.Context) {
	outletID, err := strconv.ParseUint(ctx.Param("outlet_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.NotaSettingsErrorResponse{
			Success: false,
			Message: "Invalid outlet ID",
		})
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	notas, total, err := c.notaService.GetNotasByOutlet(uint(outletID), page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.NotaSettingsErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Notas retrieved successfully",
		"data":    notas,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_page": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func (c *NotaController) PrintNota(ctx *gin.Context) {
	var input model.NotaPrintInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, model.NotaSettingsErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if err := c.validate.Struct(input); err != nil {
		ctx.JSON(http.StatusBadRequest, model.NotaSettingsErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	var printFormat *model.NotaPrintFormat
	var err error

	if input.Reprint {
		printFormat, err = c.notaService.ReprintNota(input.NotaDataID)
	} else {
		printFormat, err = c.notaService.PrepareNotaForPrint(input.NotaDataID)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.NotaSettingsErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	printData := c.generatePrintData(printFormat)

	ctx.JSON(http.StatusOK, model.NotaPrintResponse{
		Success:   true,
		Message:   "Nota prepared for printing",
		PrintData: printData,
		Format:    "text",
	})
}

func (c *NotaController) PreviewNota(ctx *gin.Context) {
	var input model.NotaPreviewInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, model.NotaSettingsErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if err := c.validate.Struct(input); err != nil {
		ctx.JSON(http.StatusBadRequest, model.NotaSettingsErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	printFormat, err := c.notaService.GenerateNotaPreview(&input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.NotaSettingsErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	previewData := c.generatePrintData(printFormat)
	previewHTML := c.generateHTMLPreview(printFormat)

	ctx.JSON(http.StatusOK, model.NotaPreviewResponse{
		Success:     true,
		Message:     "Preview generated successfully",
		PreviewHTML: previewHTML,
		PreviewData: previewData,
	})
}

func (c *NotaController) VoidNota(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.NotaSettingsErrorResponse{
			Success: false,
			Message: "Invalid nota ID",
		})
		return
	}

	var body struct {
		Reason string `json:"reason" validate:"required"`
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, model.NotaSettingsErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if err := c.notaService.VoidNota(uint(id), body.Reason); err != nil {
		ctx.JSON(http.StatusInternalServerError, model.NotaSettingsErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Nota voided successfully",
	})
}

func (c *NotaController) generatePrintData(format *model.NotaPrintFormat) string {
	settings := format.Settings
	data := format.Data

	var output string

	if settings.ShowLogo && format.LogoBase64 != "" {
		output += "[LOGO]\n"
	}

	if settings.ShowBusinessName && settings.BusinessName != "" {
		output += centerText(settings.BusinessName, format.PrintWidth) + "\n"
	}

	if settings.Address != "" {
		output += centerText(settings.Address, format.PrintWidth) + "\n"
	}

	if settings.Phone != "" {
		output += centerText(settings.Phone, format.PrintWidth) + "\n"
	}

	output += repeatChar("=", format.PrintWidth) + "\n"
	output += fmt.Sprintf("No: %s\n", data.TransactionID)
	output += fmt.Sprintf("Date: %s\n", data.TransactionDate.Format("02/01/2006 15:04"))
	output += fmt.Sprintf("Cashier: %s\n", data.CashierName)

	if data.CustomerName != "" {
		output += fmt.Sprintf("Customer: %s\n", data.CustomerName)
	}

	output += repeatChar("-", format.PrintWidth) + "\n"

	for _, line := range format.FormattedItems {
		output += line + "\n"
	}

	output += format.FormattedTotal + "\n"
	output += fmt.Sprintf("Payment: %s\n", data.PaymentMethod)

	output += repeatChar("=", format.PrintWidth) + "\n"

	if settings.ShowFooterNote && settings.FooterNote != "" {
		output += centerText(settings.FooterNote, format.PrintWidth) + "\n"
	}

	if settings.ShowQRCode && data.QRCodeData != "" {
		output += "[QR CODE]\n"
	}

	if settings.ShowWhatsappFooter && settings.WhatsappNote != "" {
		output += centerText(settings.WhatsappNote, format.PrintWidth) + "\n"
	}

	output += centerText("Thank You!", format.PrintWidth) + "\n"

	return output
}

func (c *NotaController) generateHTMLPreview(format *model.NotaPrintFormat) string {
	settings := format.Settings
	data := format.Data
	width := format.PrintWidth

	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Courier New', monospace;
            background: #f5f5f5;
            padding: 20px;
        }
        .receipt {
            background: white;
            width: ` + strconv.Itoa(width*8) + `px;
            margin: 0 auto;
            padding: 15px;
            box-shadow: 0 0 10px rgba(0,0,0,0.1);
        }
        .center {
            text-align: center;
        }
        .logo {
            text-align: center;
            margin-bottom: 10px;
        }
        .logo img {
            max-width: 100px;
            height: auto;
        }
        .qr-code {
            text-align: center;
            margin: 15px 0;
        }
        .qr-code img {
            max-width: 120px;
            height: auto;
        }
        .business-name {
            font-size: 16px;
            font-weight: bold;
            margin-bottom: 5px;
        }
        .business-info {
            font-size: 11px;
            margin-bottom: 3px;
            color: #333;
        }
        .separator {
            border-top: 1px dashed #000;
            margin: 10px 0;
        }
        .separator-solid {
            border-top: 2px solid #000;
            margin: 10px 0;
        }
        .transaction-info {
            font-size: 11px;
            margin: 5px 0;
        }
        .info-row {
            display: flex;
            justify-content: space-between;
            margin: 3px 0;
            font-size: 11px;
        }
        .info-label {
            color: #666;
        }
        .info-value {
            color: #000;
            font-weight: 500;
        }
        .items-section {
            margin: 10px 0;
        }
        .item {
            margin: 8px 0;
            font-size: 11px;
        }
        .item-header {
            display: flex;
            justify-content: space-between;
            font-weight: bold;
            margin-bottom: 2px;
        }
        .item-name {
            flex: 1;
        }
        .item-price {
            text-align: right;
            white-space: nowrap;
        }
        .item-details {
            color: #666;
            font-size: 10px;
            margin-left: 5px;
        }
        .totals-section {
            margin-top: 10px;
        }
        .total-row {
            display: flex;
            justify-content: space-between;
            margin: 5px 0;
            font-size: 11px;
        }
        .total-row.grand-total {
            font-size: 13px;
            font-weight: bold;
            margin-top: 8px;
            padding-top: 8px;
            border-top: 2px solid #000;
        }
        .total-label {
            color: #333;
        }
        .total-value {
            font-weight: 500;
        }
        .payment-section {
            margin-top: 10px;
            padding-top: 10px;
            border-top: 1px dashed #000;
        }
        .payment-row {
            display: flex;
            justify-content: space-between;
            margin: 5px 0;
            font-size: 12px;
        }
        .footer {
            margin-top: 15px;
            padding-top: 10px;
            border-top: 1px dashed #000;
        }
        .footer-note {
            text-align: center;
            font-size: 10px;
            margin: 5px 0;
            color: #666;
        }
        .thank-you {
            text-align: center;
            font-size: 12px;
            font-weight: bold;
            margin-top: 10px;
        }
        @media print {
            body {
                background: white;
                padding: 0;
            }
            .receipt {
                box-shadow: none;
                width: 100%;
            }
        }
    </style>
</head>
<body>
    <div class="receipt">`

	if settings.ShowLogo && format.LogoBase64 != "" {
		html += `
        <div class="logo">
            <img src="` + format.LogoBase64 + `" alt="Logo">
        </div>`
	}
	if settings.ShowBusinessName && settings.BusinessName != "" {
		html += `
        <div class="center business-name">` + settings.BusinessName + `</div>`
	}

	if settings.Address != "" {
		html += `
        <div class="center business-info">` + settings.Address + `</div>`
	}

	if settings.Phone != "" {
		html += `
        <div class="center business-info">` + settings.Phone + `</div>`
	}

	html += `<div class="separator-solid"></div>`

	html += `
        <div class="info-row">
            <span class="info-label">No Nota</span>
            <span class="info-value">` + data.TransactionID + `</span>
        </div>
        <div class="info-row">
            <span class="info-label">Tanggal</span>
            <span class="info-value">` + data.TransactionDate.Format("02/01/2006 15:04") + `</span>
        </div>
        <div class="info-row">
            <span class="info-label">Kasir</span>
            <span class="info-value">` + data.CashierName + `</span>
        </div>`

	if data.CustomerName != "" {
		html += `
        <div class="info-row">
            <span class="info-label">Pelanggan</span>
            <span class="info-value">` + data.CustomerName + `</span>
        </div>`
	}

	if data.CustomerPhone != "" {
		html += `
        <div class="info-row">
            <span class="info-label">No. Telp</span>
            <span class="info-value">` + data.CustomerPhone + `</span>
        </div>`
	}

	html += `<div class="separator"></div>`
	html += `<div class="items-section">`

	for _, item := range data.Items {
		html += `
        <div class="item">
            <div class="item-header">
                <span class="item-name">` + item.ProductName + `</span>
                <span class="item-price">Rp ` + formatCurrency(item.Subtotal) + `</span>
            </div>`

		if settings.ShowDescription && item.Description != "" {
			html += `
            <div class="item-details">` + item.Description + `</div>`
		}

		html += `
            <div class="item-details">
                ` + formatFloat(item.Quantity) + ` ` + item.Unit + ` x Rp ` + formatCurrency(item.Price) + `
            </div>
        </div>`
	}

	html += `</div>`

	html += `<div class="separator"></div>`

	html += `<div class="totals-section">`

	html += `
        <div class="total-row">
            <span class="total-label">Subtotal</span>
            <span class="total-value">Rp ` + formatCurrency(data.Subtotal) + `</span>
        </div>`

	if data.Tax > 0 {
		taxLabel := "Pajak"
		if data.TaxPercentage > 0 {
			taxLabel = fmt.Sprintf("Pajak (%.0f%%)", data.TaxPercentage)
		}
		html += `
        <div class="total-row">
            <span class="total-label">` + taxLabel + `</span>
            <span class="total-value">Rp ` + formatCurrency(data.Tax) + `</span>
        </div>`
	}

	if data.ServiceCharge > 0 {
		html += `
        <div class="total-row">
            <span class="total-label">Biaya Layanan</span>
            <span class="total-value">Rp ` + formatCurrency(data.ServiceCharge) + `</span>
        </div>`
	}

	if data.Discount > 0 {
		discLabel := "Diskon"
		if data.DiscountType == "percentage" {
			discLabel = fmt.Sprintf("Diskon (%.0f%%)", data.Discount)
		}
		html += `
        <div class="total-row">
            <span class="total-label">` + discLabel + `</span>
            <span class="total-value">- Rp ` + formatCurrency(data.Discount) + `</span>
        </div>`
	}

	html += `
        <div class="total-row grand-total">
            <span class="total-label">TOTAL</span>
            <span class="total-value">Rp ` + formatCurrency(data.Total) + `</span>
        </div>`

	html += `</div>`

	html += `
        <div class="payment-section">
            <div class="payment-row">
                <span class="total-label">Metode Pembayaran</span>
                <span class="total-value">` + formatPaymentMethod(data.PaymentMethod) + `</span>
            </div>
            <div class="payment-row">
                <span class="total-label">Bayar</span>
                <span class="total-value">Rp ` + formatCurrency(data.PaymentAmount) + `</span>
            </div>
            <div class="payment-row">
                <span class="total-label">Kembali</span>
                <span class="total-value">Rp ` + formatCurrency(data.Change) + `</span>
            </div>
        </div>`
	if settings.ShowQRCode && data.QRCodeData != "" {
		html += `
        <div class="qr-code">
            <img src="` + data.QRCodeData + `" alt="QR Code">
        </div>`
	} else if settings.ShowQRCode && format.QRCodeBase64 != "" {
		html += `
        <div class="qr-code">
            <img src="` + format.QRCodeBase64 + `" alt="QR Code">
        </div>`
	}

	// Footer
	html += `<div class="footer">`

	if settings.ShowFooterNote && settings.FooterNote != "" {
		html += `
        <div class="footer-note">` + settings.FooterNote + `</div>`
	}

	if settings.ShowWhatsappFooter && settings.WhatsappNote != "" {
		html += `
        <div class="footer-note">` + settings.WhatsappNote + `</div>`
	}

	if data.Notes != "" {
		html += `
        <div class="footer-note">Catatan: ` + data.Notes + `</div>`
	}

	html += `
        <div class="thank-you">Terima Kasih!</div>
        <div class="footer-note">Powered by Your POS System</div>
    </div>`

	html += `
    </div>
</body>
</html>`

	return html
}

// Helper function untuk format currency
func formatCurrency(amount float64) string {
	// Format dengan thousand separator
	formatted := fmt.Sprintf("%.2f", amount)
	parts := strings.Split(formatted, ".")

	// Add thousand separator
	intPart := parts[0]
	var result string
	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}

	return result + "," + parts[1]
}

// Helper function untuk format float
func formatFloat(num float64) string {
	s := fmt.Sprintf("%.2f", num)
	// Remove trailing zeros
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

// Helper function untuk format payment method
func formatPaymentMethod(method string) string {
	methods := map[string]string{
		"cash":     "Tunai",
		"card":     "Kartu",
		"qris":     "QRIS",
		"transfer": "Transfer",
		"ewallet":  "E-Wallet",
	}

	if translated, ok := methods[method]; ok {
		return translated
	}
	return strings.Title(method)
}

func centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}
	padding := (width - len(text)) / 2
	return repeatChar(" ", padding) + text
}

func repeatChar(char string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += char
	}
	return result
}
